/*
 * W06 数据工件契约 —— 表格数据行：项目名 + 逐格原文值。
 */
package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public record TableRow(
        @JsonProperty("item") String item,
        @JsonProperty("values") List<String> values
) {}
