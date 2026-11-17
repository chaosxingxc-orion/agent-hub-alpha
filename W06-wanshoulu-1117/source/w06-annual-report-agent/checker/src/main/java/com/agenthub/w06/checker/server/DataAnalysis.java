/*
 * DataAnalysis —— 数据汇总分析工具（数据能力补强的第二支柱）。
 *
 * 对 table_extract 产出的结构化表做确定性复算：同比变动、占比、合计校验、结构变化排名。
 * 所有派生值给出算式（formula）与输入值，遵循「数字照抄原文、派生值来自脚本」纪律，
 * writer 可直接引用 analysis 结论而不心算。
 */
package com.agenthub.w06.checker.server;

import com.agenthub.w06.contract.DataTable;
import com.agenthub.w06.contract.TableRow;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.*;

final class DataAnalysis {

    /** 解析原文数值字符串（"635,126"→635126.0；"(2,279)"→-2279.0；"10.2%"→10.2）。 */
    static Double num(String s) {
        if (s == null) return null;
        String t = s.trim().replace(",", "");
        boolean neg = t.startsWith("(") && t.endsWith(")");
        if (neg) t = t.substring(1, t.length() - 1);
        t = t.replace("%", "").trim();
        if (t.isEmpty() || t.equals("-")) return null;
        try {
            double v = Double.parseDouble(t);
            return neg ? -v : v;
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private static String fmt(double v) {
        return String.format("%.2f", v);
    }

    /**
     * 分析一张表：
     *  - 若列含 2 期（如 本期/上期 或 2025年/2024年）→ 对每行算同比变动额与增长率；
     *  - 若每行是「金额+占比」结构 → 提取占比列做结构分析；
     *  - 合计校验：若存在"合计/总计"行，检查各列是否等于其上分项之和（容差 0.5%）。
     */
    static ObjectNode analyze(DataTable t) {
        ObjectNode a = CheckerHttpServer.MAPPER.createObjectNode();
        a.put("table_id", t.tableId());
        a.put("rows", t.rows().size());

        ArrayNode yoy = a.putArray("yoy");
        ArrayNode structure = a.putArray("structure");
        ArrayNode checks = a.putArray("checks");

        int cols = t.columns().size();
        // 情形A：列含 4 列「本期/上期/增减额/增长率」——直接复核原文增长率
        if (cols == 4 && t.columns().get(2).contains("增减额")) {
            for (TableRow r : t.rows()) {
                if (r.values().size() < 4) continue;
                Double cur = num(r.values().get(0)), prior = num(r.values().get(1));
                Double dDelta = num(r.values().get(2)), dRate = num(r.values().get(3));
                if (cur == null || prior == null || prior == 0) continue;
                double calcDelta = cur - prior;
                double calcRate = calcDelta / Math.abs(prior) * 100;
                ObjectNode o = yoy.addObject();
                o.put("item", r.item());
                o.put("current", r.values().get(0));
                o.put("prior", r.values().get(1));
                o.put("calc_delta", fmt(calcDelta));
                o.put("calc_growth_pct", fmt(calcRate));
                o.put("source_growth_pct", r.values().get(3));
                boolean ok = dRate == null || Math.abs(calcRate - dRate) < 0.6;
                o.put("reconciled", ok);
                o.put("formula", "(" + cur + "-" + prior + ")/|" + prior + "|*100=" + fmt(calcRate) + "%");
            }
        }
        // 情形B：首两列数值（本期/上期）——计算增长率
        else if (cols >= 2) {
            for (TableRow r : t.rows()) {
                if (r.values().size() < 2) continue;
                Double cur = num(r.values().get(0)), prior = num(r.values().get(1));
                if (cur == null || prior == null || prior == 0) continue;
                double rate = (cur - prior) / Math.abs(prior) * 100;
                ObjectNode o = yoy.addObject();
                o.put("item", r.item());
                o.put("current", r.values().get(0));
                o.put("prior", r.values().get(1));
                o.put("calc_delta", fmt(cur - prior));
                o.put("calc_growth_pct", fmt(rate));
                o.put("formula", "(" + cur + "-" + prior + ")/|" + prior + "|*100=" + fmt(rate) + "%");
            }
        }

        // 结构分析：找占比列（列名含"占比"或值均在 0-100 且有一位小数）
        int pctCol = -1;
        for (int c = 0; c < cols; c++) {
            String cn = t.columns().get(c);
            if (cn.contains("占比") || cn.contains("%")) { pctCol = c; break; }
        }
        if (pctCol >= 0) {
            List<ObjectNode> items = new ArrayList<>();
            for (TableRow r : t.rows()) {
                if (r.values().size() <= pctCol) continue;
                Double p = num(r.values().get(pctCol));
                if (p == null || p <= 0 || p > 100) continue;
                ObjectNode o = CheckerHttpServer.MAPPER.createObjectNode();
                o.put("item", r.item());
                o.put("share_pct", r.values().get(pctCol));
                items.add(o);
            }
            items.sort((x, y) -> Double.compare(
                    num(y.path("share_pct").asText()), num(x.path("share_pct").asText())));
            for (ObjectNode o : items) structure.add(o);
        }

        // 合计校验：末行为合计/总计，逐列核对分项之和
        if (!t.rows().isEmpty()) {
            TableRow last = t.rows().get(t.rows().size() - 1);
            if (last.item().contains("合计") || last.item().contains("总计")) {
                int ncol = last.values().size();
                for (int c = 0; c < ncol; c++) {
                    Double total = num(last.values().get(c));
                    if (total == null) continue;
                    double sum = 0; boolean allNum = true;
                    for (int i = 0; i < t.rows().size() - 1; i++) {
                        Double v = num(t.rows().get(i).values().size() > c ? t.rows().get(i).values().get(c) : null);
                        if (v == null) { allNum = false; break; }
                        sum += v;
                    }
                    ObjectNode o = checks.addObject();
                    o.put("column", c < t.columns().size() ? t.columns().get(c) : ("第" + (c + 1) + "列"));
                    o.put("stated_total", last.values().get(c));
                    o.put("computed_sum", fmt(sum));
                    o.put("applicable", allNum);
                    if (allNum) {
                        double diff = Math.abs(sum - total) / Math.max(Math.abs(total), 1) * 100;
                        o.put("diff_pct", fmt(diff));
                        o.put("reconciled", diff < 0.5);
                    }
                }
            }
        }
        return a;
    }

    static ObjectNode toResultJson(List<DataTable> tables) {
        ObjectNode out = CheckerHttpServer.MAPPER.createObjectNode();
        out.put("ok", true);
        out.put("table_count", tables.size());
        ArrayNode arr = out.putArray("analyses");
        for (DataTable t : tables) arr.add(analyze(t));
        return out;
    }
}
