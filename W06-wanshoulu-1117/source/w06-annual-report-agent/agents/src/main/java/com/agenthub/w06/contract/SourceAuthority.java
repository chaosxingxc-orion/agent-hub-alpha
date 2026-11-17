/*
 * SourceRegistry — 信源注册表（P0-1）
 *
 * 对标专业报告 agent 的信源管理：每个信源条目带权威等级、获取时间、内容摘要，
 * 指标值必须来自官方一手信源。引用核验时校验信源存在性与权威等级约束。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

/** 权威等级：官方一手 > 机构二手 > 聚合三级。 */public enum SourceAuthority {
    PRIMARY_OFFICIAL("官方一手：公司披露原文/监管披露平台"),
    SECONDARY_INSTITUTIONAL("机构二手：券商研报/评级机构/行业协会"),
    TERTIARY_AGGREGATED("聚合三级：财经媒体/数据聚合商");

    private final String description;
    SourceAuthority(String description) { this.description = description; }
    public String getDescription() { return description; }
}

/** 信源条目 —— 快照 manifest 里注册，报告引用必须命中。 */
