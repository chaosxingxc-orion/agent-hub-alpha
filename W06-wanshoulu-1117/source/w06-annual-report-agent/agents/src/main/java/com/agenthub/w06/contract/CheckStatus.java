/*
 * W06 产物契约 —— Issue #16 / v2 方案 §4 的结构化工件层。
 * 先结构化、再渲染；每个数值携带 digest/page/excerpt 三元组（"来源可打开"）。
 */

package com.agenthub.w06.contract;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;

public enum CheckStatus {
    PASSED,        // 全部通过
    FAILED,        // 存在违规 → GatekeepRail 触发一次定向修订
    WAITING_HUMAN  // 无法自动判定的语义问题 → 转 waiting_human（A08）
}
