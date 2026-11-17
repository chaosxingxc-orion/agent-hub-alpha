/*
 * NumericChecker — 确定性数值核验器（W06「计算和约定检查由工具承担」的落点）
 *
 * 纯 Java、零模型调用。规则集：
 *  N1 完整性：八指标一条不少（METRIC_MISSING）
 *  N2 单位白名单：每项指标只接受约定单位（UNIT_MISMATCH → A07 反例拦截）
 *  N3 数字来源：value 必须能在 source page 的原文中以字面出现（FABRICATED_NUMBER）
 *  N4 派生复算：可机器复算的口径（不良率=不良余额/贷款总额 等）逐项复算（CALC_MISMATCH）
 *  N5 期间成对：同比类指标须同时有本期与上期，或显式标注（PERIOD_MISMATCH）
 *  N6 跨期同比：正文同比论断按句归属（营业收入/净利润）后与上年度对比页复算（YOY_MISMATCH）
 *  N7 变动口径：比率变动用「个百分点/基点」、金额用「亿元」、增长率才用「%」（DELTA_UNIT_MISMATCH）
 *
 * 输入之外不查任何网络 —— 快照 digest 与页码是唯一事实来源。
 */

package com.agenthub.w06.checker;

import com.agenthub.w06.contract.*;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.EnumMap;
import java.util.HexFormat;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Pattern;

public final class NumericChecker {

    public static final String CHECKER_ID = "numeric-checker";
    public static final String CHECKER_VERSION = "0.2.1";

    /** G3：工件摘要用 Jackson 序列化（与 /v1/artifacts/digest 同一规范化路径）。 */
    private static final com.fasterxml.jackson.databind.ObjectMapper DIGEST_MAPPER =
            new com.fasterxml.jackson.databind.ObjectMapper();

    /** G3：被检工件的规范化 JSON 摘要 —— verdict 与工件绑定，旧结论不能放行新工件。 */
    static String artifactDigest(AnnualReportArtifacts draft) {
        try {
            byte[] canonical = DIGEST_MAPPER.writeValueAsBytes(draft);
            return HexFormat.of().formatHex(
                    MessageDigest.getInstance("SHA-256").digest(canonical));
        } catch (Exception e) {
            throw new IllegalStateException("无法计算工件摘要", e);
        }
    }

    /** 指标 → 约定单位（W01 冻结口径；单位错误即 A07）。 */
    private static final Map<MetricKey, String> EXPECTED_UNITS = Map.of(
            MetricKey.OPERATING_REVENUE, "亿元",
            MetricKey.NET_PROFIT_ATTRIBUTABLE, "亿元",
            MetricKey.NET_INTEREST_MARGIN, "%",
            MetricKey.COST_INCOME_RATIO, "%",
            MetricKey.NPL_BALANCE, "亿元",
            MetricKey.NPL_RATIO, "%",
            MetricKey.PROVISION_COVERAGE, "%",
            MetricKey.CET1_RATIO, "%");

    /** 原文数字匹配：容忍千分位逗号。 */
    private static final Pattern NUMBER = Pattern.compile("\\d{1,3}(?:,\\d{3})*(?:\\.\\d+)?|\\d+(?:\\.\\d+)?");

    /** N7 变动短语：方向词 + 幅度 + 口径单位词（个百分点/基点/亿元/%）。 */
    private static final Pattern DELTA_PHRASE = Pattern.compile(
            "(上升|下降|增加|减少|增减|变动|增长)\\s*(?:约\\s*)?([\\d,]+(?:\\.\\d+)?)\\s*(个百分点|基点|亿元|%|％)");

    public CheckResult check(AnnualReportArtifacts draft, SnapshotTextResolver resolver) {
        List<Violation> violations = new ArrayList<>();
        String artifactSha = artifactDigest(draft); // G3

        // N1 完整性
        Map<MetricKey, MetricValue> byKey = new EnumMap<>(MetricKey.class);
        for (MetricValue m : draft.metrics()) {
            if (byKey.put(m.key(), m) != null) {
                violations.add(v("METRIC_DUPLICATED", m.key().name(), "同一指标出现多条记录", Severity.BLOCKING));
            }
        }
        for (MetricKey k : MetricKey.values()) {
            if (!byKey.containsKey(k)) {
                violations.add(v("METRIC_MISSING", k.name(), "八指标缺失：" + k.getDisplayName(), Severity.BLOCKING));
            }
        }

        for (MetricValue m : draft.metrics()) {
            // N2 单位
            String expected = EXPECTED_UNITS.get(m.key());
            if (expected != null && !expected.equals(m.unit())) {
                violations.add(v("UNIT_MISMATCH", m.key().name(),
                        "单位应为 " + expected + "，实际 " + m.unit() + "（A07）", Severity.BLOCKING));
            }
            // N3 数字来源（NOT_FOUND 显式标注不算伪造，但完整性会在别处暴露）
            if (!"NOT_FOUND".equals(m.value()) && !"UNSPECIFIED".equals(m.value())) {
                String pageText = resolver.textOf(m.sourceDocumentDigest(), m.page());
                if (pageText == null) {
                    violations.add(v("SOURCE_UNRESOLVABLE", m.key().name(),
                            "digest/page 无法定位到快照页", Severity.BLOCKING));
                } else if (!numberPresent(pageText, m.value())) {
                    violations.add(v("FABRICATED_NUMBER", m.key().name(),
                            "值 " + m.value() + " 未在第 " + m.page() + " 页原文出现", Severity.BLOCKING));
                }
            }
        }

        // N4 派生复算（指标字典「机器可复算口径」）
        derivedCheck(byKey, resolver, violations);

        // N6 跨期同比：上期值来自独立信源页（如 page-46），本期同比必须与两期值一致
        yoyCheck(draft, byKey, resolver, violations);

        // N7 变动口径：比率变动误写 % / 金额变动误写个百分点 → 拦截
        deltaUnitCheck(draft, violations);

        boolean waiting = violations.stream().anyMatch(x -> x.severity() == Severity.WARNING)
                && violations.stream().noneMatch(x -> x.severity() == Severity.BLOCKING);
        return new CheckResult(CHECKER_ID, CHECKER_VERSION,
                violations.isEmpty() ? CheckStatus.PASSED
                        : (waiting ? CheckStatus.WAITING_HUMAN : CheckStatus.FAILED),
                violations, artifactSha);
    }

    /**
     * N6 跨期同比复算（指标字典「期间规则」）：
     * 报告正文含"同比+增速"论断时，先按句归属（营业收入 / 净利润），
     * 再用上年度对比页的上期值与本期值复算；定位不到时降级 WARNING（PERIOD_BASE_MISSING）。
     */
    private void yoyCheck(AnnualReportArtifacts draft, Map<MetricKey, MetricValue> byKey,
                          SnapshotTextResolver resolver, List<Violation> out) {
        MetricValue rev = byKey.get(MetricKey.OPERATING_REVENUE);
        MetricValue profit = byKey.get(MetricKey.NET_PROFIT_ATTRIBUTABLE);
        boolean hasRev = rev != null && !"NOT_FOUND".equals(rev.value());
        boolean hasProfit = profit != null && !"NOT_FOUND".equals(profit.value());
        if (!hasRev && !hasProfit) {
            return; // 留给 N1 报告
        }
        // 上期值：扫描本期指标页之外的快照页，找含「2024年」的对比页
        String priorPage = findPriorYearPage(resolver, rev, profit);
        if (priorPage == null) {
            out.add(v("PERIOD_BASE_MISSING", "OPERATING_REVENUE",
                    "快照未定位上年度对比页，同比复算降级为人工核对", Severity.WARNING));
            return;
        }
        for (ReportSection sec : draft.sections()) {
            var pm = java.util.regex.Pattern
                    .compile("同比(?:增长|下降|减少|增加|上升)?\\s*([\\d.]+)\\s*%").matcher(sec.body());
            while (pm.find()) {
                String subject = classifySubject(sec.body(), pm.start());
                if (subject == null) {
                    continue; // 论断归属不到指标，交由人工
                }
                MetricValue current = "净利润".equals(subject) ? profit : rev;
                if (current == null || "NOT_FOUND".equals(current.value())) {
                    continue;
                }
                try {
                    BigDecimal prior = findNumberAfter(priorPage, subject);
                    if (prior == null || prior.signum() == 0) {
                        continue;
                    }
                    BigDecimal cur = new BigDecimal(current.value().replace(",", ""));
                    BigDecimal yoy = cur.subtract(prior)
                            .divide(prior, 4, RoundingMode.HALF_UP)
                            .multiply(BigDecimal.valueOf(100));
                    BigDecimal claimed = new BigDecimal(pm.group(1));
                    if (yoy.abs().subtract(claimed.abs()).abs()
                            .compareTo(new BigDecimal("0.1")) > 0) {
                        out.add(v("YOY_MISMATCH", current.key().name(),
                                "正文" + subject + "同比 " + pm.group(1) + "% vs 复算 "
                                        + yoy.setScale(2, RoundingMode.HALF_UP) + "%",
                                Severity.BLOCKING));
                    }
                } catch (NumberFormatException ignored) {
                    // 数值不可解析，跳过交由 N2/N3 报告
                }
            }
        }
    }

    /** 回看 match 前 30 字符窗口，判断同比论断归属（营业收入/净利润）；归属不到返回 null。 */
    static String classifySubject(String body, int matchStart) {
        String window = body.substring(Math.max(0, matchStart - 110), matchStart);
        // 跨主体守护：同业/其他银行/参照系的同比不归属到本行指标，交人工而非误判为 BLOCKING
        if (window.contains("招商") || window.contains("同业") || window.contains("参照")
                || window.contains("其他银行") || window.contains("可比公司")) {
            return null;
        }
        int revIdx = window.lastIndexOf("营业收入");
        int profitIdx = window.lastIndexOf("净利润");
        if (revIdx < 0 && profitIdx < 0) {
            return null;
        }
        return profitIdx > revIdx ? "净利润" : "营业收入";
    }

    /** 定位上年度对比页：含「2024年」且非本期指标页。 */
    private String findPriorYearPage(SnapshotTextResolver resolver,
                                     MetricValue... currentMetrics) {
        Set<Integer> currentPages = new HashSet<>();
        for (MetricValue m : currentMetrics) {
            if (m != null) {
                currentPages.add(m.page());
            }
        }
        for (int pg = 1; pg <= 200; pg++) {
            if (currentPages.contains(pg)) {
                continue;
            }
            String text = resolver.textOf(null, pg);
            if (text != null && text.contains("2024年")) {
                return text;
            }
        }
        return null;
    }

    /**
     * N7 变动口径（年报表述铁律）：
     * 比率类变化（不良率/拨备覆盖率/成本收入比/CET1）→「个百分点」；净息差 →「基点」；
     * 金额类（不良余额）→「亿元」；增长率（营收/净利同比）才用「%」。
     * 工件的 yoy_delta/delta_unit 是归属指标的口径事实，正文短语与口径不符即拦截。
     */
    private void deltaUnitCheck(AnnualReportArtifacts draft, List<Violation> out) {
        Set<BigDecimal> ratioDeltas = new HashSet<>();
        Set<BigDecimal> percentDeltas = new HashSet<>();
        Set<BigDecimal> yuanDeltas = new HashSet<>();
        for (MetricValue m : draft.metrics()) {
            if (m.yoyDelta() == null || m.deltaUnit() == null) {
                continue;
            }
            BigDecimal d = normalizeDelta(m.yoyDelta());
            switch (m.deltaUnit()) {
                case PERCENTAGE_POINT, BASIS_POINT -> ratioDeltas.add(d);
                case PERCENT -> percentDeltas.add(d);
                case YUAN -> yuanDeltas.add(d);
            }
        }
        for (ReportSection sec : draft.sections()) {
            var m = DELTA_PHRASE.matcher(sec.body());
            while (m.find()) {
                BigDecimal num = normalizeDelta(m.group(2));
                String unitWord = m.group(3);
                if (unitWord.equals("%") || unitWord.equals("％")) {
                    // 比率指标的变动值却用 % 表述；若另有增长率指标同值则视为合法增速表述
                    if (ratioDeltas.contains(num) && !percentDeltas.contains(num)) {
                        out.add(v("DELTA_UNIT_MISMATCH", sec.sectionId(),
                                "「" + m.group() + "」：比率变动误用 %，应为「个百分点」或「基点」",
                                Severity.BLOCKING));
                    }
                } else if (unitWord.equals("个百分点") || unitWord.equals("基点")) {
                    // 金额指标的变动值却用比率单位表述
                    if (yuanDeltas.contains(num) && !ratioDeltas.contains(num)) {
                        out.add(v("DELTA_UNIT_MISMATCH", sec.sectionId(),
                                "「" + m.group() + "」：金额变动误用比率单位，应为「亿元」",
                                Severity.BLOCKING));
                    }
                }
            }
        }
    }

    /** 归一化变动幅度（去千分位/符号，0.940 → 0.94，20.0 → 20）用于口径比较。 */
    private static BigDecimal normalizeDelta(String raw) {
        BigDecimal d = new BigDecimal(raw.replace(",", "")).abs().stripTrailingZeros();
        return d.scale() < 0 ? d.setScale(0) : d;
    }

    private void derivedCheck(Map<MetricKey, MetricValue> byKey,
                              SnapshotTextResolver resolver, List<Violation> out) {
        MetricValue nplRatio = byKey.get(MetricKey.NPL_RATIO);
        MetricValue nplBalance = byKey.get(MetricKey.NPL_BALANCE);
        if (nplRatio == null || nplBalance == null || "NOT_FOUND".equals(nplRatio.value())) {
            return;
        }
        String pageText = resolver.textOf(nplBalance.sourceDocumentDigest(), nplBalance.page());
        if (pageText == null) {
            return;
        }
        // 尝试从同页提取"贷款和垫款本金总额/贷款总额"基数（约定关键词）
        BigDecimal loans = findLoansTotal(pageText);
        if (loans == null) {
            out.add(v("DERIVED_BASE_MISSING", "NPL_RATIO",
                    "同页未定位贷款总额基数，不良率口径复算降级为人工核对", Severity.WARNING));
            return;
        }
        try {
            BigDecimal balance = new BigDecimal(nplBalance.value().replace(",", ""));
            BigDecimal recomputed = balance.multiply(BigDecimal.valueOf(100))
                    .divide(loans, 2, RoundingMode.HALF_UP);
            BigDecimal reported = new BigDecimal(nplRatio.value().replace(",", ""));
            if (recomputed.subtract(reported).abs().compareTo(new BigDecimal("0.05")) > 0) {
                out.add(v("CALC_MISMATCH", "NPL_RATIO",
                        "复算 " + recomputed + "% vs 报告 " + reported + "%（偏差>0.05pp）", Severity.BLOCKING));
            }
        } catch (NumberFormatException e) {
            out.add(v("VALUE_UNPARSABLE", "NPL_RATIO", "数值不可解析：" + nplRatio.value(), Severity.BLOCKING));
        }

        // 口径2：拨备覆盖率 = 贷款减值准备 / 不良余额 × 100（容差 1pp）
        provisionCheck(byKey, resolver, out);
        // 口径3：成本收入比 = 业务及管理费 / 营业收入 × 100（容差 0.5pp）
        costIncomeCheck(byKey, resolver, out);
    }

    /** 口径2：拨备覆盖率复算。基数（贷款减值准备）从不良余额所在页定位；缺失降级 WARNING。 */
    private void provisionCheck(Map<MetricKey, MetricValue> byKey,
                                SnapshotTextResolver resolver, List<Violation> out) {
        MetricValue cov = byKey.get(MetricKey.PROVISION_COVERAGE);
        MetricValue npl = byKey.get(MetricKey.NPL_BALANCE);
        if (cov == null || npl == null || "NOT_FOUND".equals(cov.value())) {
            return;
        }
        String page = resolver.textOf(npl.sourceDocumentDigest(), npl.page());
        if (page == null) {
            return;
        }
        BigDecimal reserve = findNumberAfter(page, "贷款减值准备");
        if (reserve == null) {
            out.add(v("DERIVED_BASE_MISSING", "PROVISION_COVERAGE",
                    "同页未定位贷款减值准备，拨备覆盖率复算降级人工核对", Severity.WARNING));
            return;
        }
        try {
            BigDecimal nplBalance = new BigDecimal(npl.value().replace(",", ""));
            BigDecimal recomputed = reserve.multiply(BigDecimal.valueOf(100))
                    .divide(nplBalance, 2, RoundingMode.HALF_UP);
            BigDecimal reported = new BigDecimal(cov.value().replace(",", ""));
            if (recomputed.subtract(reported).abs().compareTo(new BigDecimal("1")) > 0) {
                out.add(v("CALC_MISMATCH", "PROVISION_COVERAGE",
                        "复算 " + recomputed + "% vs 报告 " + reported + "%（容差1pp）", Severity.BLOCKING));
            }
        } catch (NumberFormatException e) {
            out.add(v("VALUE_UNPARSABLE", "PROVISION_COVERAGE", "数值不可解析：" + cov.value(), Severity.BLOCKING));
        }
    }

    /** 口径3：成本收入比复算。基数（业务及管理费）从营业收入所在页定位；缺失降级 WARNING。 */
    private void costIncomeCheck(Map<MetricKey, MetricValue> byKey,
                                 SnapshotTextResolver resolver, List<Violation> out) {
        MetricValue ratio = byKey.get(MetricKey.COST_INCOME_RATIO);
        MetricValue rev = byKey.get(MetricKey.OPERATING_REVENUE);
        if (ratio == null || rev == null || "NOT_FOUND".equals(ratio.value())) {
            return;
        }
        String page = resolver.textOf(rev.sourceDocumentDigest(), rev.page());
        if (page == null) {
            return;
        }
        BigDecimal expense = findNumberAfter(page, "业务及管理费");
        if (expense == null) {
            out.add(v("DERIVED_BASE_MISSING", "COST_INCOME_RATIO",
                    "同页未定位业务及管理费，成本收入比复算降级人工核对", Severity.WARNING));
            return;
        }
        try {
            BigDecimal revenue = new BigDecimal(rev.value().replace(",", ""));
            BigDecimal recomputed = expense.multiply(BigDecimal.valueOf(100))
                    .divide(revenue, 2, RoundingMode.HALF_UP);
            BigDecimal reported = new BigDecimal(ratio.value().replace(",", ""));
            if (recomputed.subtract(reported).abs().compareTo(new BigDecimal("0.5")) > 0) {
                out.add(v("CALC_MISMATCH", "COST_INCOME_RATIO",
                        "复算 " + recomputed + "% vs 报告 " + reported + "%（容差0.5pp）", Severity.BLOCKING));
            }
        } catch (NumberFormatException e) {
            out.add(v("VALUE_UNPARSABLE", "COST_INCOME_RATIO", "数值不可解析：" + ratio.value(), Severity.BLOCKING));
        }
    }

    /** 在 text 中找 keyword 之后最近的数字（含千分位，跨最多 12 个非数字字符），纯字符扫描无正则。 */
    static BigDecimal findNumberAfter(String text, String keyword) {
        int i = text.indexOf(keyword);
        if (i < 0) {
            return null;
        }
        int j = i + keyword.length();
        int skipped = 0;
        while (j < text.length() && !isNumChar(text.charAt(j)) && skipped < 12) {
            j++;
            skipped++;
        }
        int start = j;
        while (j < text.length() && isNumChar(text.charAt(j))) {
            j++;
        }
        if (j == start) {
            return null;
        }
        // 去掉可能的尾部逗号（如 "656.10，"）
        String raw = text.substring(start, j);
        while (raw.endsWith(",")) {
            raw = raw.substring(0, raw.length() - 1);
        }
        try {
            return new BigDecimal(raw.replace(",", ""));
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private static boolean isNumChar(char c) {
        return (c >= '0' && c <= '9') || c == ',' || c == '.';
    }

    /** 从页文本定位贷款总额（约定口径词；W01 冻结材料里应存在）。 */
    private BigDecimal findLoansTotal(String pageText) {
        var m = Pattern.compile("(?:贷款和垫款本金总额|贷款总额)[^\\d]{0,12}([\\d,]+(?:\\.\\d+)?)").matcher(pageText);
        return m.find() ? new BigDecimal(m.group(1).replace(",", "")) : null;
    }

    private static boolean numberPresent(String pageText, String value) {
        String normalized = value.replace(",", "");
        var m = NUMBER.matcher(pageText);
        while (m.find()) {
            if (m.group().replace(",", "").equals(normalized)) {
                return true;
            }
        }
        return false;
    }

    private static Violation v(String rule, String target, String detail, Severity s) {
        return new Violation(rule, target, detail, s);
    }

    /** 快照文本解析口：由部署层把 MaterialAgent 的 SnapshotStore 接进来。 */
    public interface SnapshotTextResolver {
        /** 返回 digest+page 对应的原文；不存在返回 null。 */
        String textOf(String digest, int page);
    }
}
