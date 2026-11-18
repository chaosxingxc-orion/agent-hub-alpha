/*
 * TableExtractor —— 财报表格结构化抽取（数据能力补强的核心）。
 *
 * 背景：pdf_extract 把年报表格拍平成纯文本（项目名 + 若干数字列），writer 只能堆文字。
 * 本工具把拍平文本还原成「表 + 行 + 单元格」结构，逐格保留原文字符串（千分位、括号负数、%），
 * 使报告可以用真实数据表支撑论断。算法为确定性规则，无 LLM、不臆造数值。
 */
package com.agenthub.w06.checker.server;

import com.agenthub.w06.contract.DataTable;
import com.agenthub.w06.contract.TableRow;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class TableExtractor {

    private static final Pattern NUM = Pattern.compile(
            "^[\\(（]?\\s*-?[\\d,]+(?:\\.\\d+)?\\s*[\\)）]?%?$");
    private static final Pattern YEAR = Pattern.compile("(?:19|20)\\d{2}");
    private static final Pattern UNIT = Pattern.compile("^人民币(百万|亿|万)元");

    private static boolean isNum(String tok) {
        if (tok == null || tok.isEmpty()) return false;
        Matcher m = NUM.matcher(tok);
        return m.matches() && tok.chars().anyMatch(Character::isDigit);
    }

    private static int numCount(String line) {
        int c = 0;
        for (String t : line.split("\\s+")) if (isNum(t)) c++;
        return c;
    }

    /** 从单页文本抽取所有结构化表（返回契约对象列表）。 */
    static List<DataTable> extract(String pageText, int page, String digest,
                                   List<String> canonicalTitles) {
        List<Map<String, Object>> raw = parse(pageText);
        List<DataTable> out = new ArrayList<>();
        for (int i = 0; i < raw.size(); i++) {
            @SuppressWarnings("unchecked")
            Map<String, Object> m = raw.get(i);
            @SuppressWarnings("unchecked")
            List<Map<String, Object>> rows = (List<Map<String, Object>>) m.get("rows");
            String title = (String) m.get("title");
            if (title == null || title.isBlank() && i < canonicalTitles.size()) {
                title = canonicalTitles.get(i);
            }
            @SuppressWarnings("unchecked")
            List<String> cols = (List<String>) m.get("columns");
            List<TableRow> trows = new ArrayList<>();
            for (Map<String, Object> r : rows) {
                @SuppressWarnings("unchecked")
                List<String> vals = (List<String>) r.get("values");
                trows.add(new TableRow((String) r.get("item"), vals));
            }
            String id = "T-" + page + "-" + (i + 1);
            out.add(new DataTable(id, title == null ? "" : title, (String) m.get("unit"),
                    cols, trows, page, digest, null));
        }
        return out;
    }

    /** 解析算法（与 Python 原型一致）：表头行→列标签，数据行→尾部数字序列。 */
    private static List<Map<String, Object>> parse(String text) {
        String[] lines = text.split("\\R", -1);
        List<Map<String, Object>> tables = new ArrayList<>();
        Map<String, Object> cur = null;
        int curCols = -1;
        String pendingTitle = "", pendingUnit = "";
        List<String> pendingCols = new ArrayList<>();

        for (String rawLine : lines) {
            String line = rawLine.trim();
            if (line.isEmpty()) continue;

            // 单位行
            if (UNIT.matcher(line).find() || (line.contains("百分比除外") && line.length() < 30)) {
                pendingUnit = line;
                continue;
            }

            // 候选标题行：纯中文短语、短、非数据、非表头
            if (numCount(line) == 0 && line.length() < 40
                    && line.chars().anyMatch(c -> c >= 0x4e00 && c <= 0x9fff)
                    && !line.contains("年")) {
                pendingTitle = line;
            }

            // 表头行：含"项目"/"占比"/"增长率" 或 2+年份，且数字 token 很少
            Matcher ym = YEAR.matcher(line);
            int years = 0;
            Set<String> yset = new LinkedHashSet<>();
            while (ym.find()) { years++; yset.add(ym.group()); }
            boolean headWord = line.contains("项目")
                    || (line.contains("占比") && line.contains("%"))
                    || (line.contains("增长率") && line.contains("%"));
            String[] toks = line.split("\\s+");
            boolean lastIsNum = toks.length > 0 && isNum(toks[toks.length - 1]);
            if ((headWord || years >= 2) && numCount(line) <= 2 && line.length() < 120 && !lastIsNum) {
                if (years >= 2) {
                    pendingCols = new ArrayList<>();
                    for (String y : yset) pendingCols.add(y + "年");
                } else if (line.contains("增减额") || line.contains("增长率")) {
                    pendingCols = new ArrayList<>(List.of("本期", "上期", "增减额", "增长率(%)"));
                } else if (line.contains("占比")) {
                    pendingCols = new ArrayList<>(List.of("金额", "占比(%)", "金额", "占比(%)"));
                } else {
                    pendingCols = new ArrayList<>();
                }
                if (cur != null) { tables.add(cur); cur = null; }
                continue;
            }

            // 数据行
            if (toks.length < 2) continue;
            List<String> nums = new ArrayList<>();
            for (int k = toks.length - 1; k >= 0; k--) {
                if (isNum(toks[k])) nums.add(toks[k]);
                else break;
            }
            Collections.reverse(nums);
            if (nums.size() < 2) continue;
            StringBuilder name = new StringBuilder();
            for (int k = 0; k < toks.length - nums.size(); k++) {
                if (name.length() > 0) name.append(' ');
                name.append(toks[k]);
            }
            String item = name.toString().trim();
            if (item.isEmpty() || !item.chars().anyMatch(c -> Character.isLetter(c))) continue;

            Map<String, Object> row = new LinkedHashMap<>();
            row.put("item", item);
            row.put("values", nums);
            if (cur == null || curCols != nums.size()) {
                if (cur != null) tables.add(cur);
                curCols = nums.size();
                cur = new LinkedHashMap<>();
                cur.put("title", pendingTitle);
                cur.put("unit", pendingUnit);
                cur.put("columns", pendingCols.size() == nums.size()
                        ? new ArrayList<>(pendingCols)
                        : defaultCols(nums.size()));
                cur.put("rows", new ArrayList<Map<String, Object>>());
                pendingTitle = "";
            }
            @SuppressWarnings("unchecked")
            List<Map<String, Object>> rows = (List<Map<String, Object>>) cur.get("rows");
            rows.add(row);
        }
        if (cur != null) tables.add(cur);

        // 过滤行数 < 3 的噪声表
        List<Map<String, Object>> out = new ArrayList<>();
        for (Map<String, Object> t : tables) {
            @SuppressWarnings("unchecked")
            List<Map<String, Object>> rows = (List<Map<String, Object>>) t.get("rows");
            if (rows.size() >= 3) out.add(t);
        }
        return out;
    }

    private static List<String> defaultCols(int n) {
        List<String> c = new ArrayList<>();
        for (int i = 1; i <= n; i++) c.add("第" + i + "列");
        return c;
    }

    /** 把表列表序列化成 /v1/tools/call 返回 content。 */
    static ObjectNode toResultJson(List<DataTable> tables) {
        ObjectNode out = CheckerHttpServer.MAPPER.createObjectNode();
        out.put("ok", true);
        out.put("table_count", tables.size());
        ArrayNode arr = out.putArray("tables");
        for (DataTable t : tables) arr.add(CheckerHttpServer.MAPPER.valueToTree(t));
        return out;
    }
}
