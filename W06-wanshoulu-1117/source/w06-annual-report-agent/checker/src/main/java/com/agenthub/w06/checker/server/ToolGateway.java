/*
 * ToolGateway — MCP 风格联网工具网关（0.3.5）。
 *
 * 设计约束（泛化性）：
 *  · 工具注册表配置驱动：deploy/tools.json 是唯一事实源（清单/参数/默认参数/样例参数），
 *    前端 MCP 资产页读 GET /v1/tools 渲染，前端零硬编码；
 *  · 调用统一入口 POST /v1/tools/call {tool_name, arguments}（MCP 风格）；
 *  · 执行器纯 JDK HttpClient + Apache PDFBox（jar 随 deploy/libs 分发，构建不依赖外网）；
 *  · 全部工具 fail-closed：任何失败返回 200 + {ok:false, error:{code,message}}，绝不抛裸异常；
 *  · 核验链闭环：联网抓取产物必须经 snapshot_commit 落盘快照 + sha256 注册进信源注册表
 *    才能被引用核验（C1/C2）与数值核验（N3）接受——「抓得到」与「可引用」之间有纪律门。
 *
 * 对外契约：
 *  GET  /v1/tools       注册表全文（gateway 名称 / 调用入口 / 工具清单）
 *  POST /v1/tools/call  {tool_name, arguments} → 工具执行结果（ok 标志）
 *
 * 工具清单（定义在 tools.json，执行器在本类 switch 分派）：
 *  web_search      联网搜索（Bing HTML 端点为主引擎，DuckDuckGo 备选，按配置顺序容错）
 *  web_fetch       抓取 URL 并抽取正文（HTML 去标签，限长）
 *  pdf_extract     下载 PDF 逐页抽取文本（PDFBox，单次最多 20 页）
 *  snapshot_commit 抓取产物落盘为快照页（/snapshot/live）+ sha256 注册进信源注册表
 */
package com.agenthub.w06.checker.server;

import com.agenthub.w06.contract.SourceAuthority;
import com.agenthub.w06.contract.SourceEntry;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.pdfbox.Loader;
import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.text.PDFTextStripper;

import java.io.IOException;
import java.net.URI;
import java.net.URLDecoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class ToolGateway {
    /** 搜索引擎直连域名（www.bing.com 会 302 到 cn.bing.com，直连国内版更稳）。 */
    private static final String BING_ENDPOINT = "https://cn.bing.com/search?q=%s&mkt=zh-CN";
    private static final String DDG_ENDPOINT = "https://html.duckduckgo.com/html/?q=%s";
    private static final String UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
            + "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";
    private static final int PDF_MAX_BYTES = 60 * 1024 * 1024;   // 单次下载 PDF 上限
    private static final int PDF_PAGES_PER_CALL = 20;            // 单次调用抽取页数上限
    private static final int PAGE_TEXT_CAP = 6000;               // 单页文本返回上限
    private static final int FETCH_TEXT_DEFAULT = 4000;          // web_fetch 默认正文限长
    private static final int FETCH_TEXT_CAP = 20000;             // web_fetch 正文限长上限

    /** 快照落盘专用银行目录名（/snapshot/live）：与冻结演示快照 {cmb,icbc,ccb} 完全隔离。 */
    static final String LIVE_BANK = "live";

    /** 工具注册表：tool_id → 定义（加载自 deploy/tools.json，顺序保持配置文件顺序）。 */
    static final Map<String, ObjectNode> REGISTRY = new LinkedHashMap<>();

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .followRedirects(HttpClient.Redirect.NORMAL)
            .connectTimeout(Duration.ofSeconds(10))
            .build();

    private ToolGateway() {
    }

    /** 注册表装载（deploy/tools.json）。文件缺失仅 WARN（/v1/tools 返回空注册表），与 loadReportStructure 同风格。 */
    static void load(Path staticDir) throws IOException {
        Path f = staticDir.resolve("tools.json");
        if (!Files.isRegularFile(f)) {
            System.out.println("WARN: 未找到 tools.json，/v1/tools 返回空注册表");
            return;
        }
        REGISTRY.clear();
        for (JsonNode t : CheckerHttpServer.MAPPER.readTree(f.toFile()).path("tools")) {
            REGISTRY.put(t.path("tool_id").asText(), (ObjectNode) t);
        }
        System.out.println("ToolGateway ready: tools=" + REGISTRY.keySet());
    }

    /** GET /v1/tools：注册表全文 + 调用入口说明。 */
    static ObjectNode list() {
        ObjectNode out = CheckerHttpServer.MAPPER.createObjectNode();
        out.put("gateway", "w06-tool-gateway");
        out.put("registry_source", "deploy/tools.json（配置驱动，前端零硬编码）");
        out.put("call_endpoint", "POST /v1/tools/call");
        out.put("body_format", "{\"tool_name\": \"…\", \"arguments\": {…}}");
        out.put("discipline", "联网抓取产物必须经 snapshot_commit 落盘 + sha256 注册后方可被引用核验（C1/C2）接受");
        ArrayNode arr = out.putArray("tools");
        for (ObjectNode t : REGISTRY.values()) {
            ObjectNode n = t.deepCopy();
            n.put("ep", "POST /v1/tools/call");
            arr.add(n);
        }
        return out;
    }

    /**
     * POST /v1/tools/call：MCP 风格统一入口。
     * 工具级失败一律 200 + {ok:false, error:{code,message}}（fail-closed，调用方拿到的永远是合法 JSON）。
     */
    static ObjectNode call(JsonNode body, Path snapshotDir,
                           CheckerHttpServer.SnapshotHttpResolver resolver, List<SourceEntry> sources) {
        String tool = body.path("tool_name").asText("").trim();
        if (tool.isEmpty()) {
            return err("bad_request", "tool_name 必填（GET /v1/tools 查看注册表）");
        }
        if (!REGISTRY.containsKey(tool)) {
            return err("unknown_tool", "注册表无工具: " + tool + "（GET /v1/tools 查看可用工具）");
        }
        try {
            return switch (tool) {
                case "web_search" -> webSearch(argsOf(body));
                case "web_fetch" -> webFetch(argsOf(body));
                case "pdf_extract" -> pdfExtract(argsOf(body));
                case "snapshot_commit" -> snapshotCommit(argsOf(body), snapshotDir, resolver, sources);
                case "table_extract" -> TableTools.extract(argsOf(body), resolver);
                case "table_analyze" -> TableTools.analyze(argsOf(body), resolver);
                default -> err("unknown_tool", "工具已在注册表但无执行器: " + tool);
            };
        } catch (Exception e) {
            return err("tool_error", tool + " 执行失败: " + rootMessage(e));
        }
    }

    private static JsonNode argsOf(JsonNode body) {
        return body.path("arguments");
    }

    // ———————————————————— web_search ————————————————————

    /** 联网搜索：按 tools.json 配置的引擎顺序容错，首个有结果的引擎即返回。 */
    private static ObjectNode webSearch(JsonNode args) throws Exception {
        String query = args.path("query").asText("").trim();
        if (query.isEmpty()) {
            return err("bad_request", "arguments.query 必填");
        }
        int max = Math.max(1, Math.min(10, args.path("max_results").asInt(5)));
        ObjectNode def = REGISTRY.get("web_search");
        int timeoutMs = def.path("default_args").path("timeout_ms").asInt(10000);
        List<String> tried = new ArrayList<>();
        for (JsonNode e : def.path("default_args").path("engines")) {
            String engine = e.asText().trim();
            tried.add(engine);
            try {
                ArrayNode results = switch (engine) {
                    case "bing" -> searchBing(query, max, timeoutMs);
                    case "duckduckgo" -> searchDDG(query, max, timeoutMs);
                    default -> null;
                };
                if (results != null && !results.isEmpty()) {
                    ObjectNode ok = base("web_search");
                    ok.put("ok", true);
                    ok.put("query", query);
                    ok.put("engine", engine);
                    ok.put("result_count", results.size());
                    ok.set("results", results);
                    return ok;
                }
            } catch (Exception ex) {
                // 单引擎失败 → 尝试下一引擎（最终仍失败则 fail-closed）
            }
        }
        return err("all_engines_failed", "全部引擎无结果或不可达: " + String.join(", ", tried));
    }

    private static ArrayNode searchBing(String query, int max, int timeoutMs) throws Exception {
        String html = httpGetText(String.format(BING_ENDPOINT, urlEncode(query)), timeoutMs);
        // Bing 国内版结果块：<li class="b_algo">…<h2><a href="URL">TITLE</a></h2>…<p>SNIPPET</p>
        Matcher a = Pattern.compile("<h2[^>]*>\\s*<a[^>]*href=\"([^\"]+)\"[^>]*>(.*?)</a>", Pattern.DOTALL)
                .matcher(html);
        ArrayNode out = CheckerHttpServer.MAPPER.createArrayNode();
        while (a.find() && out.size() < max) {
            String link = unescape(a.group(1)).trim();
            String title = stripTags(a.group(2));
            if (title.isEmpty() || link.isEmpty() || !link.matches("^https?://.*")) {
                continue;
            }
            out.add(resultNode(title, link, ""));
        }
        // 摘要单独补齐（与标题块按顺序对齐，尽力而为）
        Matcher s = Pattern.compile("<p[^>]*class=\"[^\"]*b_caption[^\"]*\"[^>]*>(.*?)</p>", Pattern.DOTALL)
                .matcher(html);
        for (int i = 0; i < out.size() && s.find(); i++) {
            ((ObjectNode) out.get(i)).put("snippet", stripTags(s.group(1)));
        }
        return out;
    }

    private static ArrayNode searchDDG(String query, int max, int timeoutMs) throws Exception {
        String html = httpGetText(String.format(DDG_ENDPOINT, urlEncode(query)), timeoutMs);
        // DuckDuckGo HTML 版：<a class="result__a" href="/l/?uddg=ENCODED">TITLE</a>
        Matcher a = Pattern.compile("<a[^>]*class=\"result__a\"[^>]*href=\"([^\"]+)\"[^>]*>(.*?)</a>", Pattern.DOTALL)
                .matcher(html);
        Matcher sn = Pattern.compile("<a[^>]*class=\"result__snippet\"[^>]*>(.*?)</a>", Pattern.DOTALL)
                .matcher(html);
        ArrayNode out = CheckerHttpServer.MAPPER.createArrayNode();
        while (a.find() && out.size() < max) {
            String link = unescape(a.group(1)).trim();
            Matcher u = Pattern.compile("[?&]uddg=([^&]+)").matcher(link);
            if (u.find()) {
                link = URLDecoder.decode(u.group(1), StandardCharsets.UTF_8);
            }
            String title = stripTags(a.group(2));
            String snippet = sn.find() ? stripTags(sn.group(1)) : "";
            if (title.isEmpty() || link.isEmpty() || !link.matches("^https?://.*")) {
                continue;
            }
            out.add(resultNode(title, link, snippet));
        }
        return out;
    }

    // ———————————————————— web_fetch ————————————————————

    /** 抓取 URL 抽取正文：HTML 去标签（script/style 剔除），限长截断。 */
    private static ObjectNode webFetch(JsonNode args) throws Exception {
        String url = args.path("url").asText("").trim();
        if (url.isEmpty()) {
            return err("bad_request", "arguments.url 必填");
        }
        String lower = url.toLowerCase();
        if (!lower.startsWith("http://") && !lower.startsWith("https://")) {
            return err("bad_request", "url 必须是 http/https: " + url);
        }
        int maxChars = Math.max(200, Math.min(FETCH_TEXT_CAP, args.path("max_chars").asInt(FETCH_TEXT_DEFAULT)));
        ObjectNode def = REGISTRY.get("web_fetch");
        int timeoutMs = def.path("default_args").path("timeout_ms").asInt(15000);
        HttpResponse<byte[]> resp = send(url, timeoutMs);
        String ctype = resp.headers().firstValue("content-type").orElse("");
        String text;
        try {
            text = new String(resp.body(), charsetOf(ctype));
        } catch (Exception e) {
            text = new String(resp.body(), StandardCharsets.UTF_8);
        }
        ObjectNode ok = base("web_fetch");
        ok.put("ok", true);
        ok.put("url", url);
        ok.put("status", resp.statusCode());
        ok.put("content_type", ctype);
        if (ctype.contains("html")) {
            Matcher t = Pattern.compile("<title[^>]*>(.*?)</title>", Pattern.CASE_INSENSITIVE | Pattern.DOTALL)
                    .matcher(text);
            ok.put("title", t.find() ? stripTags(t.group(1)) : "");
            text = stripHtml(text);
        }
        boolean truncated = text.length() > maxChars;
        if (truncated) {
            text = text.substring(0, maxChars);
        }
        ok.put("text", text);
        ok.put("text_length", text.length());
        ok.put("truncated", truncated);
        ok.put("fetched_at", Instant.now().toString());
        return ok;
    }

    // ———————————————————— pdf_extract ————————————————————

    /** 下载 PDF 并逐页抽取文本（PDFBox）。单次最多 PDF_PAGES_PER_CALL 页，返回总页数供续传。 */
    private static ObjectNode pdfExtract(JsonNode args) throws Exception {
        String url = args.path("url").asText("").trim();
        if (url.isEmpty()) {
            return err("bad_request", "arguments.url 必填");
        }
        String lower = url.toLowerCase();
        if (!lower.startsWith("http://") && !lower.startsWith("https://")) {
            return err("bad_request", "url 必须是 http/https: " + url);
        }
        ObjectNode def = REGISTRY.get("pdf_extract");
        int timeoutMs = def.path("default_args").path("timeout_ms").asInt(60000);
        byte[] pdf = httpGetBytes(url, timeoutMs);
        if (pdf.length > PDF_MAX_BYTES) {
            return err("too_large", "PDF 超过 " + (PDF_MAX_BYTES / 1024 / 1024) + "MB 上限: " + pdf.length + " bytes");
        }
        try (PDDocument doc = Loader.loadPDF(pdf)) {
            int count = doc.getNumberOfPages();
            int from = Math.max(1, args.path("from_page").asInt(1));
            int to = Math.min(count, args.path("to_page").asInt(Math.min(count, from + PDF_PAGES_PER_CALL - 1)));
            if (to - from + 1 > PDF_PAGES_PER_CALL) {
                to = from + PDF_PAGES_PER_CALL - 1;
            }
            ArrayNode pages = CheckerHttpServer.MAPPER.createArrayNode();
            PDFTextStripper ts = new PDFTextStripper();
            for (int p = from; p <= to; p++) {
                ts.setStartPage(p);
                ts.setEndPage(p);
                String text = ts.getText(doc);
                boolean truncated = text.length() > PAGE_TEXT_CAP;
                if (truncated) {
                    text = text.substring(0, PAGE_TEXT_CAP);
                }
                ObjectNode pg = pages.addObject();
                pg.put("page", p);
                pg.put("chars", text.length());
                pg.put("truncated", truncated);
                pg.put("text", text);
            }
            ObjectNode ok = base("pdf_extract");
            ok.put("ok", true);
            ok.put("url", url);
            ok.put("pdf_bytes", pdf.length);
            ok.put("total_page_count", count);
            ok.put("returned_pages", pages.size());
            ok.set("pages", pages);
            if (to < count) {
                ok.put("hint", "本次返回 P" + from + "-P" + to + "，剩余页请用 from_page/to_page 分批抽取");
            }
            return ok;
        }
    }

    // ———————————————————— snapshot_commit ————————————————————

    /**
     * snapshot_commit 工具：抓取产物落盘为快照页 + sha256 注册进信源注册表。
     * 参数校验后统一走 commitSource —— 与 workbench 本地文档导入共用同一落盘注册纪律门。
     */
    private static ObjectNode snapshotCommit(JsonNode args, Path snapshotDir,
                                             CheckerHttpServer.SnapshotHttpResolver resolver,
                                             List<SourceEntry> sources) throws IOException {
        String url = args.path("url").asText("").trim();
        String title = args.path("title").asText("").trim();
        if (url.isEmpty()) {
            return err("bad_request", "arguments.url 必填");
        }
        if (title.isEmpty()) {
            return err("bad_request", "arguments.title 必填");
        }
        String authority = args.path("authority").asText("TERTIARY_AGGREGATED").trim().toUpperCase();
        try {
            SourceAuthority.valueOf(authority);
        } catch (IllegalArgumentException e) {
            return err("bad_request", "authority 无效: " + authority
                    + "（可用: PRIMARY_OFFICIAL / SECONDARY_INSTITUTIONAL / TERTIARY_AGGREGATED）");
        }
        JsonNode pagesIn = args.path("pages");
        if (!pagesIn.isArray() || pagesIn.isEmpty()) {
            return err("bad_request", "arguments.pages 必填且为数组: [{\"text\": \"页文本\", \"origin_page\": 原始页码?}]");
        }
        List<String> bodies = new ArrayList<>();
        List<Integer> originPages = new ArrayList<>();
        for (int i = 0; i < pagesIn.size(); i++) {
            JsonNode p = pagesIn.get(i);
            String text = p.path("text").asText("");
            if (text.isBlank()) {
                return err("bad_request", "pages[" + i + "].text 为空");
            }
            if (!text.endsWith("\n")) {
                text = text + "\n";
            }
            bodies.add(text);
            originPages.add(p.path("origin_page").asInt(0) > 0 ? p.path("origin_page").asInt() : i + 1);
        }
        return commitSource(LIVE_BANK, url, title, authority, bodies, originPages,
                "snapshot_commit 工具实时抓取落盘（MCP 工具层）：抓取产物经 sha256 注册后方可引用",
                snapshotDir, resolver, sources);
    }

    /**
     * 本地文档导入（workbench「导入资料」）与 snapshot_commit 共用的落盘注册核心。
     *
     * 写入 /snapshot/{bank}（导入与抓取统一落 live 独立快照集）：冻结演示快照 {cmb,icbc,ccb}
     * 与其 manifest.sources 的 sha256 指纹零改动；只「追加」新页与新信源，页码从该快照集
     * 现有最大页码 +1 顺序分配。落盘后立刻重建运行时索引并追加信源注册表 → C1/C2 引用核验
     * 与 N3 数值核验即刻可用（fail-closed 纪律门，本地导入与联网抓取同规，无需重启容器）。
     */
    static ObjectNode commitSource(String bank, String url, String title, String authority,
                                   List<String> bodies, List<Integer> originPages, String note,
                                   Path snapshotDir, CheckerHttpServer.SnapshotHttpResolver resolver,
                                   List<SourceEntry> sources) throws IOException {
        synchronized (ToolGateway.class) { // 并发提交互斥（manifest 读改写 + 页码分配）
            Path liveDir = snapshotDir.resolve(bank);
            Path pagesDir = liveDir.resolve("pages");
            Files.createDirectories(pagesDir);
            Path manifestF = liveDir.resolve("manifest.json");
            ObjectNode manifest = Files.isRegularFile(manifestF)
                    ? (ObjectNode) CheckerHttpServer.MAPPER.readTree(manifestF.toFile())
                    : emptyLiveManifest();
            if (!(manifest.get("sources") instanceof ArrayNode)) {
                manifest.set("sources", CheckerHttpServer.MAPPER.createArrayNode());
            }
            if (!(manifest.get("page_digests") instanceof ObjectNode)) {
                manifest.set("page_digests", CheckerHttpServer.MAPPER.createObjectNode());
            }

            // 内容指纹仅由页文本决定（与快照页码无关）——幂等判定的唯一依据
            StringBuilder concat = new StringBuilder();
            for (String body : bodies) {
                concat.append(body);
            }
            String contentSha = CheckerHttpServer.sha256(concat.toString().getBytes(StandardCharsets.UTF_8));
            String sourceId = "live-" + contentSha.substring(0, 8);

            // 幂等：manifest 已有同 content_sha256 的信源 → 读回既有登记直接返回，零落盘零追加
            ArrayNode srcArr = (ArrayNode) manifest.get("sources");
            JsonNode existedSrc = null;
            for (JsonNode s : srcArr) {
                if (contentSha.equals(s.path("content_sha256").asText())) {
                    existedSrc = s;
                    break;
                }
            }
            ArrayNode allocated = CheckerHttpServer.MAPPER.createArrayNode();
            if (existedSrc != null) {
                int i = 0;
                for (JsonNode pno : existedSrc.path("pages")) {
                    int p = pno.asInt();
                    ObjectNode pg = allocated.addObject();
                    pg.put("snapshot_page", p);
                    pg.put("origin_page", i < originPages.size() ? originPages.get(i) : p);
                    pg.put("digest", manifest.path("page_digests").path(String.valueOf(p)).asText(""));
                    i++;
                }
                return okSnapshot(bank, existedSrc.path("source_id").asText(), title, authority, url,
                        contentSha, allocated, true);
            }

            // 新注册：页码从现有最大页码 +1 顺序分配（以落盘文件为准，跨提交单调递增）
            int maxNo = 0;
            try (var stream = Files.list(pagesDir)) {
                for (Path f : stream.filter(x -> x.getFileName().toString().matches("page-\\d+\\.txt")).toList()) {
                    maxNo = Math.max(maxNo, Integer.parseInt(
                            f.getFileName().toString().replace("page-", "").replace(".txt", "")));
                }
            }
            int no = maxNo;
            for (String body : bodies) {
                no++;
                Files.writeString(pagesDir.resolve("page-" + no + ".txt"), body, StandardCharsets.UTF_8);
                ObjectNode pg = allocated.addObject();
                pg.put("snapshot_page", no);
                pg.put("origin_page", originPages.get(no - maxNo - 1));
                pg.put("digest", CheckerHttpServer.sha256(body));
            }
            ObjectNode srcNode = CheckerHttpServer.MAPPER.createObjectNode();
            srcNode.put("source_id", sourceId);
            srcNode.put("title", title);
            srcNode.put("authority", authority);
            srcNode.put("origin_url", url);
            srcNode.put("fetched_at", Instant.now().toString());
            srcNode.put("content_sha256", contentSha);
            ArrayNode pagesArr = srcNode.putArray("pages"); // putArray 有替换语义，取引用一次逐个 add
            for (JsonNode pg : allocated) {
                pagesArr.add(pg.path("snapshot_page").asInt());
            }
            srcNode.put("note", note);
            srcArr.add(srcNode);
            ObjectNode digests = (ObjectNode) manifest.get("page_digests");
            for (JsonNode pg : allocated) {
                digests.put(String.valueOf(pg.path("snapshot_page").asInt()), pg.path("digest").asText());
            }
            Files.writeString(manifestF,
                    CheckerHttpServer.MAPPER.writerWithDefaultPrettyPrinter().writeValueAsString(manifest) + "\n",
                    StandardCharsets.UTF_8);

            // 运行时注册：重建 live 快照索引 + 信源注册表追加（核验器即刻可见，无需重启）
            CheckerHttpServer.BankSnap live = new CheckerHttpServer.BankSnap(liveDir, bank);
            resolver.banks.put(bank, live);
            for (String sha : live.sourcePages.keySet()) {
                resolver.sourceBank.put(sha, bank);
            }
            // SourceEntry 契约为 7 字段注册表形态；note 仅随 manifest 落盘，运行时注册前剥离
            ObjectNode regNode = srcNode.deepCopy();
            regNode.remove("note");
            sources.add(CheckerHttpServer.MAPPER.treeToValue(regNode, SourceEntry.class));

            return okSnapshot(bank, sourceId, title, authority, url, contentSha, allocated, false);
        }
    }

    /** 落盘注册统一响应（snapshot_commit 工具 / 本地文档导入共用；新注册 / 幂等读回同构）。 */
    static ObjectNode okSnapshot(String bank, String sourceId, String title, String authority, String url,
            String contentSha, ArrayNode allocated, boolean already) {
        ObjectNode ok = base("snapshot_commit");
        ok.put("ok", true);
        ok.put("bank", bank);
        ok.put("source_id", sourceId);
        ok.put("title", title);
        ok.put("authority", authority);
        ok.put("origin_url", url);
        ok.put("content_sha256", contentSha);
        ok.set("snapshot_pages", allocated);
        ok.put("already_registered", already);
        ok.put("snapshot_dir", "/snapshot/" + bank);
        ok.put("note", already
                ? "同内容信源已注册（幂等读回 manifest 既有登记，未写入任何新页文件）"
                : "已落盘 /snapshot/" + bank + " 并注册进信源注册表：引用核验 C1/C2 即刻可用；冻结快照（cmb/icbc/ccb）零改动");
        return ok;
    }

    // ———————————————————— 本地文档导入（workbench 资料导入通道）————————————————————

    /**
     * 本地导入文件解析（不落盘，落盘统一走 commitSource）：
     *  · pdf —— PDFBox 逐页抽取（与 pdf_extract 同引擎，单文件最多 PDF_PAGES_PER_CALL 页）；
     *  · 文本类（txt/md/json/…）—— UTF-8 解码，按段落边界切成 ≤2500 字符的快照页。
     * 返回 {ok, error?, kind, total_pages?, pages:[{text, origin_page}]}。
     */
    static ObjectNode parseUploadFile(String filename, byte[] bytes) {
        ObjectNode out = CheckerHttpServer.MAPPER.createObjectNode();
        String name = filename == null ? "" : filename.toLowerCase();
        if (name.endsWith(".pdf")) {
            try (PDDocument doc = Loader.loadPDF(bytes)) {
                PDFTextStripper stripper = new PDFTextStripper();
                stripper.setSortByPosition(true);
                int total = doc.getNumberOfPages();
                int cap = Math.min(total, PDF_PAGES_PER_CALL);
                ArrayNode pages = out.putArray("pages");
                for (int p = 1; p <= cap; p++) {
                    stripper.setStartPage(p);
                    stripper.setEndPage(p);
                    String text = stripper.getText(doc);
                    if (text == null || text.isBlank()) {
                        continue;
                    }
                    ObjectNode pg = pages.addObject();
                    pg.put("text", text.strip());
                    pg.put("origin_page", p);
                }
                out.put("kind", "pdf");
                out.put("total_pages", total);
                if (pages.isEmpty()) {
                    out.put("ok", false);
                    out.put("error", "PDF 未抽取到任何文本（可能是扫描件）");
                } else {
                    out.put("ok", true);
                    if (total > cap) {
                        out.put("truncated", "PDF 共 " + total + " 页，单次最多导入 " + cap + " 页");
                    }
                }
                return out;
            } catch (Exception e) {
                out.put("ok", false);
                out.put("error", "PDF 解析失败: " + e.getMessage());
                return out;
            }
        }
        String text = new String(bytes, StandardCharsets.UTF_8);
        if (text.startsWith("﻿")) {
            text = text.substring(1);
        }
        if (text.contains("\u0000")) {
            out.put("ok", false);
            out.put("error", "不支持的二进制文件类型（仅支持 pdf / txt / md / json 等文本文件）");
            return out;
        }
        if (text.isBlank()) {
            out.put("ok", false);
            out.put("error", "文件内容为空");
            return out;
        }
        List<String> chunks = splitTextPages(text, 2500);
        ArrayNode pages = out.putArray("pages");
        for (int i = 0; i < chunks.size(); i++) {
            ObjectNode pg = pages.addObject();
            pg.put("text", chunks.get(i));
            pg.put("origin_page", i + 1);
        }
        out.put("ok", true);
        out.put("kind", "text");
        out.put("total_pages", chunks.size());
        return out;
    }

    /** 文本分页：优先按空行段落边界聚合成 ≤cap 字符的页，超长段落硬切。 */
    private static List<String> splitTextPages(String text, int cap) {
        List<String> pages = new ArrayList<>();
        StringBuilder cur = new StringBuilder();
        for (String para : text.strip().split("\\n\\s*\\n")) {
            String p = para.strip();
            while (p.length() > cap) { // 超长段落（如单行 JSON）硬切
                int cut = p.lastIndexOf('\n', cap);
                if (cut < cap / 2) {
                    cut = cap;
                }
                pages.add(p.substring(0, cut).strip());
                p = p.substring(cut).strip();
            }
            if (p.isEmpty()) {
                continue;
            }
            if (cur.length() > 0 && cur.length() + p.length() + 2 > cap) {
                pages.add(cur.toString());
                cur.setLength(0);
            }
            if (cur.length() > 0) {
                cur.append("\n\n");
            }
            cur.append(p);
        }
        if (cur.length() > 0) {
            pages.add(cur.toString());
        }
        return pages;
    }

    private static ObjectNode emptyLiveManifest() {
        ObjectNode n = CheckerHttpServer.MAPPER.createObjectNode();
        n.put("snapshot_version", "live@0.3.5");
        n.put("bank", LIVE_BANK);
        n.put("company", "W06 联网抓取快照库（MCP 工具层落盘）");
        n.put("report_period", "");
        n.put("generated_at", Instant.now().toString());
        n.put("source_note", "全部信源来自 /v1/tools/call 联网抓取并经 snapshot_commit 落盘注册；冻结演示快照（cmb/icbc/ccb）不受影响");
        n.putArray("sources");
        n.putObject("page_digests");
        n.putArray("entities");
        n.putObject("metric_index");
        return n;
    }

    // ———————————————————— HTTP / 解析辅助 ————————————————————

    private static HttpResponse<byte[]> send(String url, int timeoutMs) throws Exception {
        HttpRequest req = HttpRequest.newBuilder(URI.create(url))
                .timeout(Duration.ofMillis(timeoutMs))
                .header("User-Agent", UA)
                .header("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
                .header("Accept", "text/html,application/xhtml+xml,application/pdf;q=0.9,*/*;q=0.8")
                .GET().build();
        HttpResponse<byte[]> resp = HTTP.send(req, HttpResponse.BodyHandlers.ofByteArray());
        if (resp.statusCode() >= 400) {
            throw new IllegalStateException("HTTP " + resp.statusCode() + " " + url);
        }
        return resp;
    }

    private static String httpGetText(String url, int timeoutMs) throws Exception {
        return new String(httpGetBytes(url, timeoutMs), StandardCharsets.UTF_8);
    }

    private static byte[] httpGetBytes(String url, int timeoutMs) throws Exception {
        return send(url, timeoutMs).body();
    }

    private static Charset charsetOf(String contentType) {
        if (contentType != null && contentType.toLowerCase().contains("charset=")) {
            String cs = contentType.substring(contentType.toLowerCase().indexOf("charset=") + 8)
                    .replaceAll("[\"']", "").split("[;\\s]")[0].trim();
            if (!cs.isEmpty()) {
                try {
                    return Charset.forName(cs);
                } catch (Exception ignored) {
                    // 非法 charset 名 → 回退 UTF-8
                }
            }
        }
        return StandardCharsets.UTF_8;
    }

    private static ObjectNode resultNode(String title, String url, String snippet) {
        ObjectNode r = CheckerHttpServer.MAPPER.createObjectNode();
        r.put("title", title);
        r.put("url", url);
        r.put("snippet", snippet);
        return r;
    }

    /** HTML → 正文：剔除 script/style/noscript 与全部标签，实体解码，压缩空白。 */
    private static String stripHtml(String html) {
        String s = html.replaceAll("(?is)<(script|style|noscript)[^>]*>.*?</\\1>", " ");
        s = s.replaceAll("(?s)<[^>]+>", " ");
        s = unescape(s);
        return s.replaceAll("[\\t\\r ]+", " ").replaceAll("\\n\\s*\\n+", "\n").trim();
    }

    private static String stripTags(String s) {
        return unescape(s.replaceAll("(?s)<[^>]+>", "")).trim();
    }

    private static String unescape(String s) {
        return s.replace("&amp;", "&").replace("&lt;", "<").replace("&gt;", ">")
                .replace("&quot;", "\"").replace("&#39;", "'").replace("&#x27;", "'")
                .replace("&nbsp;", " ").replace("&middot;", "·").replace("&mdash;", "—")
                .replace("&hellip;", "…").replace("&copy;", "©").replace("&reg;", "®");
    }

    private static String urlEncode(String s) {
        return java.net.URLEncoder.encode(s, StandardCharsets.UTF_8);
    }

    private static String rootMessage(Throwable e) {
        String msg = String.valueOf(e.getMessage());
        if (msg == null || msg.isBlank()) {
            return e.getClass().getSimpleName();
        }
        return msg.length() > 300 ? msg.substring(0, 300) : msg;
    }

    private static ObjectNode base(String tool) {
        ObjectNode n = CheckerHttpServer.MAPPER.createObjectNode();
        n.put("tool", tool);
        return n;
    }

    private static ObjectNode err(String code, String message) {
        ObjectNode n = CheckerHttpServer.MAPPER.createObjectNode();
        n.put("ok", false);
        ObjectNode e = n.putObject("error");
        e.put("code", code);
        e.put("message", message);
        return n;
    }
}
