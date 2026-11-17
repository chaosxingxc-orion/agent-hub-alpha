/*
 * CriticAgent — 语义支持度判定（引用核验 L2 层）
 *
 * finscope Critic 模式的 Java 移植：逐条核对「论断 ↔ 证据摘录」的支持度。
 * 关键约束（与上游 verify-agent 的差异）：
 *  - 输出必须是机器可判的 VERDICT 行（PASSED/FAILED/WAITING_HUMAN + 逐条明细），
 *    由确定性规则解析 verdict（RuleBasedCriteriaVerifier 的 substring 约定，
 *    上游注释里论证过这个粗但可靠的代理）；
 *  - 无法确认 → WAITING_HUMAN，绝不放行（#16：不由 LLM judge 单独批准）；
 *  - temperature=0（判定任务）。
 */

package com.agenthub.w06.checker;

import com.openjiuwen.agents.reactrails.replan.ReplanRail;
import com.openjiuwen.agents.reactrails.selfheal.RootCauseRail;
import com.openjiuwen.agents.reactrails.verification.CriteriaReplanBridgeRail;
import com.openjiuwen.agents.reactrails.verification.RuleBasedCriteriaVerifier;
import com.openjiuwen.core.singleagent.agents.ReActAgent;
import com.openjiuwen.core.singleagent.agents.ReActAgentConfig;
import com.openjiuwen.core.singleagent.schema.AgentCard;
import com.openjiuwen.agents.reactrails.observability.ReactRailsObservability;

import java.util.List;
import java.util.Map;

public final class CriticAgentFactory {

    public static ReActAgent build(CriticAgentProperties props) {
        props.requireConfigured();
        ReActAgentConfig config = ReActAgentConfig.builder()
                .maxIterations(props.getMaxIterations())
                .sysOperationId(props.getSysOperationId())
                .promptTemplate(List.of(Map.of("role", "system", "content", systemPrompt())))
                .build()
                .configureModelClient(
                        props.getProvider(), props.getApiKey(), props.getApiBase(),
                        props.getModelName(), props.isSslVerify());
        if (config.getModelConfigObj() != null) {
            config.getModelConfigObj().setTemperature(0.0); // 判定任务
        }
        AgentCard card = AgentCard.builder()
                .id(props.getAgentId())
                .name("CriticAgent")
                .description("语义支持度判定：逐条核对论断是否被证据摘录支撑；无法确认输出 WAITING_HUMAN")
                .build();
        ReActAgent agent = new ReActAgent(card);
        agent.configure(config);

        // 与上游 verify-agent 同款 rails：criteria 逐项 substring 判定
        ReplanRail replan = new ReplanRail(props.getMaxReplan());
        agent.registerRail(new CriteriaReplanBridgeRail(
                new RuleBasedCriteriaVerifier(),
                props.getCriteria(),
                replan));
        agent.registerRail(new RootCauseRail());
        ReactRailsObservability.install(agent);
        return agent;
    }

    private static String systemPrompt() {
        return """
                你是年报引用核验员。输入：论断列表（每条含 claim_id、claim_text、证据摘录）。
                逐条判定证据摘录是否在语义上支撑论断（数字、期间、口径、主体都要对上）。
                
                输出契约（逐行，无其他文字）：
                [C-1] SUPPORTED|UNSUPPORTED|UNVERIFIABLE <一句话理由>
                ...
                SEMANTIC_VERDICT: PASSED|FAILED|WAITING_HUMAN
                
                判定标准：
                - SUPPORTED：摘录明确含论断中的数字/事实，且口径（期间/单位/主体）一致；
                - UNSUPPORTED：数字对不上、口径错位、张冠李戴 —— 输出它；
                - UNVERIFIABLE：摘录与论断相关但不足以确认（如论断含推断）——输出它，
                  绝不放行为 SUPPORTED；全部 UNVERIFIABLE → SEMANTIC_VERDICT: WAITING_HUMAN。
                - 任何 UNSUPPORTED → SEMANTIC_VERDICT: FAILED。
                """;
    }
}
