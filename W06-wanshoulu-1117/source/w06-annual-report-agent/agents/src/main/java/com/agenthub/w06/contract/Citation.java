/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public record Citation(
        @JsonProperty("claim_id") String claimId,       // 正文中的论断标识，如 "C-3"
        @JsonProperty("claim_text") String claimText,   // 论断原文
        @JsonProperty("source_document_digest") String sourceDocumentDigest,
        @JsonProperty("page") int page,
        @JsonProperty("evidence_excerpt") String evidenceExcerpt, // 原文摘录（支持度核验对象）
        @JsonProperty("evidence_excerpt_digest") String evidenceExcerptDigest
) {}

/** 结构化年度报告工件 —— writer 的唯一正式输出形态。 */
