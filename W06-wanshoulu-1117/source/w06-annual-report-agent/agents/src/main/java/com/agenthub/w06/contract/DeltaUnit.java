/*
 * W06 产物契约 —— 同比变动口径单位。
 *
 * 年报表述铁律（真实年报 3.1/3.4 的写法）：
 *   比率类指标（不良率/拨备覆盖率/成本收入比/CET1）的变化 → 「个百分点」；
 *   净息差 → 「基点」；
 *   增长率（营收/净利同比）→ 「%」；
 *   金额类（不良余额）→ 「亿元」。
 * 写错口径（如「下降20.19%」实为百分点）是年报级表述错误，checker 按
 * DELTA_UNIT_MISMATCH 拦截（N7）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;

public enum DeltaUnit {
    @JsonProperty("PERCENTAGE_POINT") PERCENTAGE_POINT("个百分点"),
    @JsonProperty("BASIS_POINT") BASIS_POINT("基点"),
    @JsonProperty("PERCENT") PERCENT("%"),
    @JsonProperty("YUAN") YUAN("亿元");

    private final String label;

    DeltaUnit(String label) { this.label = label; }

    public String getLabel() { return label; }
}
