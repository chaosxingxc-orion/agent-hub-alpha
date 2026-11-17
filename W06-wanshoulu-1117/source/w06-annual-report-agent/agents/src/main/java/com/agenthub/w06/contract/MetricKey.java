/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public enum MetricKey {
    OPERATING_REVENUE("营业收入"),
    NET_PROFIT_ATTRIBUTABLE("归母净利润"),
    NET_INTEREST_MARGIN("净息差"),
    COST_INCOME_RATIO("成本收入比"),
    NPL_BALANCE("不良贷款余额"),
    NPL_RATIO("不良贷款率"),
    PROVISION_COVERAGE("拨备覆盖率"),
    CET1_RATIO("核心一级资本充足率");

    private final String displayName;

    MetricKey(String displayName) { this.displayName = displayName; }

    public String getDisplayName() { return displayName; }
}

/** 单个指标值 —— 每个字段都是 checker 的核验对象。 */
