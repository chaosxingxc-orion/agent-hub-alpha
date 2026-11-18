/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true) // 兼容数据能力补强新增的 tables 字段（核验器只关心 metrics/citations/sections）
public record AnnualReportArtifacts(
        @JsonProperty("task_id") String taskId,
        @JsonProperty("company") String company,            // 招商银行
        @JsonProperty("report_period") String reportPeriod, // 2025 年度
        @JsonProperty("generated_at") String generatedAt,   // 工件生成时间（ISO-8601，可缺省）
        @JsonProperty("metrics") List<MetricValue> metrics,          // 8 条，缺一 FAIL（完整性）
        @JsonProperty("citations") List<Citation> citations,         // 每条论断至少一条
        @JsonProperty("sections") List<ReportSection> sections,      // 正文分节
        @JsonProperty("charts") List<java.util.Map<String, Object>> charts // 图表（bar/line/pie/flow），前端内联 SVG 渲染；核验器不校验
) {}

/** 报告正文章节。 */
