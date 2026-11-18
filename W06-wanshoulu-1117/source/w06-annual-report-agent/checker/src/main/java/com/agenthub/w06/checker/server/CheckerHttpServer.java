/*
 * CheckerHttpServer — W06 核验服务的 HTTP 启动层（纯 JDK HttpServer，无 Spring）。
 *
 * 路由：
 *  GET  /health                          健康检查
 *  GET  /v1/banks                        银行清单（bank/company/report_period/pages/sources）
 *  GET  /v1/artifacts/template?bank=     正例模板工件（只读，对话流按银行取底稿）
 *  GET  /v1/sources                      信源注册表（多银行合并，含 bank 归属标注）
 *  GET  /v1/snapshot/manifest[?bank=]    快照 manifest 全文（缺省招行 cmb）
 *  GET  /v1/snapshot/pages/{n}[?bank=]   按页码读快照页文本（含该页 digest）
 *  POST /v1/checks/numeric               数值纪律检查（N1–N7）
 *  POST /v1/checks/citation              引用核验（含信源注册表与支持度）
 *  GET  /v1/tools                        MCP 风格联网工具网关：工具注册表（deploy/tools.json 配置驱动）
 *  POST /v1/tools/call                   统一工具调用 {tool_name, arguments}：
 *                                       web_search / web_fetch / pdf_extract / snapshot_commit（0.3.5）
 *  POST /v1/documents/import             本地文档导入（workbench「导入资料」）：{files|text, title?,
 *                                       authority?, url?} → 落盘 live 快照集 + sha256 注册进信源
 *                                       注册表（与 snapshot_commit 同纪律门，即刻可引用核验，无需重启）
 *  GET  /demo/{positive|error}-artifact.json  演示工件（受控样例）
 *  GET  /demo/reference-answer.json      W01 标准答案
 *  POST /v1/artifacts/digest             工件规范化 digest（Jackson 重序列化字节 → sha256）
 *  GET  /v1/approvals?task_id=           查询某任务的最新交付批准
 *  POST /v1/approvals?approver=...       申请交付批准：服务端重跑两项检查，
 *                                       任一非 PASSED → 409 REFUSED（错误草稿不能获得批准）
 *  GET  /v1/tasks                        报告库：历史任务列表（刷新不丢，服务端内存态）
 *  GET  /v1/tasks/{id}                   任务详情（task 记录 + 完整工件）
 *  POST /v1/tasks                        登记任务 {task,artifact}：服务端指派编号、
 *                                       同步工件 task_id、计算规范化 digest（upsert）
 *  GET  /v1/agents                       Agent Hub：{runtime, catalog}
 *  POST /v1/agents/install {agent_id}    市场目录 agent 安装注册 → 加入运行时
 *  POST /v1/agents/load {agent_id,task_id} 运行时 agent 按任务动态按需加载（记录最近加载）
 *  GET  /v1/skills                       技能注册表（skill→agent 绑定关系，registered_agents 实时反查）
 *  POST /v1/agents/register              自定义 agent 注册 {name,kind,description,capabilities[]} → 运行时
 *  POST /v1/agents/{id}/skills           skill 绑定/解绑给 agent {skill_id, action: bind|unbind}
 *  POST /v1/agents/editor-in-chief/team  agent 注册给主编团队/退出 {agent_id, action: register|unregister}
 *  DELETE /v1/agents/{id}                运行时注销（主编根节点除外；市场安装的退回市场目录）
 *  GET  /                                 workbench.html（工作台，五页签版）
 *
 * 环境变量：CHECKER_PORT（默认 18095）、SNAPSHOT_DIR（默认 /snapshot）、
 *          STATIC_DIR（默认 /app，含 workbench.html / testpage.html / demo/）
 */

package com.agenthub.w06.checker.server;

import com.agenthub.w06.checker.CitationChecker;
import com.agenthub.w06.checker.NumericChecker;
import com.agenthub.w06.contract.AnnualReportArtifacts;
import com.agenthub.w06.contract.DataTable;
import com.agenthub.w06.contract.CheckResult;
import com.agenthub.w06.contract.MetricValue;
import com.agenthub.w06.contract.ReportSection;
import com.agenthub.w06.contract.SourceEntry;
import com.agenthub.w06.contract.Violation;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.TreeMap;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;
import com.agenthub.w06.contract.Citation;

public final class CheckerHttpServer {

    static final ObjectMapper MAPPER = new ObjectMapper();
    static final NumericChecker NUMERIC = new NumericChecker();
    static final CitationChecker CITATION = new CitationChecker();
    static final Map<String, ObjectNode> APPROVALS = new HashMap<>(); // task_id -> 最新批准
    static final AtomicInteger APPROVAL_SEQ = new AtomicInteger();

    // 报告库（任务历史，内存态：容器存活期内刷新不丢）
    static final Map<String, ObjectNode> TASKS = new LinkedHashMap<>();               // task_id -> 任务记录
    static final Map<String, AnnualReportArtifacts> TASK_ARTIFACTS = new HashMap<>();  // task_id -> 交付工件
    static final AtomicInteger TASK_SEQ = new AtomicInteger(2);                       // 启动预置 001–004，preload 后新任务从 005 递增

    // Agent Hub：运行时（已注册，可动态加载）+ 市场目录（可安装注册）
    static final Map<String, ObjectNode> AGENT_RUNTIME = new LinkedHashMap<>();
    static final Map<String, ObjectNode> AGENT_CATALOG = new LinkedHashMap<>();

    // 技能注册表（skill→agent 绑定的合法对象）与自定义 agent 序号
    static final Map<String, ObjectNode> SKILL_REGISTRY = new LinkedHashMap<>();
    static final AtomicInteger CUSTOM_SEQ = new AtomicInteger();

    // ———————————————————— 快照解析器 ————————————————————

    /** 单银行快照：页文本（页码/摘要双向索引）+ 信源页清单 + manifest 元数据。 */
    static final class BankSnap {
        final String bank;
        final Path dir;
        final String company;
        final String reportPeriod;
        final String snapshotVersion;
        final JsonNode sourcesNode;
        final Map<String, String> byPageDigest = new HashMap<>();
        final Map<Integer, String> byPageNumber = new TreeMap<>();
        final Map<String, Set<Integer>> sourcePages = new HashMap<>();

        BankSnap(Path dir, String bank) throws IOException {
            this.dir = dir;
            this.bank = bank;
            JsonNode manifest = MAPPER.readTree(dir.resolve("manifest.json").toFile());
            this.company = manifest.path("company").asText(bank);
            this.reportPeriod = manifest.path("report_period").asText("");
            this.snapshotVersion = manifest.path("snapshot_version").asText("");
            this.sourcesNode = manifest.path("sources");
            Path pagesDir = dir.resolve("pages");
            if (Files.isDirectory(pagesDir)) {
                try (var stream = Files.list(pagesDir)) {
                    for (var f : stream.filter(p -> p.getFileName().toString()
                                    .matches("page-\\d+\\.txt"))
                            .sorted().toList()) {
                        String text = Files.readString(f, StandardCharsets.UTF_8);
                        int num = Integer.parseInt(f.getFileName().toString()
                                .replace("page-", "").replace(".txt", ""));
                        byPageNumber.put(num, text);
                        byPageDigest.put(sha256(text), text);
                    }
                }
            }
            for (JsonNode src : sourcesNode) {
                Set<Integer> pages = new HashSet<>();
                for (JsonNode p : src.path("pages")) {
                    pages.add(p.asInt());
                }
                sourcePages.put(src.path("content_sha256").asText(), pages);
            }
        }
    }

    /**
     * 多银行快照解析器：snapshotDir 下每个含 manifest.json 的子目录为一银行
     * （{cmb,icbc,ccb}）；旧单银行布局（manifest.json/pages 直接在根）视为 cmb。
     *
     * 取文语义（fail-closed）：
     *  · digest 非空 —— 页文本摘要全局唯一，跨银行精确命中才返回；信源 content_sha256
     *    （多页合并文本摘要）在工件所属银行的 sourcePages 内按页码定位；
     *  · digest 给定却未命中任何处 → 返回 null，绝不按页码凭空回退（更严防编造）；
     *  · digest 为空 —— 只在工件所属银行内按页码取文（防 N6 逐页扫描跨银行串行）。
     */
    static final class SnapshotHttpResolver implements NumericChecker.SnapshotTextResolver {
        final LinkedHashMap<String, BankSnap> banks = new LinkedHashMap<>();
        final String defaultBank;
        final Map<String, String> sourceBank = new HashMap<>(); // content_sha256 → 所属银行

        SnapshotHttpResolver(Path snapshotDir) throws IOException {
            try (var stream = Files.list(snapshotDir)) {
                for (Path d : stream.filter(Files::isDirectory).sorted().toList()) {
                    if (Files.isRegularFile(d.resolve("manifest.json"))) {
                        String bank = d.getFileName().toString();
                        banks.put(bank, new BankSnap(d, bank));
                    }
                }
            }
            if (banks.isEmpty()) { // 兼容旧单银行布局
                banks.put("cmb", new BankSnap(snapshotDir, "cmb"));
            }
            this.defaultBank = banks.containsKey("cmb") ? "cmb"
                    : banks.keySet().iterator().next();
            for (BankSnap b : banks.values()) {
                for (String sha : b.sourcePages.keySet()) {
                    sourceBank.put(sha, b.bank);
                }
            }
        }

        /** 工件公司名 → 银行目录名；未识别返回 null（scoped 到默认银行）。 */
        static String bankOf(String company) {
            if (company == null) {
                return null;
            }
            if (company.contains("招商")) {
                return "cmb";
            }
            if (company.contains("工商")) {
                return "icbc";
            }
            if (company.contains("建设")) {
                return "ccb";
            }
            return null;
        }

        private BankSnap scope(String bank) {
            return bank != null && banks.containsKey(bank)
                    ? banks.get(bank) : banks.get(defaultBank);
        }

        /** 核验用 scoped 解析视图：按工件所属银行取文，checker 接口不变。 */
        NumericChecker.SnapshotTextResolver scopedFor(AnnualReportArtifacts draft) {
            String b = bankOf(draft == null ? null : draft.company());
            final String scopeBank = b != null && banks.containsKey(b) ? b : defaultBank;
            return (digest, page) -> textOfBank(scopeBank, digest, page);
        }

        String textOfBank(String bank, String digest, int page) {
            if (digest != null) {
                for (BankSnap b : banks.values()) { // 页文本全局唯一，digest 命中即真实原文
                    String hit = b.byPageDigest.get(digest);
                    if (hit != null) {
                        return hit;
                    }
                }
                BankSnap scoped = scope(bank);
                Set<Integer> pages = scoped.sourcePages.get(digest); // 信源摘要（多页合并）
                if (pages != null && pages.contains(page)) {
                    return scoped.byPageNumber.get(page);
                }
                // 信源级 sha 同样全局唯一：scoped 未命中时跨快照库兜底（live 联网快照可被任意工件引用）
                for (BankSnap b : banks.values()) {
                    if (b == scoped) {
                        continue;
                    }
                    Set<Integer> crossPages = b.sourcePages.get(digest);
                    if (crossPages != null && crossPages.contains(page)) {
                        return b.byPageNumber.get(page);
                    }
                }
                return null; // digest 给定却未命中 → null，绝不按页码回退
            }
            return scope(bank).byPageNumber.get(page);
        }

        @Override
        public String textOf(String digest, int page) {
            return textOfBank(null, digest, page); // 未指定银行 → 默认银行（cmb）
        }
    }

    /** 信源注册表：合并各银行 manifest 的 sources（SourceEntry 不携带 bank 字段，银行归属由 resolver.sourceBank 提供）。 */
    static List<SourceEntry> loadSourceEntries(SnapshotHttpResolver resolver) throws IOException {
        java.util.List<SourceEntry> entries = new java.util.ArrayList<>();
        for (BankSnap b : resolver.banks.values()) {
            for (JsonNode src : b.sourcesNode) {
                ObjectNode node = MAPPER.createObjectNode();
                node.put("source_id", src.path("source_id").asText());
                node.put("title", src.path("title").asText());
                node.put("authority", src.path("authority").asText());
                node.put("origin_url", src.path("origin_url").asText());
                node.put("fetched_at", src.path("fetched_at").asText());
                node.put("content_sha256", src.path("content_sha256").asText());
                ArrayNode pages = node.putArray("pages");
                for (JsonNode p : src.path("pages")) {
                    pages.add(p.asInt());
                }
                entries.add(MAPPER.treeToValue(node, SourceEntry.class));
            }
        }
        return List.copyOf(entries);
    }

    // ———————————————————— main 与路由 ————————————————————

    public static void main(String[] args) throws IOException {
        int port = Integer.parseInt(System.getenv().getOrDefault("CHECKER_PORT", "18095"));
        Path snapshotDir = Path.of(System.getenv().getOrDefault("SNAPSHOT_DIR", "/snapshot"));
        Path staticDir = Path.of(System.getenv().getOrDefault("STATIC_DIR", "/app"));
        loadReportStructure(staticDir);               // 骨架配置（节序/标题/静态正文/责任 agent 唯一事实源）
        ToolGateway.load(staticDir);                  // MCP 工具网关注册表（deploy/tools.json 配置驱动）

        SnapshotHttpResolver resolver = new SnapshotHttpResolver(snapshotDir);
        // CopyOnWriteArrayList：snapshot_commit 运行时追加信源（并发核验读安全，无需重启）
        List<SourceEntry> sources = new java.util.concurrent.CopyOnWriteArrayList<>(loadSourceEntries(resolver));
        preloadTasks(staticDir, snapshotDir, resolver, sources);   // 报告库预置 4 条历史任务（三银行）
        seedAgents();                                 // Agent Hub：11 运行时 + 3 市场目录
        seedSkills();                                 // 技能注册表：8 个技能（可绑定给 agent）

        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.setExecutor(Executors.newFixedThreadPool(8));

        server.createContext("/", ex -> {
            try {
                route(ex, snapshotDir, staticDir, resolver, sources);
            } catch (Exception e) {
                respond(ex, 500, errorNode("internal_error", String.valueOf(e)));
            } finally {
                ex.close();
            }
        });

        server.start();
        System.out.println("W06 checker server up on :" + port
                + "  snapshot=" + snapshotDir + "  banks=" + resolver.banks.keySet()
                + "  sources=" + sources.size()
                + "  numeric=" + NumericChecker.CHECKER_VERSION
                + "  tools=" + ToolGateway.REGISTRY.size());
    }

    private static void route(HttpExchange ex, Path snapshotDir, Path staticDir,
                              SnapshotHttpResolver resolver, List<SourceEntry> sources) throws IOException {
        String path = ex.getRequestURI().getPath();
        String method = ex.getRequestMethod();

        if ("/health".equals(path)) {
            ObjectNode ok = MAPPER.createObjectNode();
            ok.put("status", "ok");
            ok.put("service", "w06-checker");
            ok.put("numeric_checker_version", NumericChecker.CHECKER_VERSION);
            ok.put("citation_checker_version", CitationChecker.CHECKER_VERSION);
            respond(ex, 200, ok);
            return;
        }
        if ("/v1/sources".equals(path)) {
            ArrayNode arr = MAPPER.createArrayNode();
            for (SourceEntry s : sources) {
                ObjectNode n = MAPPER.createObjectNode();
                String bk = resolver.sourceBank.get(s.contentSha256());
                n.put("bank", bk == null ? "" : bk);
                n.put("source_id", s.sourceId());
                n.put("title", s.title());
                n.put("authority", s.authority() == null ? null : s.authority().name());
                n.put("content_sha256", s.contentSha256());
                ArrayNode pages = n.putArray("pages");
                if (s.pages() != null) {
                    s.pages().forEach(pages::add);
                }
                arr.add(n);
            }
            respond(ex, 200, arr);
            return;
        }
        if ("/v1/banks".equals(path)) {
            ArrayNode arr = MAPPER.createArrayNode();
            for (BankSnap b : resolver.banks.values()) {
                ObjectNode n = MAPPER.createObjectNode();
                n.put("bank", b.bank);
                n.put("company", b.company);
                n.put("report_period", b.reportPeriod);
                ArrayNode pages = n.putArray("pages");
                b.byPageNumber.keySet().forEach(pages::add);
                ArrayNode srcs = n.putArray("sources");
                for (JsonNode src : b.sourcesNode) {
                    ObjectNode s = srcs.addObject();
                    s.put("source_id", src.path("source_id").asText());
                    s.put("title", src.path("title").asText());
                    s.put("pages", src.path("pages").size());
                }
                arr.add(n);
            }
            respond(ex, 200, arr);
            return;
        }
        if ("/v1/artifacts/template".equals(path)) { // 正例模板工件（只读）：对话流按银行取底稿
            Map<String, String> query = parseQuery(ex.getRequestURI().getRawQuery());
            String bank = query.getOrDefault("bank", resolver.defaultBank);
            BankSnap b = resolver.banks.get(bank);
            Path f = b == null ? null : b.dir.resolve("positive-artifact.json");
            if (f == null || !Files.isRegularFile(f)) {
                respond(ex, 404, errorNode("template_not_found",
                        "无模板工件: bank=" + bank + "（可用: " + resolver.banks.keySet() + "）"));
                return;
            }
            ex.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
            respondRaw(ex, 200, Files.readAllBytes(f));
            return;
        }
        if ("/v1/snapshot/manifest".equals(path)) {
            Map<String, String> query = parseQuery(ex.getRequestURI().getRawQuery());
            String bank = query.getOrDefault("bank", resolver.defaultBank);
            BankSnap b = resolver.banks.get(bank);
            if (b == null) {
                respond(ex, 404, errorNode("bank_not_found", "无银行快照: " + bank
                        + "（可用: " + resolver.banks.keySet() + "）"));
                return;
            }
            ex.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
            respondRaw(ex, 200, Files.readAllBytes(b.dir.resolve("manifest.json")));
            return;
        }
        if (path.startsWith("/v1/snapshot/pages/")) {
            int page = Integer.parseInt(path.substring("/v1/snapshot/pages/".length()));
            Map<String, String> query = parseQuery(ex.getRequestURI().getRawQuery());
            String bank = query.get("bank");
            String text = resolver.textOfBank(
                    bank != null && resolver.banks.containsKey(bank) ? bank : null, null, page);
            if (text == null) {
                respond(ex, 404, errorNode("page_not_found", "快照无第 " + page + " 页"));
                return;
            }
            ObjectNode n = MAPPER.createObjectNode();
            n.put("page", page);
            n.put("digest", sha256(text));
            n.put("text", text);
            respond(ex, 200, n);
            return;
        }
        if ("/v1/checks/numeric".equals(path) && "POST".equals(method)) {
            handleCheck(ex, resolver, draft -> NUMERIC.check(draft, resolver.scopedFor(draft)));
            return;
        }
        if ("/v1/checks/citation".equals(path) && "POST".equals(method)) {
            handleCheck(ex, resolver, draft -> CITATION.check(draft, resolver.scopedFor(draft), sources));
            return;
        }
        if ("/v1/pipeline/run".equals(path) && "POST".equals(method)) {
            handlePipelineRun(ex, resolver, sources);
            return;
        }
        if ("/v1/tools".equals(path) && "GET".equals(method)) {
            respond(ex, 200, ToolGateway.list());
            return;
        }
        if ("/v1/tools/call".equals(path) && "POST".equals(method)) {
            JsonNode toolBody = MAPPER.readTree(ex.getRequestBody());
            respond(ex, 200, ToolGateway.call(toolBody, snapshotDir, resolver, sources));
            return;
        }
        if ("/v1/documents/import".equals(path) && "POST".equals(method)) {
            handleDocumentImport(ex, snapshotDir, resolver, sources);
            return;
        }
        if (path.startsWith("/demo/") && "GET".equals(method)) {
            String name = path.substring("/demo/".length());
            if (!List.of("positive-artifact.json", "error-artifact.json",
                    "reference-answer.json", "metric-spec.json").contains(name)) {
                respond(ex, 404, errorNode("not_found", "未知演示工件: " + name));
                return;
            }
            Path f = staticDir.resolve("demo").resolve(name);
            if (!Files.isRegularFile(f)) {
                respond(ex, 404, errorNode("not_found", "演示工件缺失: " + name));
                return;
            }
            ex.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
            respondRaw(ex, 200, Files.readAllBytes(f));
            return;
        }
        if ("/v1/artifacts/digest".equals(path) && "POST".equals(method)) {
            AnnualReportArtifacts draft = MAPPER.readValue(ex.getRequestBody(), AnnualReportArtifacts.class);
            String digest = "sha256:" + sha256(MAPPER.writeValueAsBytes(draft));
            ObjectNode n = MAPPER.createObjectNode();
            n.put("task_id", draft.taskId());
            n.put("artifact_digest", digest);
            respond(ex, 200, n);
            return;
        }
        if ("/v1/approvals".equals(path)) {
            Map<String, String> query = parseQuery(ex.getRequestURI().getRawQuery());
            if ("GET".equals(method)) {
                ObjectNode approval = APPROVALS.get(query.getOrDefault("task_id", ""));
                respond(ex, 200, approval == null ? emptyApproval(query.get("task_id")) : approval);
                return;
            }
            if ("POST".equals(method)) {
                handleApproval(ex, resolver, sources, query.get("approver"));
                return;
            }
        }
        if ("/v1/tasks".equals(path)) {
            if ("GET".equals(method)) {
                ArrayNode arr = MAPPER.createArrayNode();
                for (ObjectNode t : TASKS.values()) arr.add(t.deepCopy());
                respond(ex, 200, arr);
                return;
            }
            if ("POST".equals(method)) {
                handleTaskRegister(ex);
                return;
            }
        }
        if (path.startsWith("/v1/tasks/") && "GET".equals(method)) {
            String id = path.substring("/v1/tasks/".length());
            ObjectNode task = TASKS.get(id);
            AnnualReportArtifacts artifact = TASK_ARTIFACTS.get(id);
            if (task == null || artifact == null) {
                respond(ex, 404, errorNode("task_not_found", "报告库无任务 " + id));
                return;
            }
            ObjectNode out = MAPPER.createObjectNode();
            out.set("task", task.deepCopy());
            out.set("artifact", MAPPER.valueToTree(artifact));
            respond(ex, 200, out);
            return;
        }
        if ("/v1/skills".equals(path) && "GET".equals(method)) {
            handleSkillsList(ex);
            return;
        }
        if ("/v1/agents".equals(path) && "GET".equals(method)) {
            ObjectNode out = MAPPER.createObjectNode();
            ArrayNode runtime = out.putArray("runtime");
            for (ObjectNode a : AGENT_RUNTIME.values()) runtime.add(a.deepCopy());
            ArrayNode catalog = out.putArray("catalog");
            for (ObjectNode a : AGENT_CATALOG.values()) catalog.add(a.deepCopy());
            respond(ex, 200, out);
            return;
        }
        if ("/v1/agents/install".equals(path) && "POST".equals(method)) {
            handleAgentInstall(ex);
            return;
        }
        if ("/v1/agents/load".equals(path) && "POST".equals(method)) {
            handleAgentLoad(ex);
            return;
        }
        if ("/v1/agents/register".equals(path) && "POST".equals(method)) {
            handleAgentRegister(ex);
            return;
        }
        if (path.matches("/v1/agents/[^/]+/skills") && "POST".equals(method)) {
            handleSkillBind(ex, path);
            return;
        }
        if ("/v1/agents/editor-in-chief/team".equals(path) && "POST".equals(method)) {
            handleTeamRegister(ex);
            return;
        }
        if (path.matches("/v1/agents/[^/]+") && "DELETE".equals(method)) {
            handleAgentUnregister(ex, path.substring("/v1/agents/".length()));
            return;
        }
        if ("/".equals(path) || "/workbench.html".equals(path) || "/testpage.html".equals(path)) {
            String file = "/".equals(path) ? "workbench.html" : path.substring(1);
            Path f = staticDir.resolve(file);
            if (!Files.isRegularFile(f)) {
                respond(ex, 404, errorNode("not_found", file + " 未部署"));
                return;
            }
            ex.getResponseHeaders().set("Content-Type", "text/html; charset=utf-8");
            respondRaw(ex, 200, Files.readAllBytes(f));
            return;
        }
        respond(ex, 404, errorNode("not_found", "未知路由: " + method + " " + path));
    }

    // ———————————————————— 检查与批准 ————————————————————

    private static void handleCheck(HttpExchange ex, SnapshotHttpResolver resolver,
                                    java.util.function.Function<AnnualReportArtifacts, CheckResult> checker)
            throws IOException {
        byte[] body = ex.getRequestBody().readAllBytes();
        AnnualReportArtifacts draft;
        try {
            draft = MAPPER.readValue(body, AnnualReportArtifacts.class);
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "工件反序列化失败: " + e.getMessage()));
            return;
        }
        try {
            respond(ex, 200, MAPPER.valueToTree(checker.apply(draft)));
        } catch (Exception e) {
            respond(ex, 500, errorNode("check_failed", String.valueOf(e)));
        }
    }

    /**
     * 交付批准（受控交付链）：
     * 服务端对提交工件重跑 numeric + citation 两项检查，任一非 PASSED → 409 REFUSED；
     * 全部 PASSED → 记录批准（绑定工件规范化 digest，内容一改 digest 即变、旧批准失效）。
     */
    private static void handleApproval(HttpExchange ex, SnapshotHttpResolver resolver,
                                       List<SourceEntry> sources, String approver) throws IOException {
        if (approver == null || approver.isBlank()) {
            respond(ex, 400, errorNode("bad_request", "缺少 approver 参数（批准须具名）"));
            return;
        }
        AnnualReportArtifacts draft;
        try {
            draft = MAPPER.readValue(ex.getRequestBody(), AnnualReportArtifacts.class);
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "工件反序列化失败: " + e.getMessage()));
            return;
        }
        // 完整性前置：审批 body 必须是完整工件（metrics/sections 缺失→400，不进核验器空指针）
        if (draft.metrics() == null || draft.sections() == null) {
            respond(ex, 400, errorNode("bad_request",
                    "审批 body 须为完整工件（AnnualReportArtifacts），缺少 metrics/sections 字段"));
            return;
        }
        CheckResult numeric;
        CheckResult citation;
        try {
            numeric = NUMERIC.check(draft, resolver.scopedFor(draft));
            citation = CITATION.check(draft, resolver.scopedFor(draft), sources);
        } catch (Exception e) {
            respond(ex, 500, errorNode("check_failed", String.valueOf(e)));
            return;
        }
        boolean allPassed = "PASSED".equals(String.valueOf(numeric.status()))
                && "PASSED".equals(String.valueOf(citation.status()));
        String digest = "sha256:" + sha256(MAPPER.writeValueAsBytes(draft));
        if (!allPassed) {
            ObjectNode refused = MAPPER.createObjectNode();
            refused.put("status", "REFUSED");
            refused.put("message", "错误草稿不能获得交付批准（服务端重跑检查未全部 PASSED）");
            refused.put("artifact_digest", digest);
            refused.set("numeric", MAPPER.valueToTree(numeric));
            refused.set("citation", MAPPER.valueToTree(citation));
            respond(ex, 409, refused);
            return;
        }
        ObjectNode approval = MAPPER.createObjectNode();
        approval.put("approval_id", "APR-" + String.format("%04d", APPROVAL_SEQ.incrementAndGet()));
        approval.put("task_id", draft.taskId());
        approval.put("approver", approver);
        approval.put("approved_at", Instant.now().toString());
        approval.put("artifact_digest", digest);
        ObjectNode checks = approval.putObject("checks");
        ObjectNode nc = checks.putObject("numeric");
        nc.put("checker_id", numeric.checkerId());
        nc.put("checker_version", numeric.checkerVersion());
        nc.put("status", String.valueOf(numeric.status()));
        ObjectNode cc = checks.putObject("citation");
        cc.put("checker_id", citation.checkerId());
        cc.put("checker_version", citation.checkerVersion());
        cc.put("status", String.valueOf(citation.status()));
        approval.put("status", "APPROVED");
        APPROVALS.put(draft.taskId(), approval);
        // 批准同步进报告库任务记录（若该任务已登记）：状态 APPROVED、digest 更新、工件以批准版本为准
        ObjectNode task = TASKS.get(draft.taskId());
        if (task != null) {
            task.put("status", "APPROVED");
            task.put("artifact_digest", digest);
            task.put("updated_at", Instant.now().toString());
            task.set("checks", checksNode(numeric, citation));
            task.set("approval", approval.deepCopy());
            TASK_ARTIFACTS.put(draft.taskId(), draft);
        }
        respond(ex, 200, approval);
    }

    private static ObjectNode emptyApproval(String taskId) {
        ObjectNode n = MAPPER.createObjectNode();
        n.put("task_id", taskId);
        n.put("status", "NONE");
        n.put("message", "该任务尚无交付批准记录");
        return n;
    }

    /**
     * 本地文档导入（workbench「导入资料」/对话粘贴通道）：与 MCP snapshot_commit 同一纪律门。
     * 文件解析（pdf→PDFBox 逐页抽取；文本→段落分页）后经 ToolGateway.commitSource 落盘
     * /snapshot/live 并注册进信源注册表 → 引用核验 C1/C2 与数值核验 N3 即刻接受，
     * 冻结三银行快照零改动，无需重启容器。
     * 请求：{title?, authority?, url?, files:[{filename, content_base64}], text?}
     *      （files 与 text 至少其一；pdf/txt/md/json，单文件 PDF ≤20 页，文本按 2500 字符分页）
     */
    private static void handleDocumentImport(HttpExchange ex, Path snapshotDir,
                                             SnapshotHttpResolver resolver,
                                             List<SourceEntry> sources) throws IOException {
        JsonNode body = MAPPER.readTree(ex.getRequestBody());
        String title = body.path("title").asText("").trim();
        String authority = body.path("authority").asText("PRIMARY_OFFICIAL").trim().toUpperCase();
        try {
            com.agenthub.w06.contract.SourceAuthority.valueOf(authority);
        } catch (IllegalArgumentException e) {
            respond(ex, 400, errorNode("bad_request", "authority 无效: " + authority
                    + "（可用: PRIMARY_OFFICIAL / SECONDARY_INSTITUTIONAL / TERTIARY_AGGREGATED）"));
            return;
        }
        String url = body.path("url").asText("").trim();
        JsonNode filesIn = body.path("files");
        List<Map.Entry<String, byte[]>> files = new ArrayList<>();
        if (filesIn.isArray()) {
            for (JsonNode fIn : filesIn) {
                String fn = fIn.path("filename").asText("document.txt");
                String b64 = fIn.path("content_base64").asText("");
                if (b64.isBlank()) {
                    continue;
                }
                byte[] bytes;
                try {
                    bytes = java.util.Base64.getMimeDecoder().decode(b64);
                } catch (IllegalArgumentException e) {
                    bytes = new byte[0]; // 交给 parseUploadFile 报「内容为空」
                }
                files.add(Map.entry(fn, bytes));
            }
        }
        String pasted = body.path("text").asText("");
        if (!pasted.isBlank()) {
            files.add(Map.entry("对话粘贴补充资料.txt", pasted.getBytes(StandardCharsets.UTF_8)));
        }
        if (files.isEmpty()) {
            respond(ex, 400, errorNode("bad_request",
                    "缺少导入内容：files:[{filename, content_base64}] 或 text 必填其一"));
            return;
        }
        ArrayNode imported = MAPPER.createArrayNode();
        ArrayNode errors = MAPPER.createArrayNode();
        for (Map.Entry<String, byte[]> f : files) {
            String filename = f.getKey();
            ObjectNode parsed = ToolGateway.parseUploadFile(filename, f.getValue());
            if (!parsed.path("ok").asBoolean()) {
                ObjectNode e = errors.addObject();
                e.put("filename", filename);
                e.put("error", parsed.path("error").asText("解析失败"));
                continue;
            }
            String base = filename.substring(filename.lastIndexOf('/') + 1)
                    .substring(filename.lastIndexOf('\\') + 1);
            String srcTitle = title.isEmpty() ? base
                    : (files.size() == 1 ? title : title + "（" + base + "）");
            List<String> bodies = new ArrayList<>();
            List<Integer> originPages = new ArrayList<>();
            for (JsonNode pg : parsed.path("pages")) {
                bodies.add(pg.path("text").asText());
                originPages.add(pg.path("origin_page").asInt());
            }
            ObjectNode committed = ToolGateway.commitSource(ToolGateway.LIVE_BANK,
                    url.isEmpty() ? "upload:local/" + base : url,
                    srcTitle, authority, bodies, originPages,
                    "用户本地导入（workbench 资料导入通道）：与联网抓取同经 sha256 注册纪律门",
                    snapshotDir, resolver, sources);
            ObjectNode item = imported.addObject();
            item.put("filename", filename);
            if (parsed.hasNonNull("total_pages")) {
                item.put("parsed_pages", parsed.path("total_pages").asInt());
            }
            item.set("commit", committed);
        }
        if (imported.isEmpty()) {
            ObjectNode e = errorNode("import_failed", "没有文件导入成功");
            e.set("errors", errors);
            respond(ex, 400, e);
            return;
        }
        ObjectNode out = MAPPER.createObjectNode();
        out.put("ok", true);
        out.put("bank", ToolGateway.LIVE_BANK);
        out.put("snapshot_dir", "/snapshot/" + ToolGateway.LIVE_BANK);
        out.set("imported", imported);
        out.set("errors", errors);
        out.put("note", "导入完成：资料已落盘 live 快照集并注册进信源注册表，引用核验 C1/C2 即刻接受；"
                + "生成报告时可直接引用这些资料（已挂载为 [C-n] 可引用论断）。");
        respond(ex, 200, out);
    }

    /**
     * 用户导入资料挂载（live 快照集 → 工件引用）：把已注册导入资料的每一页登记为可引用论断
     * （digest/page/excerpt 三元组由服务端按落盘文本生成，excerpt 逐字取自页文本 → L1 确定性
     * 核验天然通过；claim_text 与 excerpt 一致 → L2 支持度判定为 SUPPORTED），并挂到首个
     * 分析节的 claim_ids 使其进入写作任务书的引用清单与快照页原文——writer-agent 可直接
     * 以 [C-n] 引用用户导入的资料。无导入资料时工件原样返回。
     */
    private static AnnualReportArtifacts attachImportedClaims(AnnualReportArtifacts art,
                                                              SnapshotHttpResolver resolver,
                                                              List<String> analysisIds) {
        BankSnap live = resolver.banks.get(ToolGateway.LIVE_BANK);
        if (live == null || live.sourcesNode.isEmpty() || analysisIds.isEmpty()) {
            return art;
        }
        int next = 1;
        for (Citation c : art.citations()) {
            var m = java.util.regex.Pattern.compile("C-(\\d+)").matcher(c.claimId());
            if (m.matches()) {
                next = Math.max(next, Integer.parseInt(m.group(1)) + 1);
            }
        }
        List<Citation> added = new ArrayList<>();
        outer:
        for (JsonNode src : live.sourcesNode) {
            for (JsonNode pno : src.path("pages")) {
                int page = pno.asInt();
                String text = live.byPageNumber.get(page);
                if (text == null || text.isBlank()) {
                    continue;
                }
                String excerpt = pickExcerpt(text);
                Citation c = new Citation("C-" + next, excerpt,
                        sha256(text), page, excerpt, sha256(excerpt));
                if (!"SUPPORTED".equals(CitationChecker.assessSupport(c).name())) {
                    continue; // 摘录缺数字/实体支撑的页不挂载，避免语义核验灰区
                }
                next++;
                added.add(c);
                if (added.size() >= 12) {
                    break outer; // 单次任务导入引用上限，防任务书上下文膨胀
                }
            }
            if (added.size() >= 12) {
                break;
            }
        }
        if (added.isEmpty()) {
            return art;
        }
        List<Citation> citations = new ArrayList<>(art.citations());
        citations.addAll(added);
        List<String> newIds = added.stream().map(Citation::claimId).toList();
        List<ReportSection> sections = new ArrayList<>();
        boolean attached = false;
        for (ReportSection s : art.sections()) {
            List<String> ids = s.claimIds();
            if (!attached && analysisIds.contains(s.sectionId())) {
                ids = new ArrayList<>(ids);
                ids.addAll(newIds);
                attached = true;
            }
            sections.add(new ReportSection(s.sectionId(), s.title(), s.body(), ids));
        }
        return new AnnualReportArtifacts(art.taskId(), art.company(), art.reportPeriod(),
                art.generatedAt(), art.metrics(), citations, sections, art.charts());
    }

    /** 引用摘录选取：优先取含数字的 160 字窗口（保证 L2 确定性支持度判定通过），否则页首 120 字。 */
    private static String pickExcerpt(String text) {
        String t = text.strip();
        if (t.length() <= 160) {
            return t;
        }
        var m = java.util.regex.Pattern.compile("\\d").matcher(t);
        if (m.find()) {
            int idx = m.start();
            int start = Math.max(0, Math.min(idx - 20, t.length() - 160));
            int end = Math.min(t.length(), start + 160);
            return t.substring(start, end).strip();
        }
        return t.substring(0, 120).strip();
    }

    // ———————————————————— 报告库（任务历史）与 Agent Hub ————————————————————

    /** 骨架配置（0.3.4 泛化）：节序/role/责任 agent/静态正文模板的唯一事实源 = STATIC_DIR/report-structure.json。 */
    private static JsonNode STRUCTURE;

    private static void loadReportStructure(Path staticDir) {
        Path f = staticDir.resolve("report-structure.json");
        if (!Files.isRegularFile(f)) {
            System.out.println("WARN: 未找到 report-structure.json，/v1/pipeline/run 不可用");
            return;
        }
        try {
            STRUCTURE = MAPPER.readTree(f.toFile());
        } catch (IOException e) {
            System.out.println("WARN: report-structure.json 解析失败: " + e);
        }
    }

    /** 工件规模摘要（动态计数，杜绝硬编码「N 项指标/N 节正文」）。 */
    private static String artifactBrief(AnnualReportArtifacts art) {
        int charts = art.charts() == null ? 0 : art.charts().size();
        return art.metrics().size() + " 项指标 + " + art.sections().size() + " 节正文 + "
                + art.citations().size() + " 条引用" + (charts > 0 ? " + " + charts + " 幅图表" : "");
    }

    /**
     * 六工位受控生成流水线（确定性重放，无 LLM）：查数 → 分析 → 图表 → 写作 → 核验 → 主编。
     * 分派/标题/节序全部来自骨架配置与工件数据，服务端零机构特判；换银行=换数据配置。
     */
    private static void handlePipelineRun(HttpExchange ex, SnapshotHttpResolver resolver,
                                          List<SourceEntry> sources) throws IOException {
        if (STRUCTURE == null) {
            respond(ex, 503, errorNode("structure_missing", "骨架配置 report-structure.json 未加载"));
            return;
        }
        JsonNode body = MAPPER.readTree(ex.getRequestBody());
        String bank = body.path("bank").asText(resolver.defaultBank);
        String mode = body.path("mode").asText("brief");
        String taskId = body.path("task_id").asText("");
        BankSnap snap = resolver.scope(bank);
        Path f = snap.dir.resolve("positive-artifact.json");
        if (!Files.isRegularFile(f)) {
            respond(ex, 404, errorNode("artifact_missing", "银行快照缺少正例工件: " + snap.bank));
            return;
        }
        AnnualReportArtifacts art = MAPPER.treeToValue(MAPPER.readTree(f.toFile()), AnnualReportArtifacts.class);

        // 节 → 责任 agent / role 映射自骨架配置（不写死「第几节归谁」）
        Map<String, String> sectionAgent = new LinkedHashMap<>();
        Map<String, String> sectionRole = new LinkedHashMap<>();
        var secIt = STRUCTURE.path("sections").fields();
        while (secIt.hasNext()) {
            var e = secIt.next();
            sectionAgent.put(e.getKey(), e.getValue().path("agent").asText("editor-in-chief"));
            sectionRole.put(e.getKey(), e.getValue().path("role").asText("analysis"));
        }
        // 分析工位 agent 集合按配置 role=analysis 去重收集（顺序=配置节序）
        Set<String> analysts = new LinkedHashSet<>();
        for (Map.Entry<String, String> e : sectionAgent.entrySet()) {
            if ("analysis".equals(sectionRole.get(e.getKey()))) {
                analysts.add(e.getValue());
            }
        }

        // 用户导入资料（live 快照集）→ 挂载为可引用论断（digest/page/excerpt 服务端生成，
        // fail-closed 纪律不变）：brief/long_form 均可引用；无导入资料时工件原样返回
        List<String> analysisIds = new ArrayList<>();
        for (ReportSection sec : art.sections()) {
            if ("analysis".equals(sectionRole.getOrDefault(sec.sectionId(), "analysis"))) {
                analysisIds.add(sec.sectionId());
            }
        }
        art = attachImportedClaims(art, resolver, analysisIds);

        // 数据能力补强：从快照页原文抽取结构化财报表格 + 确定性汇总分析。
        // 结果作为 tables 注入任务书（writer 用真实数据支撑论断）并随 artifact 输出供渲染。
        ArrayNode dataTables = extractDataTables(resolver, bank);

        // long_form 模式（0.3.7 架构分工）：Java 只做确定性编排——组装任务书（纯数据）、
        // 解析 [S-n] 成稿、调度核验与反思回路；任务分解与成稿由 openjiuwen deep agent
        // （writer-agent）依据 annual-report 技能方法论自主完成；框架按 sessionId 维护
        // 会话记忆，续写与反思修订轮在同一会话内进行。
        LongFormOutcome lf = null;
        if ("long_form".equals(mode)) {
            lf = runLongForm(art, snap, sectionRole, sectionAgent, resolver, sources, dataTables);
            art = lf.artifacts();
        }

        ArrayNode stages = MAPPER.createArrayNode();

        // 1) 查数：快照装载 + 指标索引
        ObjectNode s1 = stageNode("material", "查数", "material-agent", "装载冻结快照与指标索引");
        ObjectNode d1 = s1.putObject("data");
        d1.put("snapshot_version", snap.snapshotVersion);
        d1.put("sources", snap.sourcesNode.size());
        d1.put("pages", snap.byPageNumber.size());
        d1.put("metrics_indexed", art.metrics().size());
        ArrayNode det1 = s1.withArray("details");
        for (JsonNode src : snap.sourcesNode) {
            det1.add(src.path("title").asText("（未命名信源）") + " — " + src.path("pages").size() + " 页");
        }
        BankSnap liveSnap = resolver.banks.get(ToolGateway.LIVE_BANK);
        if (liveSnap != null && liveSnap.sourcesNode.size() > 0) {
            det1.add("用户导入资料 " + liveSnap.sourcesNode.size() + " 份（live 快照集已注册，"
                    + "已挂载为可引用论断）: " + liveSnap.sourcesNode.path(0).path("title").asText("") + " 等");
        }
        det1.add("指标索引 " + art.metrics().size() + " 项，口径与单位以指标字典为准");
        if (!ToolGateway.REGISTRY.isEmpty()) {
            d1.put("tools_available", ToolGateway.REGISTRY.size());
            det1.add("联网工具 " + ToolGateway.REGISTRY.size() + " 个（POST /v1/tools/call）: "
                    + String.join(" / ", ToolGateway.REGISTRY.keySet())
                    + " —— 外部素材须 snapshot_commit 落盘注册后方可引用");
        }
        stages.add(s1);

        // 2) 分析：逐指标洞察（值·单位·同比口径），口径文案由 DeltaUnit 承载
        ObjectNode s2 = stageNode("analysis", "分析", String.join(", ", analysts), "逐指标洞察（值 · 单位 · 同比口径）");
        ArrayNode det2 = s2.withArray("details");
        for (MetricValue m : art.metrics()) {
            StringBuilder ins = new StringBuilder();
            ins.append(m.key().getDisplayName()).append(" ").append(m.value()).append(" ").append(m.unit());
            if (m.yoyDelta() != null && !m.yoyDelta().isEmpty()) {
                boolean sign = m.yoyDelta().startsWith("-") || m.yoyDelta().startsWith("+");
                ins.append("，同比 ").append(sign ? "" : "+").append(m.yoyDelta())
                        .append(m.deltaUnit() == null ? "" : m.deltaUnit().getLabel());
            }
            det2.add(ins.toString());
        }
        stages.add(s2);

        // 3) 图表：类型/标题/挂载节随图表数据文件，chart-agent 只渲染不改数
        ObjectNode s3 = stageNode("chart", "图表", "chart-agent", "内联 SVG 图表绘制与挂载校验");
        ArrayNode det3 = s3.withArray("details");
        for (Map<String, Object> ch : art.charts()) {
            det3.add("[" + ch.get("id") + " " + ch.get("type") + "] " + ch.get("title")
                    + " → 挂载 " + ch.get("section_id"));
        }
        stages.add(s3);

        // 4) 写作：成稿执行统一归 writer-agent（openjiuwen deep agent）；骨架 sectionAgent
        //    表示各节分析洞察的责任 agent（Stage 2 分析工位），两者职责互补不冲突。
        String s4Headline = "long_form".equals(mode)
                ? "任务书分派 → writer-agent 自主分解成稿 → 核验违规反思修订（openjiuwen）"
                : "按骨架配置逐节成稿（名称保留 / 深度收缩）";
        ObjectNode s4 = stageNode("writing", "写作", "writer-agent", s4Headline);
        s4.put("mode", mode);
        ArrayNode det4 = s4.withArray("details");
        if (lf != null) {
            for (String r : lf.rounds()) {
                det4.add(r);
            }
        }
        for (ReportSection sec : art.sections()) {
            String role = sectionRole.getOrDefault(sec.sectionId(), "analysis");
            String source = "long_form".equals(mode) && "analysis".equals(role)
                    ? "writer-agent 生成" : "模板填充";
            det4.add(sec.sectionId() + " " + sec.title()
                    + " · 分析责任 " + sectionAgent.getOrDefault(sec.sectionId(), "editor-in-chief")
                    + " · " + source
                    + " · 正文 " + sec.body().length() + " 字 · 引用 " + sec.claimIds().size() + " 条");
        }
        stages.add(s4);

        // 5) 核验：真实跑两项确定性核验（N1-N7 / C1-C2）
        CheckResult num = NUMERIC.check(art, resolver.scopedFor(art));
        CheckResult cit = CITATION.check(art, resolver.scopedFor(art), sources);
        ObjectNode s5 = stageNode("checking", "核验", "numeric-checker, citation-checker", "N1-N7 数值核验 + C1-C2 引用核验（确定性，无 LLM）");
        s5.withArray("details").add("数值核验 " + num.status()
                + (num.violations().isEmpty() ? "" : "，违规 " + num.violations().size() + " 条"));
        s5.withArray("details").add("引用核验 " + cit.status()
                + (cit.violations().isEmpty() ? "" : "，违规 " + cit.violations().size() + " 条"));
        stages.add(s5);

        // 交付登记：产物 upsert 进报告库——对话流传入预登记 task_id 则沿用该编号，
        // 直跑（curl/脚本）则自动指派；报告生成后即从报告库（GET /v1/tasks）可见。
        boolean passed = "PASSED".equals(String.valueOf(num.status()))
                && "PASSED".equals(String.valueOf(cit.status()));
        ObjectNode regMeta = MAPPER.createObjectNode();
        regMeta.put("status", passed ? "GENERATED" : "FAILED");
        regMeta.put("summary", ("long_form".equals(mode) ? "长文流水线（writer-agent 成稿）：" : "简报流水线：")
                + art.company() + " " + art.reportPeriod()
                + "，两项核验 " + (passed ? "PASSED，待具名批准" : "FAILED，已拦截"));
        ObjectNode taskRec = registerTask(taskId, regMeta, art);
        String finalTaskId = taskRec.path("task_id").asText();
        String digest = taskRec.path("artifact_digest").asText();

        // 6) 主编：汇总 + 批准指引（批准绑定工件内容指纹）
        ObjectNode s6 = stageNode("consolidation", "主编", "editor-in-chief", "汇总核验结论并出具批准指引");
        s6.withArray("details").add("工件内容指纹 " + digest + "（规模：" + artifactBrief(art) + "）");
        s6.withArray("details").add("批准指引：POST /v1/approvals?task_id=" + finalTaskId
                + "（approver 具名；批准绑定内容指纹，工件一改批准即失效）");
        stages.add(s6);

        ObjectNode out = MAPPER.createObjectNode();
        out.put("task_id", finalTaskId);
        out.put("bank", snap.bank);
        out.put("company", art.company());
        out.put("report_period", art.reportPeriod());
        out.put("artifact_digest", digest);
        ObjectNode artifactOut = MAPPER.valueToTree(TASK_ARTIFACTS.get(finalTaskId));
        if (dataTables != null && !dataTables.isEmpty()) artifactOut.set("tables", dataTables.deepCopy());
        out.set("artifact", artifactOut);
        out.set("checks", checksNode(num, cit));
        out.set("stages", stages);
        respond(ex, 200, out);
    }

    // ------------------------------------------------------------------
    // long_form 写作（0.3.7 架构分工）：
    //   skill  承载方法与流程（写作方法/口径句式/引用纪律/修订流程，见 annual-report 技能）；
    //   agent  负责任务分解、成稿与反思优化（openjiuwen deep agent：task planning + 会话记忆）；
    //   Java   只做确定性编排：任务书组装（纯运行时数据）、[S-n] 解析、核验与反思回路调度。
    // Java 侧不含任何写作方法论——「怎么写」由 agent 侧技能决定。
    // ------------------------------------------------------------------

    /** agent 输出的章节块首行格式：整行仅 [S-n]（容许前置 # 标记与括号内空白，接受后统一规范化）。 */
    private static final java.util.regex.Pattern SECTION_HEAD =
            java.util.regex.Pattern.compile("^\\s*(?:#+\\s*)?\\[\\s*(S-\\d{1,2})\\s*\\]\\s*$");

    /** 采纳成稿的最小正文字数（低于此值视为异常响应，不采纳）。 */
    private static final int MIN_BODY_CHARS = 50;

    /** long_form 生成结果：最终工件 + agent 协作轮次记录（供 Stage 4 展示）。 */
    private record LongFormOutcome(AnnualReportArtifacts artifacts, List<String> rounds) {}

    /**
     * 数据能力补强：从快照页原文抽取结构化财报表格并做确定性汇总分析。
     * 只消费已落盘快照，无联网、无臆造。表格必须与正文 metrics 同源——按 bank 取对应银行快照集：
     * bank=live 用联网抓取/导入的 live 集；其余银行用各自的冻结节选集，杜绝跨银行数据混用。
     * 返回 ArrayNode（每表含 table_id/title/unit/columns/rows/source_page/analysis）。
     */
    private static ArrayNode extractDataTables(SnapshotHttpResolver resolver, String bank) {
        ArrayNode out = MAPPER.createArrayNode();
        try {
            // 与正文同源：正文用 resolver.scope(bank)，表格也必须用它；live 集仅在 bank=live 时取用。
            BankSnap snap = ToolGateway.LIVE_BANK.equals(bank)
                    ? resolver.banks.get(ToolGateway.LIVE_BANK)
                    : resolver.scope(bank);
            if (snap == null || snap.byPageNumber.isEmpty()) return out;
            Map<Integer, String> pageDigest = new HashMap<>();
            for (var e : snap.sourcePages.entrySet()) {
                for (int pg : e.getValue()) pageDigest.put(pg, e.getKey());
            }
            for (Map.Entry<Integer, String> e : snap.byPageNumber.entrySet()) {
                int pg = e.getKey();
                String text = e.getValue();
                long numLines = java.util.regex.Pattern.compile("\\d[\\d,]*\\.\\d{2}")
                        .matcher(text).results().count();
                // 放宽阈值：整数型表格（如五级分类/行业结构）也纳入数据支撑
                if (numLines < 2 && text.split("\\R").length < 20) continue;
                String digest = pageDigest.getOrDefault(pg, "sha256:page-" + pg);
                List<DataTable> tables = TableExtractor.extract(text, pg, digest, List.of());
                for (DataTable t : tables) {
                    ObjectNode tn = MAPPER.valueToTree(t);
                    String nice = canonicalTableTitle(pg, t.tableId(), t.title(), t.rows().isEmpty() ? "" : t.rows().get(0).item());
                    tn.put("title", nice);
                    tn.set("analysis", DataAnalysis.analyze(t));
                    out.add(tn);
                }
            }
        } catch (Exception e) {
            System.err.println("extractDataTables: " + e.getMessage());
        }
        return out;
    }

    /** 财报表格规范标题（按 live 页号/表序号，取自年报原文栏目名；兜底用算法标题或首行项目）。 */
    private static String canonicalTableTitle(int pg, String tableId, String algoTitle, String firstItem) {
        String key = pg + ":" + tableId;
        Map<String, String> M = new HashMap<>();
        M.put("6:T-6-1", "主要财务数据（利息净收入/手续费及佣金/营业收入等，三年对比）");
        M.put("8:T-8-1", "主要财务指标（盈利能力/资产质量/资本充足率，2025-2023）");
        M.put("9:T-9-1", "分季度经营数据（营业收入/归母净利润）");
        M.put("18:T-18-1", "利润表主要项目变动");
        M.put("19:T-19-1", "生息资产平均收益率（资产端）");
        M.put("19:T-19-3", "计息负债平均付息率（负债端）");
        M.put("23:T-23-1", "手续费及佣金净收入构成");
        M.put("25:T-25-1", "营业收入地区分布");
        M.put("26:T-26-1", "资产构成及占比");
        M.put("27:T-27-1", "贷款结构（公司类/个人类，金额与占比）");
        M.put("28:T-28-1", "投资结构（债券/权益工具）");
        M.put("31:T-31-1", "负债结构（客户存款/同业存放等）");
        M.put("32:T-32-2", "存款地区分布");
        M.put("36:T-36-1", "分部营业收入及占比（公司/个人/资管/金融市场）");
        M.put("48:T-48-1", "托管资产类别分布");
        M.put("61:T-61-1", "机构分布（总行及各地区）");
        M.put("76:T-76-1", "贷款五级分类分布情况");
        M.put("76:T-76-2", "贷款和不良贷款结构（按业务类型）");
        M.put("77:T-77-1", "按行业划分的公司类贷款和不良贷款结构");
        M.put("78:T-78-1", "按地域划分的贷款和不良贷款结构");
        M.put("78:T-78-2", "贷款减值准备变动情况");
        M.put("79:T-79-1", "按担保方式划分的贷款结构");
        M.put("86:T-86-1", "流动性比例");
        M.put("92:T-92-1", "资本充足率情况（核心一级/一级/资本充足率）");
        M.put("93:T-93-1", "风险加权资产构成");
        String hit = M.get(key);
        if (hit != null) return hit;
        if (algoTitle != null && !algoTitle.isBlank()
                && !algoTitle.matches(".*\\d[\\d,]*\\.\\d.*")
                && algoTitle.length() <= 24) return algoTitle;
        return firstItem == null || firstItem.isBlank() ? "财报数据表（第" + pg + "页）" : firstItem + "等项目数据表（第" + pg + "页）";
    }

    /**
     * long_form 主流程：任务书分派 → 缺节续写 → 核验违规反思修订。
     * 所有轮次复用同一 openjiuwen 会话（框架按 sessionId 维护上下文与记忆）。
     */
    private static LongFormOutcome runLongForm(AnnualReportArtifacts art, BankSnap snap,
            Map<String, String> sectionRole, Map<String, String> sectionAgent,
            SnapshotHttpResolver resolver, List<SourceEntry> sources, ArrayNode dataTables) {
        List<String> rounds = new ArrayList<>();
        List<String> analysisIds = new ArrayList<>();
        for (ReportSection sec : art.sections()) {
            if ("analysis".equals(sectionRole.getOrDefault(sec.sectionId(), "analysis"))) {
                analysisIds.add(sec.sectionId());
            }
        }
        if (analysisIds.isEmpty()) {
            return new LongFormOutcome(art, rounds);
        }

        // 会话仅限单次生成流程内复用（任务书→续写→反思共享上下文）；跨 run 必须隔离，
        // 否则框架记忆残留上次产出，agent 会复述旧稿而不按新任务书重新成稿。
        String sessionId = "w06-lf-" + art.taskId() + "-" + System.currentTimeMillis();

        // 1) 任务书分派：agent 自主分解章节写作任务并成稿（框架 task planning + 会话记忆）
        String draft = AgentChatClient.chat(
                buildTaskBook(art, snap, resolver.banks.get(ToolGateway.LIVE_BANK),
                        sectionRole, sectionAgent, analysisIds, dataTables), sessionId);
        if (draft == null) {
            rounds.add("agent 不可达，本次回退模板正文（可稍后重试）");
            return new LongFormOutcome(art, rounds);
        }
        Map<String, String> bodies = parseSectionBodies(draft, analysisIds);
        rounds.add("写作轮：任务书分派（会话 " + sessionId + "），agent 产出 "
                + bodies.size() + "/" + analysisIds.size() + " 节");

        // 2) 缺节续写：输出截断或漏节时由同一会话续写（框架记忆支撑），最多 4 轮
        for (int round = 1; round <= 4; round++) {
            List<String> missing = new ArrayList<>();
            for (String id : analysisIds) {
                if (!bodies.containsKey(id)) missing.add(id);
            }
            if (missing.isEmpty()) break;
            String cont = AgentChatClient.chat(buildContinueDirective(missing), sessionId);
            if (cont == null) break;
            bodies.putAll(parseSectionBodies(cont, missing));
            rounds.add("续写轮 " + round + "：补写 " + String.join("、", missing)
                    + "，累计 " + bodies.size() + "/" + analysisIds.size() + " 节");
        }

        // 3) 应用成稿并核验
        AnnualReportArtifacts best = applyBodies(art, bodies);
        CheckResult num = NUMERIC.check(best, resolver.scopedFor(best));
        CheckResult cit = CITATION.check(best, resolver.scopedFor(best), sources);
        int violations = num.violations().size() + cit.violations().size();

        // 4) 反思修订回路：违规清单回灌同一会话，agent 按技能修订流程自纠，最多 2 轮；不劣化原则
        if (violations > 0) {
            rounds.add("核验：数值 " + num.status() + " / 引用 " + cit.status()
                    + "，违规 " + violations + " 条，进入反思修订");
        }
        for (int round = 1; round <= 2 && violations > 0; round++) {
            String revised = AgentChatClient.chat(buildRevisionFeedback(num, cit), sessionId);
            if (revised == null) break;
            Map<String, String> patches = parseSectionBodies(revised, analysisIds);
            if (patches.isEmpty()) break;
            AnnualReportArtifacts cand = applyBodies(best, patches);
            CheckResult n2 = NUMERIC.check(cand, resolver.scopedFor(cand));
            CheckResult c2 = CITATION.check(cand, resolver.scopedFor(cand), sources);
            int v2 = n2.violations().size() + c2.violations().size();
            if (v2 < violations) {
                rounds.add("反思轮 " + round + "：修订 " + patches.size() + " 节，违规 "
                        + violations + " → " + v2);
                best = cand;
                num = n2;
                cit = c2;
                violations = v2;
            } else {
                rounds.add("反思轮 " + round + "：复核违规 " + v2 + " 条未改善，保留上一版");
                break;
            }
        }
        return new LongFormOutcome(best, rounds);
    }

    /**
     * 构造长文写作任务书（Java → writer-agent 的分派契约，纯运行时数据，不含写作方法论）。
     * 方法与流程由 agent 侧 annual-report 技能承载（任务书首句指令 agent 先经 skill_tool
     * 加载技能与 references/long-form-writing.md）；任务书只注入骨架派生的节清单、
     * 指标表、引用清单、快照页原文，以及 Java 解析所需的输出协议。
     */
    private static String buildTaskBook(AnnualReportArtifacts art, BankSnap snap, BankSnap liveSnap,
            Map<String, String> sectionRole, Map<String, String> sectionAgent, List<String> analysisIds,
            ArrayNode dataTables) {
        StringBuilder sb = new StringBuilder();
        sb.append("【长文写作任务书】\n");
        sb.append("请先用 skill_tool 加载 annual-report 技能及 references/long-form-writing.md，");
        sb.append("按技能方法自主分解本任务书的章节写作计划，并逐节成稿。\n\n");
        sb.append("任务目标：为以下银行撰写年报完整章节正文（正式年报口径，非简报）。\n");
        sb.append("写作深度要求：每个分析章节正文不少于 1500 字，按「经营背景→数据陈述（引用原文）→结构拆解→同比与口径分析→风险与管理层应对→小结」多维展开；");
        sb.append("充分使用快照页原文的事实、数字、管理层论述，分段落撰写，不得写成要点罗列或三两句之简报。\n");
        sb.append("银行：").append(art.company()).append("\n");
        sb.append("报告期：").append(art.reportPeriod()).append("\n\n");

        // 章节清单（骨架派生，只含 analysis 节；分析洞察责任 agent 用于写作侧重参考）
        sb.append("章节清单（每节对应一个输出块）：\n");
        for (ReportSection sec : art.sections()) {
            if (!analysisIds.contains(sec.sectionId())) continue;
            sb.append("- ").append(sec.sectionId()).append(" ").append(sec.title())
                    .append("（分析洞察责任：")
                    .append(sectionAgent.getOrDefault(sec.sectionId(), "editor-in-chief"))
                    .append("；关联引用：");
            sb.append(sec.claimIds().isEmpty() ? "无" : String.join("、", sec.claimIds()));
            sb.append("）\n");
        }
        sb.append("\n");

        // 指标数据表
        sb.append("指标数据表（值与单位为准，不得修改）：\n");
        for (MetricValue m : art.metrics()) {
            sb.append("- ").append(m.key().getDisplayName())
                    .append("：").append(m.value()).append(m.unit());
            if (m.yoyDelta() != null && !m.yoyDelta().isEmpty()) {
                boolean sign = m.yoyDelta().startsWith("-") || m.yoyDelta().startsWith("+");
                sb.append("，同比").append(sign ? "" : "+").append(m.yoyDelta());
                if (m.deltaUnit() != null) sb.append(m.deltaUnit().getLabel());
            }
            sb.append("\n");
        }
        sb.append("\n");

        // 引用清单（全部分析节的 claims）
        Set<String> wantedClaims = new LinkedHashSet<>();
        for (ReportSection sec : art.sections()) {
            if (analysisIds.contains(sec.sectionId())) wantedClaims.addAll(sec.claimIds());
        }
        sb.append("引用清单：\n");
        for (var c : art.citations()) {
            if (!wantedClaims.contains(c.claimId())) continue;
            String excerpt = c.evidenceExcerpt();
            if (excerpt.length() > 200) excerpt = excerpt.substring(0, 200) + "…";
            sb.append("- [").append(c.claimId()).append("] 第").append(c.page()).append("页：")
                    .append(excerpt).append("\n");
        }
        sb.append("\n");

        // 快照页原文
        Set<Integer> citedPages = new TreeSet<>();
        for (var c : art.citations()) {
            if (wantedClaims.contains(c.claimId())) citedPages.add(c.page());
        }
        sb.append("快照页原文（数据来源，仅供引用；含用户导入资料的 live 快照页）：\n");
        for (int pageNum : citedPages) {
            String pageText = snap.byPageNumber.get(pageNum);
            if (pageText == null && liveSnap != null) { // 导入资料页在 live 快照集内
                pageText = liveSnap.byPageNumber.get(pageNum);
            }
            if (pageText != null) {
                String trimmed = pageText.length() > 3000 ? pageText.substring(0, 3000) + "…" : pageText;
                sb.append("--- 第").append(pageNum).append("页 ---\n").append(trimmed).append("\n");
            }
        }
        sb.append("\n");

        // 结构化数据表（数据能力补强）：table_extract + table_analyze 产出的真实财报表格。
        // 每节写作必须优先引用这些表格中的具体数字（含单位/年份/来源页），用数据支撑论断。
        if (dataTables != null && !dataTables.isEmpty()) {
            sb.append("结构化数据表（数据来源：table_extract/table_analyze 工具从快照页原文抽取，"
                    + "逐格为原文数值；务必在正文引用其中的具体数字，标注来源页，不要只写定性文字）：\n");
            int shown = 0;
            for (JsonNode t : dataTables) {
                // 优先注入规范命名的关键表（跳过算法兜底标题的低价值页），控制任务书长度
                String ttl0 = t.path("title").asText("");
                boolean keyTable = !(ttl0.contains("第") && ttl0.contains("页）")) && !ttl0.isEmpty();
                if (!keyTable) continue;
                if (shown++ >= 22) { sb.append("（其余表略，完整 50 表随工件输出）\n"); break; }
                String title = t.path("title").asText("");
                String unit = t.path("unit").asText("");
                int pg = t.path("source_page").asInt();
                JsonNode cols = t.path("columns");
                sb.append("- [").append(t.path("table_id").asText(""))
                        .append("] 第").append(pg).append("页 ")
                        .append(title.isEmpty() ? "（财报数据表）" : title);
                if (!unit.isEmpty()) sb.append("（").append(unit).append("）");
                sb.append(" 列：").append(cols.isArray() ? cols.toString() : "[]").append("\n");
                JsonNode rows = t.path("rows");
                int rshown = 0;
                for (JsonNode r : rows) {
                    if (rshown++ >= 14) { sb.append("    …（余行略）\n"); break; }
                    sb.append("    ").append(r.path("item").asText(""))
                            .append("：").append(r.path("values").toString()).append("\n");
                }
                JsonNode an = t.path("analysis");
                if (!an.isNull() && an.has("yoy") && an.path("yoy").isArray()
                        && !an.path("yoy").isEmpty()) {
                    sb.append("    汇总分析示例：");
                    int y = 0;
                    for (JsonNode yo : an.path("yoy")) {
                        if (y++ >= 3) break;
                        sb.append(yo.path("item").asText(""))
                                .append(" 同比").append(yo.path("calc_growth_pct").asText(""))
                                .append("%（").append(yo.path("formula").asText("")).append("）；");
                    }
                    sb.append("\n");
                }
            }
            sb.append("\n");
        }

        // 输出协议（Java 解析契约；写作口径/句式等验收标准见技能）
        sb.append("输出协议（服务端按此解析，须严格遵守）：\n");
        sb.append("1. 只输出章节清单所列各节的正文，不输出任务书解读、开场白或收尾总结；\n");
        sb.append("2. 每节一个块，块首行只写 [S-n]（n 为章节编号，如 [S-1]），正文从下一行开始；\n");
        sb.append("3. 数据陈述标注对应引用编号 [C-n]；派生值标注「基于 [C-k] 计算」并给出算式；\n");
        sb.append("4. 不编造快照页原文之外的数据。\n");
        return sb.toString();
    }

    /** 缺节续写指令（输出协议不变；同一会话内 agent 可见此前任务书与成稿）。 */
    private static String buildContinueDirective(List<String> missing) {
        return "上一轮输出不完整，缺少章节：" + String.join("、", missing) + "\n"
                + "请按任务书继续完成缺失章节的写作，只输出缺失章节的块，"
                + "输出协议不变（块首行只写 [S-n]，正文从下一行开始）。";
    }

    /** 核验违规反馈（反思修订轮输入）：违规清单 + 修订契约。 */
    private static String buildRevisionFeedback(CheckResult num, CheckResult cit) {
        StringBuilder sb = new StringBuilder();
        sb.append("核验未通过，违规清单如下。请按 annual-report 技能的修订流程反思并修订正文。\n\n");
        sb.append("## 数值核验（").append(num.status()).append("）\n");
        for (Violation v : num.violations()) {
            sb.append("- [").append(v.rule()).append("] ");
            if (v.target() != null && !v.target().isEmpty()) sb.append(v.target()).append("：");
            sb.append(v.detail()).append("\n");
        }
        sb.append("\n## 引用核验（").append(cit.status()).append("）\n");
        for (Violation v : cit.violations()) {
            sb.append("- [").append(v.rule()).append("] ");
            if (v.target() != null && !v.target().isEmpty()) sb.append(v.target()).append("：");
            sb.append(v.detail()).append("\n");
        }
        sb.append("\n修订要求：只修订与违规相关的章节；未修订的章节不要输出。")
                .append("输出协议：仅输出被修订章节的块，块首行只写 [S-n]，正文从下一行开始，")
                .append("数据陈述标注 [C-n]。");
        return sb.toString();
    }

    /**
     * 解析 agent 输出的 [S-n] 块。块首行整行仅 [S-n]，块体为正文；
     * 只收章节清单内的节，正文不足 MIN_BODY_CHARS 视为异常响应不采纳。
     */
    private static Map<String, String> parseSectionBodies(String text, List<String> wantedIds) {
        Map<String, String> out = new LinkedHashMap<>();
        if (text == null || text.isBlank()) return out;
        String currentId = null;
        StringBuilder body = new StringBuilder();
        for (String line : text.split("\\R", -1)) {
            var m = SECTION_HEAD.matcher(line);
            if (m.matches()) {
                acceptSection(out, currentId, body, wantedIds);
                currentId = m.group(1);
                body.setLength(0);
            } else if (currentId != null) {
                body.append(line).append('\n');
            }
        }
        acceptSection(out, currentId, body, wantedIds);
        return out;
    }

    /** 输出协议 token 解码容差：[ S-7 ]/[ C-22 ] 统一规范化为 [S-7]/[C-22]（仅去括号内空白）。 */
    private static final java.util.regex.Pattern BRACKET_TOKEN =
            java.util.regex.Pattern.compile("\\[\\s*(S-\\d{1,2}|C-\\d{1,3})\\s*\\]");

    private static void acceptSection(Map<String, String> out, String id,
            StringBuilder body, List<String> wantedIds) {
        if (id == null || !wantedIds.contains(id)) return;
        String b = BRACKET_TOKEN.matcher(body.toString().trim()).replaceAll("[$1]");
        if (b.length() >= MIN_BODY_CHARS) out.put(id, b);
    }

    /** 将 agent 成稿应用到工件：仅替换出现的节，其余节保留原正文。 */
    private static AnnualReportArtifacts applyBodies(AnnualReportArtifacts art, Map<String, String> bodies) {
        if (bodies.isEmpty()) return art;
        List<ReportSection> out = new ArrayList<>();
        for (ReportSection sec : art.sections()) {
            String b = bodies.get(sec.sectionId());
            out.add(b != null
                    ? new ReportSection(sec.sectionId(), sec.title(), b, sec.claimIds())
                    : sec);
        }
        return new AnnualReportArtifacts(art.taskId(), art.company(), art.reportPeriod(),
                art.generatedAt(), art.metrics(), art.citations(), out, art.charts());
    }

    private static ObjectNode stageNode(String id, String name, String agent, String headline) {
        ObjectNode n = MAPPER.createObjectNode();
        n.put("stage_id", id);
        n.put("name", name);
        n.put("agent", agent);
        n.put("headline", headline);
        n.putArray("details");
        n.putObject("data");
        return n;
    }

    /**
     * 预置 4 条历史任务（报告库初始记录，覆盖三银行）：
     *  TASK-2025-AR-001 招行正例 —— 启动时真实重跑两项核验（均 PASSED），绑定批准 APR-0001（Alice）；
     *  TASK-2025-AR-002 招行受控错误样例 —— 真实重跑数值核验（N7 拦截 FAILED），状态 REFUSED 未交付；
     *  TASK-2025-AR-003 工行正例（Bob / APR-0002）、TASK-2025-AR-004 建行正例（Carol / APR-0003）。
     */
    private static void preloadTasks(Path staticDir, Path snapshotDir, SnapshotHttpResolver resolver,
                                     List<SourceEntry> sources) throws IOException {
        // —— AR-001 招行正例 / AR-002 招行错误样例（受控演示件）——
        Path posF = staticDir.resolve("demo").resolve("positive-artifact.json");
        Path errF = staticDir.resolve("demo").resolve("error-artifact.json");
        if (!Files.isRegularFile(posF) || !Files.isRegularFile(errF)) {
            System.out.println("WARN: demo 工件缺失，报告库未预置招行历史任务");
        } else {
            ObjectNode posArt = (ObjectNode) MAPPER.readTree(posF.toFile());
            AnnualReportArtifacts pos = MAPPER.treeToValue(posArt, AnnualReportArtifacts.class);
            CheckResult posNum = NUMERIC.check(pos, resolver.scopedFor(pos));
            CheckResult posCit = CITATION.check(pos, resolver.scopedFor(pos), sources);
            String posDigest = "sha256:" + sha256(MAPPER.writeValueAsBytes(pos));
            ObjectNode t1 = approvedTask("TASK-2025-AR-001", pos, "APPROVED",
                    "2025 年度经营简报：" + artifactBrief(pos) + "，两项核验 PASSED，Alice 具名批准交付",
                    posDigest, "2026-03-28T09:30:00Z", "2026-03-28T09:41:00Z", posNum, posCit);
            ObjectNode apr = MAPPER.createObjectNode();
            apr.put("approval_id", "APR-0001");
            apr.put("task_id", "TASK-2025-AR-001");
            apr.put("approver", "Alice");
            apr.put("approved_at", "2026-03-28T09:41:00Z");
            apr.put("artifact_digest", posDigest);
            apr.put("status", "APPROVED");
            apr.set("checks", checksNode(posNum, posCit));
            t1.set("approval", apr);
            TASKS.put("TASK-2025-AR-001", t1);
            TASK_ARTIFACTS.put("TASK-2025-AR-001", pos);
            APPROVALS.put("TASK-2025-AR-001", apr);

            ObjectNode errArt = (ObjectNode) MAPPER.readTree(errF.toFile());
            errArt.put("task_id", "TASK-2025-AR-002");
            AnnualReportArtifacts err = MAPPER.treeToValue(errArt, AnnualReportArtifacts.class);
            CheckResult errNum = NUMERIC.check(err, resolver.scopedFor(err));
            CheckResult errCit = CITATION.check(err, resolver.scopedFor(err), sources);
            ObjectNode t2 = MAPPER.createObjectNode();
            t2.put("task_id", "TASK-2025-AR-002");
            t2.put("company", err.company());
            t2.put("report_period", err.reportPeriod());
            t2.put("status", "REFUSED");
            t2.put("summary", "受控错误样例：拨备变动误写成「下降20.19%」，被 N7 变动口径规则拦截，未获交付");
            t2.put("artifact_digest", "sha256:" + sha256(MAPPER.writeValueAsBytes(err)));
            t2.put("created_at", "2026-03-28T10:05:00Z");
            t2.put("updated_at", "2026-03-28T10:07:00Z");
            t2.set("checks", checksNode(errNum, errCit));
            TASKS.put("TASK-2025-AR-002", t2);
            TASK_ARTIFACTS.put("TASK-2025-AR-002", err);
        }
        // —— AR-003 工行正例 / AR-004 建行正例（快照内多银行工件）——
        loadBankTask(snapshotDir, resolver, sources, "icbc", "TASK-2025-AR-003",
                "Bob", "APR-0002", "2026-03-28T11:10:00Z", "2026-03-28T11:22:00Z");
        loadBankTask(snapshotDir, resolver, sources, "ccb", "TASK-2025-AR-004",
                "Carol", "APR-0003", "2026-03-28T12:10:00Z", "2026-03-28T12:24:00Z");
        APPROVAL_SEQ.set(3);
        TASK_SEQ.set(4);
        System.out.println("报告库预置 4 条历史任务（AR-001/003/004 已批准 / AR-002 已拦截）");
    }

    /** 从快照目录读某银行正例工件，真实重跑两项核验后登记为历史任务（PASSED 才记批准）。 */
    private static void loadBankTask(Path snapshotDir, SnapshotHttpResolver resolver,
                                     List<SourceEntry> sources, String bank, String taskId,
                                     String approver, String approvalId,
                                     String createdAt, String updatedAt) throws IOException {
        Path f = snapshotDir.resolve(bank).resolve("positive-artifact.json");
        if (!Files.isRegularFile(f)) {
            System.out.println("WARN: 缺少 " + bank + " 正例工件，跳过 " + taskId);
            return;
        }
        AnnualReportArtifacts art = MAPPER.treeToValue(MAPPER.readTree(f.toFile()), AnnualReportArtifacts.class);
        CheckResult num = NUMERIC.check(art, resolver.scopedFor(art));
        CheckResult cit = CITATION.check(art, resolver.scopedFor(art), sources);
        String digest = "sha256:" + sha256(MAPPER.writeValueAsBytes(art));
        boolean passed = "PASSED".equals(String.valueOf(num.status()))
                && "PASSED".equals(String.valueOf(cit.status()));
        String summary = art.company() + " " + art.reportPeriod() + " 年度经营简报：" + artifactBrief(art) + "，"
                + "两项核验 " + (passed ? "PASSED" : "FAILED") + "，" + approver + " 具名批准交付";
        ObjectNode t = approvedTask(taskId, art, passed ? "APPROVED" : "FAILED", summary,
                digest, createdAt, updatedAt, num, cit);
        ObjectNode apr = MAPPER.createObjectNode();
        apr.put("approval_id", approvalId);
        apr.put("task_id", taskId);
        apr.put("approver", approver);
        apr.put("approved_at", updatedAt);
        apr.put("artifact_digest", digest);
        apr.put("status", passed ? "APPROVED" : "FAILED");
        apr.set("checks", checksNode(num, cit));
        t.set("approval", apr);
        TASKS.put(taskId, t);
        TASK_ARTIFACTS.put(taskId, art);
        if (passed) {
            APPROVALS.put(taskId, apr);
        }
    }

    private static ObjectNode approvedTask(String taskId, AnnualReportArtifacts art, String status,
                                           String summary, String digest, String createdAt, String updatedAt,
                                           CheckResult num, CheckResult cit) {
        ObjectNode t = MAPPER.createObjectNode();
        t.put("task_id", taskId);
        t.put("company", art.company());
        t.put("report_period", art.reportPeriod());
        t.put("status", status);
        t.put("summary", summary);
        t.put("artifact_digest", digest);
        t.put("created_at", createdAt);
        t.put("updated_at", updatedAt);
        t.set("checks", checksNode(num, cit));
        return t;
    }

    /** 报告库登记：请求体 {task:{…}, artifact:{…}}，服务端指派/沿用编号后 upsert（POST /v1/tasks）。 */
    private static void handleTaskRegister(HttpExchange ex) throws IOException {
        JsonNode body;
        try {
            body = MAPPER.readTree(ex.getRequestBody());
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败: " + e.getMessage()));
            return;
        }
        if (!body.path("task").isObject() || !body.path("artifact").isObject()) {
            respond(ex, 400, errorNode("bad_request", "请求体须为 {task:{…}, artifact:{…}}"));
            return;
        }
        AnnualReportArtifacts artifact;
        try {
            artifact = MAPPER.treeToValue(body.path("artifact").deepCopy(), AnnualReportArtifacts.class);
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "工件反序列化失败: " + e.getMessage()));
            return;
        }
        JsonNode metaIn = body.path("task");
        ObjectNode task = registerTask(metaIn.path("task_id").asText(""), metaIn, artifact);
        respond(ex, 200, task.deepCopy());
    }

    /**
     * 报告库 upsert 登记（对话流登记与流水线产物落库共用同一入口）：沿用传入编号，
     * 未传则自动指派；工件内 task_id 与登记编号同步；计算规范化 digest 写入任务记录。
     *
     * @param id       期望任务编号（空则自动指派 TASK-2025-AR-%03d）
     * @param metaIn   任务元数据（status/summary；null 时按 GENERATED 登记）
     * @param artifact 交付工件
     * @return 登记后的任务记录（TASKS 内引用）
     */
    private static ObjectNode registerTask(String id, JsonNode metaIn, AnnualReportArtifacts artifact)
            throws IOException {
        if (id == null || id.isBlank()) {
            id = String.format("TASK-2025-AR-%03d", TASK_SEQ.incrementAndGet());
        }
        ObjectNode artNode = (ObjectNode) MAPPER.valueToTree(artifact);
        artNode.put("task_id", id);
        AnnualReportArtifacts synced = MAPPER.treeToValue(artNode, AnnualReportArtifacts.class);
        String digest = "sha256:" + sha256(MAPPER.writeValueAsBytes(synced));
        ObjectNode task = TASKS.get(id);
        if (task == null) {
            task = MAPPER.createObjectNode();
            task.put("task_id", id);
            task.put("created_at", Instant.now().toString());
        }
        task.put("company", synced.company());
        task.put("report_period", synced.reportPeriod());
        task.put("status", metaIn == null ? "GENERATED" : metaIn.path("status").asText("GENERATED"));
        task.put("artifact_digest", digest);
        task.put("updated_at", Instant.now().toString());
        String summary = metaIn == null ? "" : metaIn.path("summary").asText("");
        if (!summary.isBlank()) {
            task.put("summary", summary);
        }
        TASKS.put(id, task);
        TASK_ARTIFACTS.put(id, synced);
        return task;
    }

    /** 两项核验结果的结构化摘要（供任务记录与批准记录内嵌）。 */
    static ObjectNode checksNode(CheckResult n, CheckResult c) {
        ObjectNode checks = MAPPER.createObjectNode();
        checks.set("numeric", checkSummary(n));
        checks.set("citation", checkSummary(c));
        return checks;
    }

    private static ObjectNode checkSummary(CheckResult r) {
        ObjectNode node = MAPPER.createObjectNode();
        node.put("checker_id", r.checkerId());
        node.put("checker_version", r.checkerVersion());
        node.put("status", String.valueOf(r.status()));
        node.set("violations", MAPPER.valueToTree(r).path("violations").deepCopy());
        return node;
    }

    /** Agent Hub 种子：11 个内置运行时 agent（含默认 skill 绑定与主编团队）+ 3 个市场目录可安装 agent。 */
    private static void seedAgents() {
        String boot = Instant.now().toString();
        putAgent(AGENT_RUNTIME, "editor-in-chief", "主编编排 Agent", "ORCHESTRATOR", "1.0.0",
                "受控对话流主编：受理任务、澄清确认需求、分派专业 agent、汇总交付",
                new String[]{"任务受理", "需求澄清", "多agent分派", "交付汇总"},
                new String[]{"annual-report", "orchestrator-skill"}, boot);
        putAgent(AGENT_RUNTIME, "material-agent", "资料管理 Agent", "ANALYST", "1.0.0",
                "信源快照装载与页读取（read_page / list_metrics_index），编制资料导入的唯一入口",
                new String[]{"快照装载", "页读取", "指标索引"},
                new String[]{"annual-report", "source-library"}, boot);
        putAgent(AGENT_RUNTIME, "finance-analyst", "财务分析 Agent", "ANALYST", "1.0.0",
                "经营业绩与盈利结构分析：营收、净利润、净息差、成本收入比",
                new String[]{"经营业绩分析", "盈利结构分析"},
                new String[]{"annual-report", "banking-domain", "metric-dictionary"}, boot);
        putAgent(AGENT_RUNTIME, "risk-analyst", "风险与拨备 Agent", "ANALYST", "1.0.0",
                "贷款质量与风险抵补分析：不良余额/不良率、拨备覆盖率，负责变动口径澄清",
                new String[]{"贷款质量分析", "口径澄清"},
                new String[]{"annual-report", "banking-domain"}, boot);
        putAgent(AGENT_RUNTIME, "capital-analyst", "资本管理 Agent", "ANALYST", "1.0.0",
                "资本充足分析：高级法/权重法 CET1 等资本口径",
                new String[]{"资本充足分析"},
                new String[]{"annual-report", "banking-domain", "metric-dictionary"}, boot);
        putAgent(AGENT_RUNTIME, "disclosure-officer", "披露合规 Agent", "ANALYST", "1.0.0",
                "披露规范审查：口径铁律、释义与脚注体例（证监会 2 号准则对齐）",
                new String[]{"披露口径审查", "释义脚注"},
                new String[]{"annual-report", "citation-style"}, boot);
        putAgent(AGENT_RUNTIME, "writer-agent", "长文写作 Agent", "WRITER", "1.0.0",
                "受理长文写作任务书，经 openjiuwen deep agent 自主分解章节写作任务并逐节成稿；"
                        + "写作方法与流程来自 annual-report 技能，核验违规在同一会话内反思修订",
                new String[]{"任务分解", "章节成稿", "反思修订"},
                new String[]{"annual-report", "long-form-writing"}, boot);
        putAgent(AGENT_RUNTIME, "chart-agent", "图表绘制 Agent", "RENDERER", "1.0.0",
                "受控图表绘制：按图表数据文件与骨架配置挂载节位渲染内联 SVG，只渲染不改数、不发明图表",
                new String[]{"图表绘制", "挂载校验"},
                new String[]{"annual-report", "report-skeleton", "agent-division"}, boot);
        putAgent(AGENT_RUNTIME, "numeric-checker", "数值核验器", "CHECKER", NumericChecker.CHECKER_VERSION,
                "N1–N7 数值纪律：指标完整性、单位白名单、数字页内字面存在、派生复算、同比复算、变动口径",
                new String[]{"数值复算", "口径拦截"},
                new String[]{"annual-report", "metric-dictionary"}, boot);
        putAgent(AGENT_RUNTIME, "citation-checker", "引用核验器", "CHECKER", CitationChecker.CHECKER_VERSION,
                "引用溯源：摘录指纹比对、页内字面存在、信源注册表校验",
                new String[]{"引用溯源", "防篡改"},
                new String[]{"annual-report", "citation-style"}, boot);
        putAgent(AGENT_RUNTIME, "approver", "交付批准官", "APPROVER", "1.0.0",
                "具名批准交付：批准记录与工件 digest 绑定，内容一改批准即失效",
                new String[]{"具名批准", "digest 绑定"},
                new String[]{"annual-report"}, boot);

        // 主编团队：内置 9 个专业执行 agent 默认注册给主编编排 Agent。
        // approver 治理独立：批准权不经主编分派（分派链不得覆盖治理链）。
        ArrayNode team = AGENT_RUNTIME.get("editor-in-chief").putArray("team");
        for (String member : new String[]{"material-agent", "finance-analyst", "risk-analyst", "capital-analyst",
                "disclosure-officer", "writer-agent", "chart-agent", "numeric-checker", "citation-checker"}) {
            team.add(member);
        }

        putAgent(AGENT_CATALOG, "peer-benchmark", "同业对标 Agent", "ANALYST", "0.9.0",
                "四大行关键指标横向对标，为报告补充同业排名论断",
                new String[]{"同业对标"},
                new String[]{}, null);
        putAgent(AGENT_CATALOG, "macro-review", "宏观审慎 Agent", "ANALYST", "0.9.0",
                "宏观环境与监管动态审阅，为风险章节补充环境论断",
                new String[]{"宏观审阅"},
                new String[]{}, null);
        putAgent(AGENT_CATALOG, "layout-reviewer", "排版审校 Agent", "ANALYST", "0.9.0",
                "年报版式审校：A4 版式、表格规范、脚注体例",
                new String[]{"版式审校"},
                new String[]{}, null);
    }

    /** 技能注册表种子：8 个技能（与 skills/annual-report 目录一致），可绑定给任意运行时 agent。 */
    private static void seedSkills() {
        putSkill("annual-report", "annual-report 主技能", "skills/annual-report/SKILL.md",
                "年报编制主技能：受控流程编排——资料装载 → 口径澄清 → 分节成稿 → 双重核验 → 具名批准，内含五份 reference（领域/引用/口径/信源/长文写作方法）。",
                new String[]{"编排", "成稿", "核验", "批准"});
        putSkill("long-form-writing", "长文写作方法", "skills/annual-report/references/long-form-writing.md",
                "长文流水线写作方法与流程：任务书解读与分解、分节成稿法、引用标注法、口径句式规范、跨节一致性要求、核验违规修订流程。",
                new String[]{"长文写作", "任务分解", "修订流程"});
        putSkill("report-skeleton", "年报骨架方法", "skills/annual-report/references/report-skeleton.md",
                "真实年报骨架方法（名称保留/深度收缩）：report-structure.json 是节序/标题/静态正文/责任 agent 的唯一事实源，新增银行零代码。",
                new String[]{"骨架配置"});
        putSkill("agent-division", "六工位分工", "skills/annual-report/references/agent-division.md",
                "多 Agent 六工位分工：查数/分析/图表/写作/核验/主编的输入-输出-产物与分派纪律，分派来自骨架配置。",
                new String[]{"多agent协同"});
        putSkill("banking-domain", "银行业领域知识", "skills/annual-report/references/banking-domain.md",
                "银行业术语与监管口径基线：巴塞尔资本口径、不良/拨备/资本概念辨析、监管指标体系。",
                new String[]{"领域知识"});
        putSkill("citation-style", "引用体例规范", "skills/annual-report/references/citation-style.md",
                "引用三元组（digest/page/excerpt）与脚注体例、信源注册表使用规则、摘录指纹计算约定。",
                new String[]{"引用规范"});
        putSkill("metric-dictionary", "指标口径词典", "skills/annual-report/references/metric-dictionary.md",
                "八指标口径词典：单位白名单、变动口径铁律（个百分点 vs %）、派生复算与同比复算规则。",
                new String[]{"指标口径"});
        // P8（2026-11-17）：主编编排技能包——四步规划纪律+alk05 能力卡+案例经验（主编图生成知识层）
        putSkill("orchestrator-skill", "主编编排技能", "skills/orchestrator-skill/SKILL.md",
                "主 Agent 图生成知识包：四步规划纪律（召回→适用性→排序→选择，每步留记录）+ 能力卡四元组（alk05-annual-report.json）+ 失败三态处理 + 案例三例（完整图/裁剪图/口径排除）。",
                new String[]{"候选图生成", "能力选择", "口径排除"});
        putSkill("source-library", "信源库管理", "skills/annual-report/references/source-library.md",
                "一手信源铁律与快照装载流程：页文本读取、digest 校验、二手信源禁入。",
                new String[]{"信源管理"});
    }

    private static void putSkill(String id, String name, String path, String desc, String[] tags) {
        ObjectNode s = MAPPER.createObjectNode();
        s.put("skill_id", id);
        s.put("name", name);
        s.put("path", path);
        s.put("description", desc);
        ArrayNode t = s.putArray("tags");
        for (String tag : tags) {
            t.add(tag);
        }
        SKILL_REGISTRY.put(id, s);
    }

    private static void putAgent(Map<String, ObjectNode> pool, String id, String name, String kind,
                                 String version, String desc, String[] caps, String[] skills, String registeredAt) {
        ObjectNode a = MAPPER.createObjectNode();
        a.put("agent_id", id);
        a.put("name", name);
        a.put("kind", kind);
        a.put("version", version);
        a.put("description", desc);
        a.put("source", registeredAt == null ? "MARKET" : "BUILT_IN");
        a.put("status", registeredAt == null ? "AVAILABLE" : "RUNNING");
        ArrayNode arr = a.putArray("capabilities");
        for (String c : caps) {
            arr.add(c);
        }
        ArrayNode sk = a.putArray("enabled_skills");
        for (String s : skills) {
            sk.add(s);
        }
        if (registeredAt == null) {
            a.putNull("registered_at");
        } else {
            a.put("registered_at", registeredAt);
        }
        a.putNull("last_loaded_task");
        a.putNull("last_loaded_at");
        pool.put(id, a);
    }

    /** 市场目录 → 运行时：安装注册。 */
    private static void handleAgentInstall(HttpExchange ex) throws IOException {
        String id;
        try {
            id = MAPPER.readTree(ex.getRequestBody()).path("agent_id").asText("");
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败"));
            return;
        }
        ObjectNode a = AGENT_CATALOG.remove(id);
        if (a == null) {
            if (AGENT_RUNTIME.containsKey(id)) {
                respond(ex, 409, errorNode("already_installed", id + " 已在运行时"));
            } else {
                respond(ex, 404, errorNode("unknown_agent", "市场目录无此 agent: " + id));
            }
            return;
        }
        a.put("status", "RUNNING");
        a.put("registered_at", Instant.now().toString());
        AGENT_RUNTIME.put(id, a);
        respond(ex, 200, a.deepCopy());
    }

    /** 运行时 agent 按任务动态按需加载：记录最近加载任务与时间。 */
    private static void handleAgentLoad(HttpExchange ex) throws IOException {
        JsonNode body;
        try {
            body = MAPPER.readTree(ex.getRequestBody());
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败"));
            return;
        }
        String id = body.path("agent_id").asText("");
        String taskId = body.path("task_id").asText("");
        ObjectNode a = AGENT_RUNTIME.get(id);
        if (a == null) {
            if (AGENT_CATALOG.containsKey(id)) {
                respond(ex, 409, errorNode("not_installed", id + " 尚未安装注册，请先在 Agent Hub 安装"));
            } else {
                respond(ex, 404, errorNode("unknown_agent", "无此 agent: " + id));
            }
            return;
        }
        a.put("last_loaded_task", taskId);
        a.put("last_loaded_at", Instant.now().toString());
        respond(ex, 200, a.deepCopy());
    }

    /** 技能注册表：GET /v1/skills，含每个技能已注册（绑定）的运行时 agent 反查。 */
    private static void handleSkillsList(HttpExchange ex) throws IOException {
        ArrayNode arr = MAPPER.createArrayNode();
        for (ObjectNode s : SKILL_REGISTRY.values()) {
            ObjectNode out = s.deepCopy();
            String skillId = s.path("skill_id").asText();
            ArrayNode bound = out.putArray("registered_agents");
            for (ObjectNode a : AGENT_RUNTIME.values()) {
                if (hasSkill(a, skillId)) {
                    bound.add(a.path("agent_id").asText());
                }
            }
            out.put("registered_count", bound.size());
            arr.add(out);
        }
        respond(ex, 200, arr);
    }

    private static boolean hasSkill(ObjectNode agent, String skillId) {
        for (JsonNode s : agent.path("enabled_skills")) {
            if (s.asText().equals(skillId)) {
                return true;
            }
        }
        return false;
    }

    /** skill→agent 绑定/解绑：POST /v1/agents/{id}/skills {skill_id, action: bind|unbind}。 */
    private static void handleSkillBind(HttpExchange ex, String path) throws IOException {
        String agentId = path.substring("/v1/agents/".length(), path.length() - "/skills".length());
        JsonNode body;
        try {
            body = MAPPER.readTree(ex.getRequestBody());
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败"));
            return;
        }
        String skillId = body.path("skill_id").asText("");
        String action = body.path("action").asText("");
        ObjectNode a = AGENT_RUNTIME.get(agentId);
        if (a == null) {
            respond(ex, 404, errorNode("unknown_agent", "运行时无此 agent: " + agentId));
            return;
        }
        if (!SKILL_REGISTRY.containsKey(skillId)) {
            respond(ex, 404, errorNode("unknown_skill", "技能注册表无此 skill: " + skillId));
            return;
        }
        ArrayNode skills = (ArrayNode) a.get("enabled_skills");
        switch (action) {
            case "bind" -> {
                if (!hasSkill(a, skillId)) {
                    skills.add(skillId);
                }
            }
            case "unbind" -> {
                for (int i = skills.size() - 1; i >= 0; i--) {
                    if (skills.get(i).asText().equals(skillId)) {
                        skills.remove(i);
                    }
                }
            }
            default -> {
                respond(ex, 400, errorNode("bad_request", "action 须为 bind|unbind"));
                return;
            }
        }
        respond(ex, 200, a.deepCopy());
    }

    /** agent→主编团队注册/退出：POST /v1/agents/editor-in-chief/team {agent_id, action: register|unregister}。 */
    private static void handleTeamRegister(HttpExchange ex) throws IOException {
        JsonNode body;
        try {
            body = MAPPER.readTree(ex.getRequestBody());
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败"));
            return;
        }
        String agentId = body.path("agent_id").asText("");
        String action = body.path("action").asText("");
        if ("editor-in-chief".equals(agentId)) {
            respond(ex, 400, errorNode("bad_request", "主编编排 Agent 不能注册给自己名下"));
            return;
        }
        ObjectNode a = AGENT_RUNTIME.get(agentId);
        if (a == null) {
            if (AGENT_CATALOG.containsKey(agentId)) {
                respond(ex, 409, errorNode("not_installed",
                        agentId + " 尚未安装注册（在市场目录），请先在 Agent Hub 安装"));
            } else {
                respond(ex, 404, errorNode("unknown_agent", "运行时无此 agent: " + agentId));
            }
            return;
        }
        ObjectNode chief = AGENT_RUNTIME.get("editor-in-chief");
        ArrayNode team = (ArrayNode) chief.get("team");
        boolean inTeam = false;
        for (JsonNode m : team) {
            if (m.asText().equals(agentId)) {
                inTeam = true;
                break;
            }
        }
        switch (action) {
            case "register" -> {
                if (!inTeam) {
                    team.add(agentId);
                }
            }
            case "unregister" -> {
                for (int i = team.size() - 1; i >= 0; i--) {
                    if (team.get(i).asText().equals(agentId)) {
                        team.remove(i);
                    }
                }
            }
            default -> {
                respond(ex, 400, errorNode("bad_request", "action 须为 register|unregister"));
                return;
            }
        }
        respond(ex, 200, chief.deepCopy());
    }

    /**
     * 运行时注销：DELETE /v1/agents/{id}。
     * 主编编排 Agent 是运行时根节点，不可注销；市场安装来源的注销后退回市场目录（可重新安装）；
     * 内置与自定义来源的注销即从运行时删除。
     */
    private static void handleAgentUnregister(HttpExchange ex, String id) throws IOException {
        if ("editor-in-chief".equals(id)) {
            respond(ex, 409, errorNode("protected_agent",
                    "主编编排 Agent 是运行时根节点，不可注销（可先在团队中移除其他成员）"));
            return;
        }
        ObjectNode a = AGENT_RUNTIME.remove(id);
        if (a == null) {
            respond(ex, 404, errorNode("unknown_agent", "运行时无此 agent: " + id));
            return;
        }
        // 同步从主编团队移除（若在团队中）
        ObjectNode chief = AGENT_RUNTIME.get("editor-in-chief");
        if (chief != null && chief.has("team")) {
            ArrayNode team = (ArrayNode) chief.get("team");
            for (int i = team.size() - 1; i >= 0; i--) {
                if (team.get(i).asText().equals(id)) {
                    team.remove(i);
                }
            }
        }
        boolean toMarket = "MARKET".equals(a.path("source").asText());
        if (toMarket) {
            a.put("status", "AVAILABLE");
            a.putNull("registered_at");
            a.putNull("last_loaded_task");
            a.putNull("last_loaded_at");
            AGENT_CATALOG.put(id, a);
        }
        ObjectNode out = MAPPER.createObjectNode();
        out.put("status", "UNREGISTERED");
        out.put("agent_id", id);
        out.put("returned_to_market", toMarket);
        respond(ex, 200, out);
    }

    /** 自定义 agent 注册：POST /v1/agents/register {name, kind, description, capabilities[]}。 */
    private static void handleAgentRegister(HttpExchange ex) throws IOException {
        JsonNode body;
        try {
            body = MAPPER.readTree(ex.getRequestBody());
        } catch (Exception e) {
            respond(ex, 400, errorNode("bad_request", "请求体解析失败"));
            return;
        }
        String name = body.path("name").asText("").trim();
        if (name.isEmpty()) {
            respond(ex, 400, errorNode("bad_request", "name 不能为空"));
            return;
        }
        String kind = body.path("kind").asText("ANALYST");
        if (!List.of("ORCHESTRATOR", "ANALYST", "WRITER", "RENDERER", "CHECKER", "APPROVER").contains(kind)) {
            respond(ex, 400, errorNode("bad_request",
                    "kind 须为 ORCHESTRATOR|ANALYST|WRITER|RENDERER|CHECKER|APPROVER"));
            return;
        }
        String desc = body.path("description").asText("");
        if (desc.isBlank()) {
            desc = "自定义 agent（经 Agent Hub 注册进入运行时）";
        }
        ArrayNode caps = MAPPER.createArrayNode();
        if (body.path("capabilities").isArray()) {
            for (JsonNode c : body.path("capabilities")) {
                if (!c.asText().isBlank()) {
                    caps.add(c.asText().trim());
                }
            }
        }
        if (caps.isEmpty()) {
            caps.add("自定义能力");
        }
        String id = String.format("custom-agent-%02d", CUSTOM_SEQ.incrementAndGet());
        ObjectNode a = MAPPER.createObjectNode();
        a.put("agent_id", id);
        a.put("name", name);
        a.put("kind", kind);
        a.put("version", "0.1.0");
        a.put("description", desc);
        a.put("source", "CUSTOM");
        a.put("status", "RUNNING");
        a.set("capabilities", caps);
        a.putArray("enabled_skills");
        a.put("registered_at", Instant.now().toString());
        a.putNull("last_loaded_task");
        a.putNull("last_loaded_at");
        AGENT_RUNTIME.put(id, a);
        respond(ex, 200, a.deepCopy());
    }

    // ———————————————————— 帮手 ————————————————————

    private static Map<String, String> parseQuery(String rawQuery) {
        Map<String, String> out = new HashMap<>();
        if (rawQuery == null || rawQuery.isBlank()) {
            return out;
        }
        for (String pair : rawQuery.split("&")) {
            int eq = pair.indexOf('=');
            if (eq < 0) {
                out.put(URLDecoder.decode(pair, StandardCharsets.UTF_8), "");
            } else {
                out.put(URLDecoder.decode(pair.substring(0, eq), StandardCharsets.UTF_8),
                        URLDecoder.decode(pair.substring(eq + 1), StandardCharsets.UTF_8));
            }
        }
        return out;
    }

    static String sha256(String s) {
        return sha256(s.getBytes(StandardCharsets.UTF_8));
    }

    static String sha256(byte[] bytes) {
        try {
            MessageDigest md = MessageDigest.getInstance("SHA-256");
            return HexFormat.of().formatHex(md.digest(bytes));
        } catch (Exception e) {
            throw new IllegalStateException(e);
        }
    }

    private static ObjectNode errorNode(String code, String message) {
        ObjectNode n = MAPPER.createObjectNode();
        n.put("error", code);
        n.put("message", message);
        return n;
    }

    private static void respond(HttpExchange ex, int code, JsonNode body) throws IOException {
        ex.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
        respondRaw(ex, code, MAPPER.writeValueAsBytes(body));
    }

    private static void respondRaw(HttpExchange ex, int code, byte[] bytes) throws IOException {
        ex.sendResponseHeaders(code, bytes.length);
        try (OutputStream os = ex.getResponseBody()) {
            os.write(bytes);
        }
    }
}
