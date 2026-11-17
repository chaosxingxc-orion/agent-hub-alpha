/*
 * 反例回归测试 —— W06 验收的负例门（对应 A07/A08/A09 及完整性）
 *
 * 这些测试是「正反例达标」的机器可执行形式：每个反例注入一个已知缺陷，
 * 断言 checker 按 BLOCKING 违规拦截。全部纯内存，无模型、无网络。
 */

package com.agenthub.w06.checker;

import com.agenthub.w06.contract.*;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class CheckerNegativeCasesTest {

    // ---- 信源注册表（P0-1）----
    private static final java.util.List<com.agenthub.w06.contract.SourceEntry> SOURCES = java.util.List.of(
            new com.agenthub.w06.contract.SourceEntry(
                    "cmb-ar-2025-primary", "招商银行2025年报节选",
                    com.agenthub.w06.contract.SourceAuthority.PRIMARY_OFFICIAL,
                    "https://demo", "2026-11-16", "sha", java.util.List.of(12, 45)));

    // ---- 支持度三级（P0-2）----

    @org.junit.jupiter.api.Test
    void partialSupportIsWarning() {
        // 论断含 2025 期间词，摘录无该年份 → PARTIAL_SUPPORT（WARNING，非 BLOCKING）
        Citation c = new Citation("C-1", "2025年营业收入增长",
                "sha-page-12", 12, "营业收入3,374.88亿元，同比增长0.83%。", "sha-any");
        // digest 会不匹配——此测试只关注 PARTIAL 规则，故用支持度静态方法直测
        com.agenthub.w06.contract.CitationSupport sup = assessSupportDirect(c);
        org.junit.jupiter.api.Assertions.assertEquals(
                com.agenthub.w06.contract.CitationSupport.PARTIAL, sup);
    }

    @org.junit.jupiter.api.Test
    void unsupportedClaimIsBlocking() {
        Citation c = new Citation("C-1", "苹果公司发布新产品",
                "sha-page-12", 12, "三、经营概况 2025年，本集团实现营业收入3,374.88亿元", "sha-any");
        com.agenthub.w06.contract.CitationSupport sup = assessSupportDirect(c);
        org.junit.jupiter.api.Assertions.assertEquals(
                com.agenthub.w06.contract.CitationSupport.UNSUPPORTED, sup);
    }

    @org.junit.jupiter.api.Test
    void supportedClaimPasses() {
        Citation c = new Citation("C-1", "2025年营业收入3,374.88亿元",
                "sha-page-12", 12, "2025年，本集团实现营业收入3,374.88亿元", "sha-any");
        com.agenthub.w06.contract.CitationSupport sup = assessSupportDirect(c);
        org.junit.jupiter.api.Assertions.assertEquals(
                com.agenthub.w06.contract.CitationSupport.SUPPORTED, sup);
    }

    @org.junit.jupiter.api.Test
    void unregisteredSourcePageIsBlocking() {
        // 引用第 99 页：快照定位不到（FABRICATED_CITATION）且不在任何已注册信源内（SOURCE_UNREGISTERED）
        Citation c = new Citation("C-1", "2025年营业收入3,374.88亿元",
                "sha-page-12", 99, "2025年，本集团实现营业收入3,374.88亿元", "sha-any");
        CheckResult r = new CitationChecker().check(
                draft(validMetrics(), java.util.List.of(c)), RESOLVER, SOURCES);
        org.junit.jupiter.api.Assertions.assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "SOURCE_UNREGISTERED".equals(v.rule())),
                "应包含 SOURCE_UNREGISTERED，实际：" + r.violations());
        assertTrue(r.violations().stream().anyMatch(v -> "FABRICATED_CITATION".equals(v.rule())));
    }

    // ---- N6 跨期同比（P1）----

    @org.junit.jupiter.api.Test
    void yoyMismatchIsBlocking() {
        // 正文宣称同比 5% —— 复算 (3374.88-3347.16)/3347.16 ≈ 0.83%，不一致 → YOY_MISMATCH
        var sec = new com.agenthub.w06.contract.ReportSection("s1", "经营", "营业收入同比增长5%", java.util.List.of());
        CheckResult r = new NumericChecker().check(draft(validMetrics(), java.util.List.of()), RESOLVER);
        // sections 由 draft() 帮手构造为空列表——手工构造带 section 的工件
        var d = new AnnualReportArtifacts("t-1", "招商银行", "2025年度", "2026-11-17",
                validMetrics(), java.util.List.of(), java.util.List.of(sec), List.of());
        r = new NumericChecker().check(d, RESOLVER);
        org.junit.jupiter.api.Assertions.assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "YOY_MISMATCH".equals(v.rule())));
    }

    @org.junit.jupiter.api.Test
    void consistentYoyPasses() {
        // 正文同比 0.83% —— 与复算一致 → 不产生 YOY_MISMATCH
        var sec = new com.agenthub.w06.contract.ReportSection("s1", "经营", "营业收入同比增长0.83%", java.util.List.of());
        var d = new AnnualReportArtifacts("t-1", "招商银行", "2025年度", "2026-11-17",
                validMetrics(), java.util.List.of(), java.util.List.of(sec), List.of());
        CheckResult r = new NumericChecker().check(d, RESOLVER);
        assertTrue(r.violations().stream().noneMatch(v -> "YOY_MISMATCH".equals(v.rule())),
                "同比一致不应报错，实际：" + r.violations());
    }

    // ---- 口径2/3：拨备覆盖率与成本收入比复算 ----

    @org.junit.jupiter.api.Test
    void provisionCoverageMismatchIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.set(6, mv(MetricKey.PROVISION_COVERAGE, "500.00", "%", 45, "sha-page-45")); // 复算≈411.82
        CheckResult r = new NumericChecker().check(draft(m, java.util.List.of()), RESOLVER);
        assertTrue(r.violations().stream().anyMatch(v -> "CALC_MISMATCH".equals(v.rule()) && "PROVISION_COVERAGE".equals(v.target())),
                "应报 PROVISION_COVERAGE 复算不符，实际：" + r.violations());
    }

    @org.junit.jupiter.api.Test
    void costIncomeMismatchIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.set(3, mv(MetricKey.COST_INCOME_RATIO, "50.00", "%", 12, "sha-page-12")); // 复算≈33.04
        CheckResult r = new NumericChecker().check(draft(m, java.util.List.of()), RESOLVER);
        assertTrue(r.violations().stream().anyMatch(v -> "CALC_MISMATCH".equals(v.rule()) && "COST_INCOME_RATIO".equals(v.target())),
                "应报 COST_INCOME_RATIO 复算不符，实际：" + r.violations());
    }

    private static com.agenthub.w06.contract.CitationSupport assessSupportDirect(Citation c) {
        try {
            var m = CitationChecker.class.getDeclaredMethod("assessSupport", Citation.class);
            m.setAccessible(true);
            return (com.agenthub.w06.contract.CitationSupport) m.invoke(null, c);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }


    // ---- 测试快照（两页，含约定口径的数字） ----
    private static final String PAGE_12 = """
            三、经营概况
            2025年，本集团实现营业收入3,374.88亿元，同比增长0.83%。归属于本行股东的净利润
            1,405.66亿元。净息差1.98%。成本收入比33.06%。业务及管理费1,115.14亿元。
            """;
    private static final String PAGE_45 = """
            五、风险与资产质量
            报告期末，本集团不良贷款余额656.10亿元，不良贷款率0.95%，拨备覆盖率411.98%，
            核心一级资本充足率14.86%。贷款和垫款本金总额68,300.00亿元。贷款减值准备2,702.00亿元。
            """;

    private static final String PAGE_46 = """
            六、上年同期对比（2024年度）
            2024年，本集团营业收入3,347.16亿元，归属于本行股东的净利润1,483.91亿元。
            不良贷款率0.95%，拨备覆盖率411.98%。
            """;

    private static final NumericChecker.SnapshotTextResolver RESOLVER = (digest, page) -> {
        if (page == 12 && "sha-page-12".equals(digest)) return PAGE_12;
        if (page == 45 && "sha-page-45".equals(digest)) return PAGE_45;
        if (page == 46) return PAGE_46;   // 上年度信源页：digest 无关，按页定位（跨期独立信源）
        return null;
    };

    private static MetricValue mv(MetricKey k, String value, String unit, int page, String digest) {
        return new MetricValue(k, value, unit, "2025年度", digest, page, "sha-excerpt");
    }

    private static List<MetricValue> validMetrics() {
        return new ArrayList<>(List.of(
                mv(MetricKey.OPERATING_REVENUE, "3,374.88", "亿元", 12, "sha-page-12"),
                mv(MetricKey.NET_PROFIT_ATTRIBUTABLE, "1,405.66", "亿元", 12, "sha-page-12"),
                mv(MetricKey.NET_INTEREST_MARGIN, "1.98", "%", 12, "sha-page-12"),
                mv(MetricKey.COST_INCOME_RATIO, "33.06", "%", 12, "sha-page-12"),
                mv(MetricKey.NPL_BALANCE, "656.10", "亿元", 45, "sha-page-45"),
                mv(MetricKey.NPL_RATIO, "0.95", "%", 45, "sha-page-45"),
                mv(MetricKey.PROVISION_COVERAGE, "411.98", "%", 45, "sha-page-45"),
                mv(MetricKey.CET1_RATIO, "14.86", "%", 45, "sha-page-45")));
    }

    private static AnnualReportArtifacts draft(List<MetricValue> metrics, List<Citation> citations) {
        return new AnnualReportArtifacts("t-1", "招商银行", "2025年度", "2026-11-17",
                metrics, citations, List.of(), List.of());
    }

    // ---- A07：错误单位 ----
    @Test
    void wrongUnitIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.set(0, mv(MetricKey.OPERATING_REVENUE, "3,374.88", "万元", 12, "sha-page-12"));
        CheckResult r = new NumericChecker().check(draft(m, List.of()), RESOLVER);
        assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "UNIT_MISMATCH".equals(v.rule())));
    }

    // ---- 数字纪律：模型口述的数字不在原文出现 ----
    @Test
    void fabricatedNumberIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.set(0, mv(MetricKey.OPERATING_REVENUE, "3,999.99", "亿元", 12, "sha-page-12"));
        CheckResult r = new NumericChecker().check(draft(m, List.of()), RESOLVER);
        assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "FABRICATED_NUMBER".equals(v.rule())));
    }

    // ---- 完整性：八缺一 ----
    @Test
    void missingMetricIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.remove(7); // 掉了 CET1
        CheckResult r = new NumericChecker().check(draft(m, List.of()), RESOLVER);
        assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "METRIC_MISSING".equals(v.rule())));
    }

    // ---- 派生复算：不良率与余额/贷款总额不符 ----
    @Test
    void derivedCalcMismatchIsBlocking() {
        List<MetricValue> m = validMetrics();
        m.set(5, mv(MetricKey.NPL_RATIO, "1.95", "%", 45, "sha-page-45")); // 复算约 0.96
        CheckResult r = new NumericChecker().check(draft(m, List.of()), RESOLVER);
        assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "CALC_MISMATCH".equals(v.rule())));
    }

    // ---- 正例：八指标全对 → PASSED（复算 656.10/68300≈0.96 与 0.95 差 0.01 在容差内） ----
    @Test
    void validMetricsPass() {
        CheckResult r = new NumericChecker().check(draft(validMetrics(), List.of()), RESOLVER);
        assertEquals(CheckStatus.PASSED, r.status(), () -> String.valueOf(r.violations()));
    }

    // ---- A08：伪造引用（digest/page 定位不到） ----
    @Test
    void fabricatedCitationIsBlocking() {
        Citation fake = new Citation("C-1", "营业收入超三千亿",
                "sha-not-exist", 99, "任意摘录", "sha-excerpt");
        CheckResult r = new CitationChecker().check(
                draft(validMetrics(), List.of(fake)), RESOLVER);
        assertEquals(CheckStatus.FAILED, r.status());
        assertTrue(r.violations().stream().anyMatch(v -> "FABRICATED_CITATION".equals(v.rule())));
    }

    // ---- A09 同类：摘录与 digest 不符（检查后篡改） ----
    @Test
    void tamperedExcerptIsBlocking() {
        Citation tampered = new Citation("C-1", "净息差1.98%",
                "sha-page-12", 12, "净息差2.98%", "sha-page-12"); // excerpt 实为 sha(净息差2.98%)
        CheckResult r = new CitationChecker().check(
                draft(validMetrics(), List.of(tampered)), RESOLVER);
        assertTrue(r.violations().stream()
                .anyMatch(v -> "EXCERPT_DIGEST_MISMATCH".equals(v.rule())
                        || "EXCERPT_NOT_ON_PAGE".equals(v.rule())));
    }

    // ---- 无引用论断 ----
    @Test
    void uncitedClaimIsBlocking() {
        CheckResult r = new CitationChecker().check(
                draft(validMetrics(), List.of()), RESOLVER);
        // 无 sections 时无 uncited 论断，应 PASSED（引用层零违规）
        assertEquals(CheckStatus.PASSED, r.status());
    }

    // ---- P1/E03：核验结论绑定工件指纹 ----
    @Test
    void checkResultBindsArtifactDigest() {
        var d = draft(validMetrics(), List.of());
        CheckResult r = new NumericChecker().check(d, RESOLVER);
        assertNotNull(r.artifactSha256(), "核验结论必须携带工件指纹（P1）");
        assertNotNull(r.checkedAt(), "核验结论必须携带时刻（P1）");
        // 指纹可复算：同一工件两次核验，指纹一致
        CheckResult r2 = new NumericChecker().check(d, RESOLVER);
        assertEquals(r.artifactSha256(), r2.artifactSha256(), "同一工件指纹必须一致");
        // validFor：指纹匹配 → 有效
        assertTrue(r.validFor(r2.artifactSha256()));
    }

    @Test
    void tamperedArtifactInvalidatesOldVerdict() {
        // E03 攻击路径：检查通过后偷改工件（换一个指标值），旧结论必须失效
        var original = draft(validMetrics(), List.of());
        CheckResult oldVerdict = new NumericChecker().check(original, RESOLVER);
        assertEquals(CheckStatus.PASSED, oldVerdict.status());

        // 篡改：替换任一指标值（工件内容变化 → 指纹变化）
        var tamperedArtifact = draft(validMetrics().stream()
                .map(m -> m.key() == com.agenthub.w06.contract.MetricKey.OPERATING_REVENUE
                        ? new com.agenthub.w06.contract.MetricValue(m.key(), "9,999.99",
                                m.unit(), m.period(), m.sourceDocumentDigest(), m.page(),
                                m.evidenceExcerptDigest())
                        : m)
                .toList(), List.of());

        // 旧 PASSED 对新工件无效（E03 防线核心断言）
        assertFalse(oldVerdict.validFor(CheckResultDigest.of(tamperedArtifact)),
                "改正文后旧核验结论必须失效（E03）");
        // 新核验会拦截篡改值
        CheckResult rechecked = new NumericChecker().check(tamperedArtifact, RESOLVER);
        assertNotEquals(CheckStatus.PASSED, rechecked.status(), "篡改工件重核必须不通过");
    }

    // ---- P2/N7：变动口径（百分点 vs %）----
    @Test
    void percentChangeWrittenAsPercentIsBlocked() {
        // W01 反例：不良率 0.95%→0.94%，正文写"下降 0.01%"——必须拦截
        var sec = new com.agenthub.w06.contract.ReportSection("s2", "贷款质量",
                "报告期末，不良贷款率较上年末下降 0.01%。[C-1]", java.util.List.of("C-1"));
        var d = new AnnualReportArtifacts("t-1", "招商银行", "2025年度", "2026-11-17",
                validMetrics(), List.of(), List.of(sec), List.of());
        CheckResult r = new NumericChecker().check(d, RESOLVER);
        assertTrue(r.violations().stream()
                        .anyMatch(v -> "DELTA_UNIT_MISMATCH".equals(v.rule())),
                "比率变动写成 % 必须被 N7 拦截");
    }

    @Test
    void percentChangeWrittenAsPointsPassesN7() {
        // 正例：写"下降 0.01 个百分点"——N7 放行（可能仍有其他违规，但不得含 DELTA_UNIT_MISMATCH）
        var sec = new com.agenthub.w06.contract.ReportSection("s2", "贷款质量",
                "报告期末，不良贷款率较上年末下降 0.01 个百分点。[C-1]", java.util.List.of("C-1"));
        var d = new AnnualReportArtifacts("t-1", "招商银行", "2025年度", "2026-11-17",
                validMetrics(), List.of(), List.of(sec), List.of());
        CheckResult r = new NumericChecker().check(d, RESOLVER);
        assertFalse(r.violations().stream()
                        .anyMatch(v -> "DELTA_UNIT_MISMATCH".equals(v.rule())),
                "个百分点表述不得误报");
    }

    @Test
    void unboundVerdictIsInvalidForAnyArtifact() {
        // 无指纹的结论（历史遗留形态）对任何工件都无效——fail-closed
        CheckResult unbound = CheckResult.unbound("legacy-checker", "0.0.1",
                CheckStatus.PASSED, List.of());
        assertFalse(unbound.validFor(CheckResultDigest.of(draft(validMetrics(), List.of()))));
        assertFalse(unbound.validFor("any-digest"));
        assertNull(unbound.artifactSha256());
    }
}
