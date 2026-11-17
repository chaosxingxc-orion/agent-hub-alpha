/*
 * MaterialAgent — 资料角色（W06 四角色之一）
 *
 * 职责（v2 方案 §3.3）：
 *  - 只读挂载的固定年报快照（manifest + 分页文本），无任何网络能力；
 *  - 按 metrics/analysis 请求返回指定页的原文片段与 digest；
 *  - 缺页/缺材料时返回明确的 material_missing，不编造。
 *
 * 照 multi-deep-research-demo agent-search 的模式：库层 ReActAgent +
 * LocalFunction 工具注册；wrapper 层（agents 模块的 runtime 包）负责 A2A 暴露。
 */

package com.agenthub.w06.agents.material;

import com.agenthub.w06.contract.MetricKey;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.openjiuwen.core.foundation.tool.Tool;
import com.openjiuwen.core.foundation.tool.ToolCard;
import com.openjiuwen.core.foundation.tool.function.LocalFunction;
import com.openjiuwen.core.runner.Runner;
import java.util.HashMap;
import com.openjiuwen.core.singleagent.agents.ReActAgent;
import com.openjiuwen.core.singleagent.agents.ReActAgentConfig;
import com.openjiuwen.core.singleagent.schema.AgentCard;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * 库层工厂：构建资料 agent。与上游示例同构 —— 工具注册在
 * {@code AbilityManager}，远端可见性由 wrapper 层的 A2A 装配完成。
 */
public final class MaterialAgentFactory {

    /** 快照 manifest 条目：一页 = 一个可寻址证据单元。 */
    public record SnapshotPage(String digest, int page, String text) {}

    /**
     * 构建资料 agent。
     *
     * @param props      配置（模型、迭代上限）
     * @param snapshotDir 固定年报快照目录（manifest.json + pages/page-NNN.txt）
     */
    public static ReActAgent build(MaterialAgentProperties props, Path snapshotDir) {
        props.requireConfigured();
        SnapshotStore store = new SnapshotStore(snapshotDir);

        ReActAgentConfig config = ReActAgentConfig.builder()
                .maxIterations(props.getMaxIterations())
                .sysOperationId(props.getSysOperationId())
                .promptTemplate(List.of(Map.of("role", "system", "content", systemPrompt())))
                .build()
                .configureModelClient(
                        props.getProvider(), props.getApiKey(), props.getApiBase(),
                        props.getModelName(), props.isSslVerify());
        AgentCard card = AgentCard.builder()
                .id(props.getAgentId())
                .name("MaterialAgent")
                .description("资料角色：从固定年报快照检索原文片段，返回 digest+page+excerpt；缺材料如实报告")
                .build();
        ReActAgent agent = new ReActAgent(card);
        agent.configure(config);
        registerReadPageTool(agent, store);
        registerListMetricsIndexTool(agent, store);
        return agent;
    }

    private static String systemPrompt() {
        return """
                你是年报资料专员。你只有一个事实来源：固定年报快照（只读）。
                规则：
                1. 任何数值/表述必须来自 read_page 返回的原文片段；禁止凭记忆或推理补充数字。
                2. 每次返回证据时必须携带 digest、page、excerpt 三元组，缺一不可。
                3. 快照中没有的内容，明确回复 MATERIAL_MISSING 并说明查找过的页码；绝不编造。
                4. 不访问网络，不猜测页码内容；先 list_metrics_index 再 read_page。
                """;
    }

    /** read_page 工具：确定性读快照页（LLM 只决定读哪页，取数由代码完成）。 */
    private static void registerReadPageTool(ReActAgent agent, SnapshotStore store) {
        ToolCard toolCard = ToolCard.builder()
                .id("read_page")
                .name("read_page")
                .description("读取固定年报快照指定页原文，返回 {status,digest,page,excerpt}；缺页返回 material_missing")
                .inputParams(readPageSchema())
                .build();
        LocalFunction fn = new LocalFunction(toolCard, args -> {
            Object pageArg = args.get("page");
            int page = pageArg instanceof Number n ? n.intValue() : Integer.parseInt(String.valueOf(pageArg));
            SnapshotPage sp = store.page(page);
            if (sp == null) {
                return Map.of("status", "material_missing", "page", page,
                        "message", "快照中没有第 " + page + " 页；已检索范围 1-" + store.maxPage());
            }
            return Map.of("status", "ok", "digest", sp.digest(), "page", sp.page(), "excerpt", sp.text());
        });
        Runner.resourceMgr().addTool(fn, agent.getCard().getId());
        agent.getAbilityManager().add(toolCard);
    }

    private static Map<String, Object> readPageSchema() {
        Map<String, Object> props = new HashMap<>();
        props.put("page", Map.of("type", "integer", "description", "1 起始页码"));
        Map<String, Object> schema = new HashMap<>();
        schema.put("type", "object");
        schema.put("properties", props);
        schema.put("required", List.of("page"));
        return schema;
    }

    /** list_metrics_index 工具：列出快照 manifest 的指标索引页（W01 冻结的入口）。 */
    private static void registerListMetricsIndexTool(ReActAgent agent, SnapshotStore store) {
        ToolCard toolCard = ToolCard.builder()
                .id("list_metrics_index")
                .name("list_metrics_index")
                .description("列出固定快照的指标索引（指标名→页码），供 read_page 定位。")
                .inputParams(Map.of("type", "object", "properties", Map.of()))
                .build();
        LocalFunction fn = new LocalFunction(toolCard, args -> {
            Map<String, Object> idx = new LinkedHashMap<>();
            for (MetricKey k : MetricKey.values()) {
                Integer pg = store.indexPage(k);
                idx.put(k.name(), pg == null ? "NOT_INDEXED" : pg);
            }
            return Map.of("status", "ok", "index", idx);
        });
        Runner.resourceMgr().addTool(fn, agent.getCard().getId());
        agent.getAbilityManager().add(toolCard);
    }

    /** 快照仓库    /** 快照仓库：load 时机在 wrapper 启动（内存页索引，读多写零）。 */
    public static final class SnapshotStore {
        private final Path snapshotDir;
        private final Map<Integer, SnapshotPage> pages = new LinkedHashMap<>();
        private final Map<MetricKey, Integer> metricIndex = new LinkedHashMap<>();

        public SnapshotStore(Path snapshotDir) {
            this.snapshotDir = snapshotDir;
        }

        /** 由 wrapper 层在启动时调用一次；失败即启动失败（fail-fast，不静默空快照）。 */
        public synchronized void load() {
            try {
                ObjectMapper mapper = new ObjectMapper();
                var manifest = mapper.readTree(snapshotDir.resolve("manifest.json").toFile());
                var idx = manifest.get("metric_index");
                if (idx != null) {
                    idx.fields().forEachRemaining(e ->
                            metricIndex.put(MetricKey.valueOf(e.getKey()), e.getValue().asInt()));
                }
                Path pagesDir = snapshotDir.resolve("pages");
                try (var stream = Files.list(pagesDir)) {
                    stream.filter(p -> p.getFileName().toString().matches("page-\\d+\\.txt"))
                            .sorted()
                            .forEach(p -> {
                                try {
                                    String text = Files.readString(p, StandardCharsets.UTF_8);
                                    int page = Integer.parseInt(
                                            p.getFileName().toString().replaceAll("\\D", ""));
                                    pages.put(page, new SnapshotPage(sha256(text), page, text));
                                } catch (Exception e) {
                                    throw new IllegalStateException("快照页读取失败: " + p, e);
                                }
                            });
                }
                if (pages.isEmpty()) {
                    throw new IllegalStateException("快照目录无页面: " + snapshotDir);
                }
            } catch (Exception e) {
                throw new IllegalStateException("年报快照装载失败（fail-fast）: " + snapshotDir, e);
            }
        }

        public SnapshotPage page(int page) { return pages.get(page); }
        public Integer indexPage(MetricKey key) { return metricIndex.get(key); }
        public int maxPage() { return pages.keySet().stream().max(Integer::compare).orElse(0); }

        static String sha256(String s) {
            try {
                MessageDigest md = MessageDigest.getInstance("SHA-256");
                return HexFormat.of().formatHex(md.digest(s.getBytes(StandardCharsets.UTF_8)));
            } catch (Exception e) {
                throw new IllegalStateException(e);
            }
        }
    }
}
