/*
 * TableTools —— ToolGateway 的两个数据工具入口：
 *   table_extract  : 从快照页原文抽取结构化财报表格（逐格保留原文数值）
 *   table_analyze  : 对结构化表做确定性汇总分析（同比/占比/结构/合计校验）
 * 遵循纪律：只消费已落盘快照页（live / 指定银行），不联网、不臆造数值。
 */
package com.agenthub.w06.checker.server;

import com.agenthub.w06.contract.DataTable;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.*;

final class TableTools {

    /** table_extract 入口。args: {bank?, pages?:[int], page?:int, title?} */
    static ObjectNode extract(JsonNode args, CheckerHttpServer.SnapshotHttpResolver resolver) {
        String bank = args.path("bank").asText("live").trim();
        if (bank.isEmpty()) bank = "live";
        CheckerHttpServer.BankSnap snap = resolver.banks.get(bank);
        if (snap == null) return err("bank_missing", "快照集不存在: " + bank + "（可用: "
                + String.join(",", resolver.banks.keySet()) + "）");

        List<Integer> pages = new ArrayList<>();
        JsonNode pagesNode = args.path("pages");
        if (pagesNode.isArray() && !pagesNode.isEmpty()) {
            for (JsonNode p : pagesNode) pages.add(p.asInt());
        } else if (args.hasNonNull("page")) {
            pages.add(args.path("page").asInt());
        } else {
            pages.addAll(snap.byPageNumber.keySet()); // 全扫描
        }

        // 页 → 信源 digest 反查
        Map<Integer, String> pageDigest = new HashMap<>();
        for (var e : snap.sourcePages.entrySet()) {
            for (int pg : e.getValue()) pageDigest.put(pg, e.getKey());
        }

        List<DataTable> all = new ArrayList<>();
        for (int pg : pages) {
            String text = snap.byPageNumber.get(pg);
            if (text == null) continue;
            String digest = pageDigest.getOrDefault(pg, "sha256:page-" + pg);
            List<String> titles = new ArrayList<>();
            JsonNode tn = args.path("titles");
            if (tn.isArray()) for (JsonNode t : tn) titles.add(t.asText());
            all.addAll(TableExtractor.extract(text, pg, digest, titles));
        }
        return TableExtractor.toResultJson(all);
    }

    /** table_analyze 入口。args: {pages?:[int], tables?:[{...}]}，tables 可由调用方直接传入。 */
    static ObjectNode analyze(JsonNode args, CheckerHttpServer.SnapshotHttpResolver resolver) {
        JsonNode tablesNode = args.path("tables");
        List<DataTable> tables = new ArrayList<>();
        if (tablesNode.isArray() && !tablesNode.isEmpty()) {
            for (JsonNode t : tablesNode) {
                try {
                    tables.add(CheckerHttpServer.MAPPER.treeToValue(t, DataTable.class));
                } catch (Exception ignored) { }
            }
        } else {
            // 未传 tables：先对指定页做抽取再分析
            ObjectNode ex = extract(args, resolver);
            if (!ex.path("ok").asBoolean(false)) return ex;
            ArrayNode arr = ex.withArray("tables");
            for (JsonNode t : arr) {
                try {
                    tables.add(CheckerHttpServer.MAPPER.treeToValue(t, DataTable.class));
                } catch (Exception ignored) { }
            }
        }
        return DataAnalysis.toResultJson(tables);
    }

    private static ObjectNode err(String code, String message) {
        ObjectNode e = CheckerHttpServer.MAPPER.createObjectNode();
        e.put("ok", false);
        ObjectNode er = e.putObject("error");
        er.put("code", code);
        er.put("message", message);
        return e;
    }
}
