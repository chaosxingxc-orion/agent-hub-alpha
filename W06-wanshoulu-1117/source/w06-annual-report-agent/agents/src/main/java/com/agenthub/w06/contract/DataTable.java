/*
 * W06 数据工件契约 —— 结构化财报数据表（数据能力补强）。
 * 由 table_extract 工具从快照页原文抽取，逐格保留原文字符串（含千分位/括号负数/%），
 * source_page + source_document_digest 绑定来源可打开；table_analyze 在其上做确定性复算。
 */
package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import java.util.Map;

public record DataTable(
        @JsonProperty("table_id") String tableId,          // T-1、T-2 …
        @JsonProperty("title") String title,                // 表名（规范中文）
        @JsonProperty("unit") String unit,                  // 原文单位（人民币百万元，百分比除外）
        @JsonProperty("columns") List<String> columns,      // 列标签（2025年/2024年/占比(%)…）
        @JsonProperty("rows") List<TableRow> rows,          // 数据行
        @JsonProperty("source_page") int sourcePage,        // 来源页（快照页号）
        @JsonProperty("source_document_digest") String sourceDocumentDigest, // 信源指纹
        @JsonProperty("analysis") Map<String, Object> analysis // table_analyze 的复算结论（可空）
) {}
