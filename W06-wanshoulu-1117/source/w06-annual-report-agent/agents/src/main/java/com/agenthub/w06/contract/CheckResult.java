/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public record CheckResult(
        @JsonProperty("checker_id") String checkerId,       // numeric-checker / citation-checker
        @JsonProperty("checker_version") String checkerVersion,
        @JsonProperty("status") CheckStatus status,
        @JsonProperty("violations") List<Violation> violations,
        @JsonProperty("artifact_sha256") String artifactSha256 // G3：被检工件（规范化 JSON）摘要，verdict 与工件绑定，防旧结论放行新工件
) {}
