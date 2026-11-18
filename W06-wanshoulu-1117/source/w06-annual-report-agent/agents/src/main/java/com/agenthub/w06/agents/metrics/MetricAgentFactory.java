/*
 * MetricAgent — 指标抽取角色（W06）
 *
 * 数字纪律的第一道关口：八指标的结构化抽取只能由本 agent 从 MaterialAgent
 * 返回的原文片段完成，输出必须是 MetricValue JSON —— value/unit/period/digest/page
 * 五元组齐全。下游 NumericChecker 会用确定性规则复算（如 净息差=利息净收入/生息资产
 * 的口径核对、单位换算），writer 口述数字与本 artifact 不一致一律 FAIL。
 */

package com.agenthub.w06.agents.metrics;

import com.agenthub.w06.contract.MetricKey;
import com.openjiuwen.core.singleagent.agents.ReActAgent;
import com.openjiuwen.core.singleagent.agents.ReActAgentConfig;
import com.openjiuwen.core.singleagent.schema.AgentCard;

import java.util.List;
import java.util.Map;

public final class MetricAgentFactory {

    public static ReActAgent build(MetricAgentProperties props) {
        props.requireConfigured();
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
                .name("MetricAgent")
                .description("指标角色：从资料 agent 的原文片段抽取八指标结构化值，输出 MetricValue JSON 数组")
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


    private static String systemPrompt() {
        // 八指标枚举在 prompt 里显式列出，避免模型自由发挥指标集
        StringBuilder keys = new StringBuilder();
        for (MetricKey k : MetricKey.values()) {
            keys.append("- ").append(k.name()).append("（").append(k.getDisplayName()).append("）\n");
        }
        return """
                你是银行年报指标抽取专员。输入是资料 agent 返回的原文片段（带 digest/page）。
                
                固定八项指标（不得增删）：
                %s
                输出契约（严格 JSON，无其他文字）：
                [{"key":"<枚举名>","value":"<原文口径字符串>","unit":"<原文单位>","period":"<期间>",
                  "source_document_digest":"<输入片段的 digest>","page":<页码>,
                  "evidence_excerpt_digest":"<引用片段的 sha256，由输入给出>"}]
                  
                铁律：
                1. value 只能照抄原文数字（含千分位），禁止换算、四舍五入或推算；
                2. unit 用原文单位（亿元/% 等）；原文未写单位则输出 "UNSPECIFIED"，禁止猜；
                3. 找不到的指标输出该条并标 value="NOT_FOUND"，禁止编造数字；
                4. 每条必须携带输入片段的 digest 与 page；无来源的指标禁止输出。
                """.formatted(keys);
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
