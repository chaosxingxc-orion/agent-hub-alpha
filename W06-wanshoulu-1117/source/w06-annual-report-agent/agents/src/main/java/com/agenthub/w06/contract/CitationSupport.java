/*
 * SourceRegistry — 信源注册表（P0-1）
 *
 * 对标专业报告 agent 的信源管理：每个信源条目带权威等级、获取时间、内容摘要，
 * 指标值必须来自官方一手信源。引用核验时校验信源存在性与权威等级约束。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

/** 权威等级：官方一手 > 机构二手 > 聚合三级。 */public enum CitationSupport {
    SUPPORTED,    // 摘录完整支撑论断（数字/期间/口径/主体全对上）
    PARTIAL,      // 相关但不完整（如只有数字没有口径）→ WARNING + 建议 human
    UNSUPPORTED   // 不支撑或引用伪造 → BLOCKING
}
