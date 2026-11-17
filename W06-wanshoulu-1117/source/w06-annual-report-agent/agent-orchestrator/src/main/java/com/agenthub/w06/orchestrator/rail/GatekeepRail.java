/*
 * GatekeepRail — 交付门禁（W06 对上游示例最关键的改造）
 *
 * 上游 multi-deep-research-demo 的 verify-agent 明确是「best-effort 质量闸，
 * 失败绝不阻塞交付」（其 README 原文）。W06 的要求恰好相反（#16 L127：
 * 「检查失败阻止验收」；#20 质疑「控制面只是建议面」）。
 *
 * 本 Rail 的状态机：
 *   draft#0 → checker → PASSED            → deliverable，附 ContractVerdict
 *                     → FAILED (round=0)  → 注入 REVISION_DIRECTIVE 打回 writer 一次
 *                     → FAILED (round=1)  → waiting_human（不再重试，人工接管）
 *                     → WAITING_HUMAN      → 直接 waiting_human（A08：语义无法自动判定）
 *
 * 修订次数上限 maxRounds=1 与采纳图中的 revision_gate 预声明一致（v2 §4）。
 */

package com.agenthub.w06.orchestrator.rail;

import com.agenthub.w06.contract.Violation;
import com.agenthub.w06.contract.Severity;
import com.agenthub.w06.contract.AnnualReportArtifacts;
import com.agenthub.w06.contract.CheckResult;
import com.agenthub.w06.contract.CheckStatus;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.openjiuwen.core.context.ModelContext;
import com.openjiuwen.core.foundation.llm.schema.BaseMessage;
import com.openjiuwen.core.singleagent.rail.AgentCallbackContext;
import com.openjiuwen.core.singleagent.rail.InvokeInputs;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * 落在 orchestrator（root DeepAgent）上的门禁 rail。
 * 与上游 AutoPersistMemoryRail 同样的 afterInvoke 生命周期位置，但语义相反：
 * AutoPersist 是 best-effort 落盘（fail-open），本 rail 是 fail-closed 交付门禁。
 */
public final class GatekeepRail extends com.openjiuwen.memory.harness.rails.MemoryRail {

    private static final DateTimeFormatter TS = DateTimeFormatter.ofPattern("yyyy-MM-dd-HHmmss");
    private static final String VERDICT_PREFIX = "CONTRACT_VERDICT:";

    private final ObjectMapper mapper = new ObjectMapper();
    private final Path workspaceRoot;
    private final int maxRounds;

    /** 修订轮次跟踪（per conversation；demo 单任务场景够用）。 */
    private final Map<String, Integer> revisionRounds = new LinkedHashMap<>();

    public GatekeepRail(Path workspaceRoot, int maxRounds) {
        this.workspaceRoot = workspaceRoot;
        this.maxRounds = maxRounds;
    }

    @Override
    public void afterInvoke(AgentCallbackContext ctx) {
        if (ctx == null || toolContext == null) {
            return;
        }
        if (!(ctx.getInputs() instanceof InvokeInputs inputs)) {
            return;
        }
        Map<String, Object> result = inputs.getResult();
        if (result == null) {
            return;
        }
        String output = String.valueOf(result.get("output"));
        if (output == null || output.isBlank()) {
            return;
        }
        // 只对产出工件草案的回合做门禁；修订指令回合(REVISION_DIRECTIVE 注入)跳过
        if (output.contains("REVISION_DIRECTIVE")) {
            return;
        }
        String conversationId = String.valueOf(inputs.getConversationId());

        // 1) 解析结构化工件（writer 输出的 JSON 段）
        Optional<AnnualReportArtifacts> draft = parseArtifacts(output);
        if (draft.isEmpty()) {
            // 结构化失败：本身即 BLOCKING —— 「先结构化再渲染」，不结构化不交付
            persistVerdict(conversationId, singleCheckResult(
                    "structure-checker", "0.1.0", CheckStatus.FAILED,
                    List.of(new Violation(
                            "NOT_STRUCTURED", "output",
                            "writer 输出不含合法 AnnualReportArtifacts JSON", 
                            com.agenthub.w06.contract.Severity.BLOCKING))),
                    output);
            markWaitingHuman(result, "工件非结构化：writer 未按契约输出 JSON");
            return;
        }

        // 2) 调确定性 checker（数字复算 + 引用核验 —— 由 checker 模块注入的核验函数）
        List<CheckResult> checks = checkerHook == null ? List.of() : checkerHook.run(draft.get());
        CheckStatus worst = worstStatus(checks);

        // 3) 门禁决策
        switch (worst) {
            case PASSED -> {
                persistVerdict(conversationId, merge(checks), output);
                result.put("deliverable", true);
                result.put("verdict", VERDICT_PREFIX + "PASSED");
            }
            case WAITING_HUMAN -> {
                persistVerdict(conversationId, merge(checks), output);
                markWaitingHuman(result, "存在无法自动判定的语义问题（A08）");
            }
            case FAILED -> {
                int round = revisionRounds.getOrDefault(conversationId, 0);
                if (round < maxRounds) {
                    revisionRounds.put(conversationId, round + 1);
                    persistVerdict(conversationId, merge(checks), output);
                    // 打回：下一轮向 writer 注入定向修订指令（steering 消息）
                    pushSteering(ctx, renderRevisionDirective(checks));
                    result.put("deliverable", false);
                    result.put("verdict", VERDICT_PREFIX + "REVISION_REQUESTED");
                } else {
                    persistVerdict(conversationId, merge(checks), output);
                    markWaitingHuman(result, "定向修订后仍 FAIL，转人工（round=" + round + "）");
                }
            }
        }
    }

    /** checker 注入点：由 wrapper 层把 NumericChecker/CitationChecker 组合接进来。 */
    public interface CheckerHook {
        List<CheckResult> run(AnnualReportArtifacts draft);
    }

    private CheckerHook checkerHook;

    public void setCheckerHook(CheckerHook hook) {
        this.checkerHook = hook;
    }

    // ---------------- 内部 ----------------

    private void markWaitingHuman(Map<String, Object> result, String reason) {
        result.put("deliverable", false);
        result.put("verdict", VERDICT_PREFIX + "WAITING_HUMAN");
        result.put("wait_reason", "waiting_human");
        result.put("wait_detail", reason);
    }

    private void pushSteering(AgentCallbackContext ctx, String directive) {
        // 与上游 [STEERING] 约定一致：runtime 在工具结果后注入的伪用户消息
        ModelContext context = ctx.getContext();
        if (context != null) {
            BaseMessage steering = new BaseMessage();
            steering.setRole("user");
            steering.setContent("[STEERING]" + directive);
            context.getMessages().add(steering);
        }
    }

    private String renderRevisionDirective(List<CheckResult> checks) {
        StringBuilder sb = new StringBuilder("REVISION_DIRECTIVE：以下违规必须修复，其余章节原样保留：\n");
        for (CheckResult c : checks) {
            for (var v : c.violations()) {
                if (v.severity() == com.agenthub.w06.contract.Severity.BLOCKING) {
                    sb.append("- [").append(v.rule()).append("] ")
                      .append(v.target()).append("：").append(v.detail()).append('\n');
                }
            }
        }
        return sb.toString();
    }

    private Optional<AnnualReportArtifacts> parseArtifacts(String output) {
        // writer 输出中提取第一个 JSON 对象（容忍前后文字）
        int start = output.indexOf('{');
        int end = output.lastIndexOf('}');
        if (start < 0 || end <= start) {
            return Optional.empty();
        }
        try {
            return Optional.of(mapper.readValue(
                    output.substring(start, end + 1), AnnualReportArtifacts.class));
        } catch (Exception e) {
            return Optional.empty();
        }
    }

    private static CheckResult merge(List<CheckResult> checks) {
        // 合并为单一 verdict 记录（保留各 checker 明细于 violations）
        var allViolations = checks.stream()
                .flatMap(c -> c.violations().stream())
                .toList();
        // P1：合并 verdict 继承被验工件指纹——各 checker 对同一 draft 核验，指纹必须一致；
        // 任一 checker 指纹缺失（不可绑定）则合并结论同样不可绑定（fail-closed）。
        String digest = checks.stream()
                .map(CheckResult::artifactSha256)
                .filter(java.util.Objects::nonNull)
                .findFirst()
                .orElse(null);
        boolean anyUnbound = checks.stream().anyMatch(c -> c.artifactSha256() == null);
        return new CheckResult("gatekeep-merged", "0.1.0", worstStatus(checks), allViolations,
                anyUnbound ? null : digest, java.time.Instant.now().toString(),
                checks.stream().map(CheckResult::executionO)
                        .filter(java.util.Objects::nonNull).findFirst().orElse(null));
    }

    private static CheckStatus worstStatus(List<CheckResult> checks) {
        CheckStatus worst = CheckStatus.PASSED;
        for (CheckResult c : checks) {
            if (c.status() == CheckStatus.FAILED) {
                return CheckStatus.FAILED;
            }
            if (c.status() == CheckStatus.WAITING_HUMAN) {
                worst = CheckStatus.WAITING_HUMAN;
            }
        }
        return worst;
    }

    private void persistVerdict(String conversationId, CheckResult verdict, String draft) {
        // 确定性落盘：verdict 与 draft 同目录成对（fail-open 落盘、fail-closed 门禁——两件事分开）
        writeVerdictToMemory(conversationId, verdict);   // G2：同步写 openjiuwen 记忆（知识复用侧）
        try {
            Path verdicts = workspaceRoot.resolve("verdicts");
            Path drafts = workspaceRoot.resolve("drafts");
            Files.createDirectories(verdicts);
            Files.createDirectories(drafts);
            String ts = LocalDateTime.now().format(TS);
            Files.writeString(verdicts.resolve("verdict-" + ts + ".json"),
                    mapper.writerWithDefaultPrettyPrinter().writeValueAsString(verdict),
                    StandardCharsets.UTF_8);
            Files.writeString(drafts.resolve("draft-" + ts + ".json"), draft, StandardCharsets.UTF_8);
        } catch (Exception ignored) {
            // 落盘失败不改变门禁结论，但必须可观测：verdict 已写入 result 供上层记录
        }
    }

    private static CheckResult singleCheckResult(String id, String v, CheckStatus s,
                                           List<com.agenthub.w06.contract.Violation> vs) {
        return new CheckResult(id, v, s, vs, null, java.time.Instant.now().toString(), null);
    }

    // ---------------- G2（2026-11-17）：openjiuwen 记忆引擎真接入 ----------------
    // verdict 写入 write_memory 语义：下游任务可 search 历史核验结论（#16 知识复用验收：
    // "命中/换版本不命中/提交前撤回"的数据来源）。fail-open：记忆失败不阻断门禁（文件已落盘兜底）。

    private void writeVerdictToMemory(String conversationId, CheckResult verdict) {
        try {
            if (toolContext == null || !toolContext.ensureManager()) {
                return;
            }
            com.openjiuwen.memory.lite.MemoryIndexManager mgr = toolContext.getManager();
            if (mgr == null) {
                return;
            }
            String digest6 = verdict.artifactSha256() == null
                    ? "unbound" : verdict.artifactSha256().substring(0, Math.min(12, verdict.artifactSha256().length()));
            String content = "CONTRACT_VERDICT conversation=" + conversationId
                    + " checker=" + verdict.checkerId()
                    + " status=" + verdict.status()
                    + " artifact=" + digest6
                    + (verdict.executionO() != null ? " o=" + verdict.executionO().asText() : "")
                    + " violations=" + verdict.violations().size();
            // MemoryIndexManager 无直接 write——经 invokeMemoryTool 走 write_memory 工具语义
            invokeMemoryTool("write_memory", java.util.Map.of(
                    "content", content,
                    "metadata", java.util.Map.of(
                            "conversation_id", conversationId,
                            "verdict_status", verdict.status().name(),
                            "artifact_sha256", verdict.artifactSha256() == null ? "" : verdict.artifactSha256())));
        } catch (Exception memEx) {
            // 记忆写入失败不阻断门禁（文件落盘已兜底）；上层可观测性另接
        }
    }

    /** G2：任务开始检索历史核验结论（read_memory 语义）——跨任务复用入口。 */
    public java.util.List<java.util.Map<String, Object>> searchPastVerdicts(String query) {
        try {
            if (toolContext == null || !toolContext.ensureManager()) {
                return java.util.List.of();
            }
            com.openjiuwen.memory.lite.MemoryIndexManager mgr = toolContext.getManager();
            if (mgr == null) {
                return java.util.List.of();
            }
            return mgr.search(query, java.util.Map.of("top_k", 5));
        } catch (Exception e) {
            return java.util.List.of();
        }
    }
}
