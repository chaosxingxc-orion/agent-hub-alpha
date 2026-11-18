/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public record ReportSection(
        @JsonProperty("section_id") String sectionId,
        @JsonProperty("title") String title,
        @JsonProperty("body") String body,                // 正文内所有论断以 [C-n] 标注
        @JsonProperty("claim_ids") List<String> claimIds  // 本章论断 id 清单
) {}

/** 检查结论 —— checker 的输出，也是 GatekeepRail 的门禁输入。 */
