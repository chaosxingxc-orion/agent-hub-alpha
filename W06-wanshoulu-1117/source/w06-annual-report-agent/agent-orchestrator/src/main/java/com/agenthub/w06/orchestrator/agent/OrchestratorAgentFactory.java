/*
 * OrchestratorAgentFactory — W06 root 编排 agent（库层，无 Spring）
 *
 * 照 multi-deep-research-demo 的 DeepResearchAgentFactory 模式：
 *  - root DeepAgent + Rails；五个角色子 agent 与 critic 经 A2A remote-agents 注入
 *    （wrapper 层 application.yml 的 openjiuwen.service.a2a.remote-agents）；
 *  - 与上游的差异：GatekeepRail（fail-closed 交付门禁）替换 best-effort verify。
 */

package com.agenthub.w06.orchestrator.agent;

import com.agenthub.w06.contract.AnnualReportArtifacts;

import com.agenthub.w06.checker.CitationChecker;
import com.agenthub.w06.checker.NumericChecker;
import com.agenthub.w06.orchestrator.rail.GatekeepRail;
import com.openjiuwen.harness.deep_agent.DeepAgent;
import com.openjiuwen.harness.factory.HarnessFactory;
import com.openjiuwen.harness.schema.config.DeepAgentConfig;
import com.openjiuwen.harness.workspace.Workspace;
import com.openjiuwen.core.singleagent.schema.AgentCard;

import java.nio.file.Path;
import java.util.List;

public final class OrchestratorAgentFactory {

    public static DeepAgent build(OrchestratorProperties props,
                                  NumericChecker.SnapshotTextResolver resolver,
                                  Path workspaceRoot) {
        props.requireConfigured();

        // 门禁 rail：checker hook = 确定性数值核验 + 引用核验（语义层由 critic 子 agent 补）
        GatekeepRail gatekeep = new GatekeepRail(workspaceRoot, props.getGatekeepMaxRounds());
        gatekeep.setCheckerHook(draft -> List.of(
                new NumericChecker().check(draft, resolver),
                new CitationChecker().check(draft, resolver)));

        // G4（2026-11-17）：主编澄清机制化——AskUserRail 进 rails 列表（上游 interrupt 标准）。
        // 对应 orchestrator-skill 失败三态之"未配置规约→进入澄清"：任务要素不明时中断等待用户。
        com.openjiuwen.harness.rails.interrupt.AskUserRail askUser = new com.openjiuwen.harness.rails.interrupt.AskUserRail();

        DeepAgentConfig config = DeepAgentConfig.builder()
                .systemPrompt(systemPrompt())
                .maxIterations(props.getMaxIterations())
                .enableTaskLoop(true)
                .enableTaskPlanning(true)
                .completionTimeout((double) props.getCompletionTimeout().toSeconds())
                .workspacePath(props.getWorkspacePath())
                .model(props.modelConfig())
                .backend(props.backendConfig())
                .rails(List.of(gatekeep, askUser))
                .build();
        Workspace workspace = Workspace.builder()
                .rootPath(props.getWorkspacePath())
                .language("zh-CN")
                .build();
        AgentCard card = AgentCard.builder()
                .id("w06-orchestrator")
                .name("AnnualReportOrchestrator")
                .description("W06 年报报告编排：资料→指标→双分析并行→写作→确定性+语义双门禁→交付/修订/人工")
                .build();
        com.openjiuwen.harness.deep_agent.DeepAgent agent =
                HarnessFactory.createDeepAgent(card, config, workspace);
        // G4：ask_user 工具挂 harness tool 面（AskUserRail 拦截，resume 以用户输入回注）
        mountAskUserTool(agent);
        return agent;
    }

    /** G4：ask_user 工具（AskUserRail 拦截——body 不执行，resume 时以用户输入作结果回注）。 */
    private static void mountAskUserTool(com.openjiuwen.harness.deep_agent.DeepAgent agent) {
        com.openjiuwen.core.foundation.tool.ToolCard askCard = com.openjiuwen.core.foundation.tool.ToolCard.builder()
                .id("ask_user")
                .name("ask_user")
                .description("任务要素不明（银行/期间/口径/输出范围）时向用户提问。不要自行猜测。")
                .inputParams(java.util.Map.of("type", "object",
                        "properties", java.util.Map.of("question", java.util.Map.of("type", "string")),
                        "required", java.util.List.of("question")))
                .build();
        com.openjiuwen.core.foundation.tool.function.LocalFunction askFn =
                new com.openjiuwen.core.foundation.tool.function.LocalFunction(askCard,
                        args -> "（AskUserRail 拦截：中断等待用户输入）");
        com.openjiuwen.core.runner.Runner.resourceMgr().addTool(askFn, agent.getCard().getId());
        agent.getAbilityManager().add(askCard);
    }

    private static String systemPrompt() {
        return """
                你是年报报告编排专员（root）。按以下次序推进，禁止跳步：
                1. 资料：调 material-agent 定位八指标所在页并取原文片段；
                2. 指标：调 metrics-agent 抽取八指标结构化值（含 digest/page）；
                3. 分析【并行】：同时调 operation-agent 与 asset-quality-agent，
                   两者互不感知，各自产出带引用草稿；
                4. 写作：把两路草稿+指标交给 writer-agent，产出 AnnualReportArtifacts JSON；
                5. 门禁：等待系统核验结果（CONTRACT_VERDICT）。PASSED 才可宣布完成；
                   REVISION_REQUESTED 时按修订指令重调 writer 一次；WAITING_HUMAN 时
                   如实报告等待原因，禁止自行放行。
                铁律：报告里任何数字必须来自步骤 2 的指标 artifact；你不得口述数字。
                """;
    }
}
