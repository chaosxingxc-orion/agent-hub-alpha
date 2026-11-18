/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public record Violation(
        @JsonProperty("rule") String rule,               // 如 METRIC_MISSING / UNIT_MISMATCH / FABRICATED_CITATION
        @JsonProperty("target") String target,           // 指标 key / claim_id
        @JsonProperty("detail") String detail,
        @JsonProperty("severity") Severity severity
) {}
