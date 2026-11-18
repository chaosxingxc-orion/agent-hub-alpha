/*
 * SourceRegistry — 信源注册表（P0-1）
 *
 * 对标专业报告 agent 的信源管理：每个信源条目带权威等级、获取时间、内容摘要，
 * 指标值必须来自官方一手信源。引用核验时校验信源存在性与权威等级约束。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

/** 权威等级：官方一手 > 机构二手 > 聚合三级。 */public record SourceEntry(
        @JsonProperty("source_id") String sourceId,
        @JsonProperty("title") String title,                    // 如「招商银行 2025 年度报告（经营概况节选）」
        @JsonProperty("authority") SourceAuthority authority,
        @JsonProperty("origin_url") String originUrl,           // 原始获取地址（快照声明，不实时访问）
        @JsonProperty("fetched_at") String fetchedAt,           // W01 抓取时间
        @JsonProperty("content_sha256") String contentSha256,   // 该信源内容的摘要
        @JsonProperty("pages") List<Integer> pages              // 该信源覆盖的快照页
) {}

/** 引用支持度三级（对齐 CiteEval full/partial/no support）。 */
