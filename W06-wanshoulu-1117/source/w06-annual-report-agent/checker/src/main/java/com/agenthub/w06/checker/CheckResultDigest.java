/*
 * CheckResultDigest —— 工件指纹计算（P1/E03 防线的共用工具）。
 *
 * 核验结论必须回答"验的是哪份工件"：对 AnnualReportArtifacts 的规范
 * JSON 序列化取 sha256。Go 侧交付谓词要求三条记录绑定同一指纹；
 * 指纹缺失或与当前工件不一致的结论一律无效（validFor 判定）。
 */

package com.agenthub.w06.checker;

import com.agenthub.w06.contract.AnnualReportArtifacts;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.security.MessageDigest;
import java.util.HexFormat;

public final class CheckResultDigest {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private CheckResultDigest() {}

    /** 工件规范序列化的 sha256（十六进制，64 字符）。序列化失败返回 null（不可绑定）。 */
    public static String of(AnnualReportArtifacts draft) {
        if (draft == null) {
            return null;
        }
        try {
            byte[] bytes = MAPPER.writeValueAsBytes(draft);
            MessageDigest md = MessageDigest.getInstance("SHA-256");
            return HexFormat.of().formatHex(md.digest(bytes));
        } catch (Exception e) {
            return null;   // 不可绑定的结论由消费方按无效处理（fail-closed）
        }
    }
}
