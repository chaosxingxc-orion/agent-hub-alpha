/*
 * AssetQualityAgentFactory — 资产质量分析角色（W06）
 * 与 OperationAgentFactory 同构；差异仅在允许指标集合（隔离边界的另一半）。
 */

package com.agenthub.w06.agents.assetquality;

import com.agenthub.w06.contract.MetricKey;
import com.agenthub.w06.agents.operation.OperationAgentFactory;
import com.agenthub.w06.agents.operation.AnalysisAgentProperties;
import com.openjiuwen.core.singleagent.agents.ReActAgent;
import com.openjiuwen.core.singleagent.agents.ReActAgentConfig;
import com.openjiuwen.core.singleagent.schema.AgentCard;

import java.util.List;
import java.util.Map;
import java.util.Set;

public final class AssetQualityAgentFactory {

    /** 资产质量分支允许消费的指标（与经营分支互斥，交集为空=硬隔离）。 */
    public static final Set<MetricKey> ALLOWED_METRICS = Set.of(
            MetricKey.NPL_BALANCE,
            MetricKey.NPL_RATIO,
            MetricKey.PROVISION_COVERAGE,
            MetricKey.CET1_RATIO);

    public static ReActAgent build(AnalysisAgentProperties props) {
        props.requireConfigured();
        ReActAgentConfig config = ReActAgentConfig.builder()
                .maxIterations(props.getMaxIterations())
                .sysOperationId(props.getSysOperationId())
                .promptTemplate(List.of(Map.of("role", "system", "content",
                        OperationAgentFactory.analysisPrompt("资产质量分析", ALLOWED_METRICS))))
                .build()
                .configureModelClient(
                        props.getProvider(), props.getApiKey(), props.getApiBase(),
                        props.getModelName(), props.isSslVerify());
        AgentCard card = AgentCard.builder()
                .id(props.getAgentId())
                .name("AssetQualityAnalysisAgent")
                .description("资产质量分析角色：基于资产质量类指标产出带引用的分行分析草稿")
                .build();
        ReActAgent agent = new ReActAgent(card);
        agent.configure(config);

        // ===== G1-a/G4（2026-11-17）：loop+rail 规范升级（照上游 search-agent 模式）=====
        // 纪律 rail：歧义/缺料时模型可挂 ask_user 中断等待人工（AskUserRail 短路：首调中断、
        // resume 后把用户答复作为工具结果回注，模型不得自行放行）
        mountComputeTool(agent);
        mountAskUserTool(agent);
        agent.registerRail(new com.openjiuwen.harness.rails.interrupt.AskUserRail());
        // Skill 渐进披露（G1-b）：registerSkill 挂技能包（lazyInitSkill 后生效），SkillUseRail
        // 按 skill 目录加载——节点内知识层与线上注册表的 bind 语义对齐
        try {
            String skillDir = System.getenv().getOrDefault("SKILLS_DIR", "/skills");
            java.nio.file.Path p = java.nio.file.Path.of(skillDir, "annual-report");
            if (java.nio.file.Files.isDirectory(p)) {
                agent.registerSkill(java.util.List.of(p.toString()));
                agent.registerRail(new com.openjiuwen.harness.rails.skills.SkillUseRail(
                        java.util.List.of(p.toString()), "auto-list"));
            }
        } catch (Exception skillEx) {
            // skill 目录缺失不阻断角色启动（无 skill 时 prompt 已含纪律）
        }
        return agent;
    }

    /** ask_user 工具注册（AskUserRail 拦截；body 不执行——resume 时 rail 以用户输入作结果回注）。 */
    private static void mountAskUserTool(ReActAgent agent) {
        com.openjiuwen.core.foundation.tool.ToolCard askCard = com.openjiuwen.core.foundation.tool.ToolCard.builder()
                .id("ask_user")
                .name("ask_user")
                .description("当输入歧义、口径不清或缺材料时，向用户提问并等待答复。不要自行猜测。")
                .build();
        com.openjiuwen.core.foundation.tool.function.LocalFunction askFn =
                new com.openjiuwen.core.foundation.tool.function.LocalFunction(askCard, args -> "（AskUserRail 拦截：中断等待用户输入）");
        com.openjiuwen.core.runner.Runner.resourceMgr().addTool(askFn, agent.getCard().getId());
        agent.getAbilityManager().add(askCard);
    }


    /** G1-b：w06-compute 派生计算工具挂载（MCP 能力→LocalFunction，上游 ToolCard 模式）。
     *  分析角色的"不心算"从纪律变机制：模型派生值必须走本工具，算式与结果回显进 loop 供核验回溯。 */
    private static void mountComputeTool(ReActAgent agent) {
        // eval_expression：安全算式求值
        com.openjiuwen.core.foundation.tool.ToolCard evalCard = com.openjiuwen.core.foundation.tool.ToolCard.builder()
                .id("eval_expression")
                .name("eval_expression")
                .description("安全算式求值（仅数字与 + - * / ( ) . %）：派生值计算一律经本工具，禁止心算。返回 算式=结果。")
                .inputParams(evalExprSchema())
                .build();
        com.openjiuwen.core.foundation.tool.function.LocalFunction evalFn =
                new com.openjiuwen.core.foundation.tool.function.LocalFunction(evalCard,
                        args -> safeEval(String.valueOf(args.get("expression"))));
        com.openjiuwen.core.runner.Runner.resourceMgr().addTool(evalFn, agent.getCard().getId());
        agent.getAbilityManager().add(evalCard);

        // yoy_calc：同比计算
        com.openjiuwen.core.foundation.tool.ToolCard yoyCard = com.openjiuwen.core.foundation.tool.ToolCard.builder()
                .id("yoy_calc")
                .name("yoy_calc")
                .description("同比计算 (cur-prior)/prior*100：返回算式与百分比结果。prior 为基期值。")
                .inputParams(yoySchema())
                .build();
        com.openjiuwen.core.foundation.tool.function.LocalFunction yoyFn =
                new com.openjiuwen.core.foundation.tool.function.LocalFunction(yoyCard, args -> {
                    try {
                        double cur = ((Number) args.get("current")).doubleValue();
                        double prior = ((Number) args.get("prior")).doubleValue();
                        if (prior == 0) return "error: prior=0（基期为零无法计算同比）";
                        return String.format("(%.4f-%.4f)/%.4f*100 = %.2f%%", cur, prior, prior, (cur - prior) / prior * 100);
                    } catch (Exception e) {
                        return "error: " + e.getMessage();
                    }
                });
        com.openjiuwen.core.runner.Runner.resourceMgr().addTool(yoyFn, agent.getCard().getId());
        agent.getAbilityManager().add(yoyCard);
    }

    /** 递归下降四则求值（与 w06-compute MCP 的 safeEval 同源；三形态同源纪律）。 */
    static String safeEval(String expr) {
        String cleaned = expr.replaceAll("[\\s]", "");
        if (!cleaned.matches("[0-9+*/().%-]+")) {
            return "error: expression contains illegal characters";
        }
        try {
            double v = new Object() {
                int i = 0; String s = cleaned;
                double parse() { double r = parseAdd(); if (i < s.length()) throw new ArithmeticException("syntax"); return r; }
                double parseAdd() { double r = parseMul(); while (i < s.length() && (s.charAt(i)=='+'||s.charAt(i)=='-')) { char op=s.charAt(i++); double b=parseMul(); r = op=='+'?r+b:r-b; } return r; }
                double parseMul() { double r = parseAtom(); while (i < s.length() && (s.charAt(i)=='*'||s.charAt(i)=='/')) { char op=s.charAt(i++); double b=parseAtom(); if (op=='/' && b==0) throw new ArithmeticException("div0"); r = op=='*'?r*b:r/b; } return r; }
                double parseAtom() { if (i<s.length() && s.charAt(i)=='(') { i++; double r=parseAdd(); if (i>=s.length()||s.charAt(i)!=')') throw new ArithmeticException("paren"); i++; return r; }
                    int st=i; if (i<s.length()&&(s.charAt(i)=='+'||s.charAt(i)=='-')) i++;
                    while (i<s.length() && (Character.isDigit(s.charAt(i))||s.charAt(i)=='.')) i++;
                    if (st==i) throw new ArithmeticException("number");
                    return Double.parseDouble(s.substring(st,i)); }
            }.parse();
            return String.format("%s = %.4f", expr, v);
        } catch (Exception e) {
            return "error: " + e.getMessage();
        }
    }

    /** eval_expression 的 inputSchema（与 material readPageSchema 同构）。 */
    private static java.util.Map<String, Object> evalExprSchema() {
        java.util.Map<String, Object> props = new java.util.HashMap<>();
        props.put("expression", java.util.Map.of("type", "string", "description", "算式：仅数字与 + - * / ( ) . %"));
        java.util.Map<String, Object> schema = new java.util.HashMap<>();
        schema.put("type", "object");
        schema.put("properties", props);
        schema.put("required", java.util.List.of("expression"));
        return schema;
    }

    /** yoy_calc 的 inputSchema。 */
    private static java.util.Map<String, Object> yoySchema() {
        java.util.Map<String, Object> props = new java.util.HashMap<>();
        props.put("current", java.util.Map.of("type", "number", "description", "本期值"));
        props.put("prior", java.util.Map.of("type", "number", "description", "基期值"));
        java.util.Map<String, Object> schema = new java.util.HashMap<>();
        schema.put("type", "object");
        schema.put("properties", props);
        schema.put("required", java.util.List.of("current", "prior"));
        return schema;
    }
}
