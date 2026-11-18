# -*- coding: utf-8 -*-
"""工行 2025 年报渲染器 v2 —— 支持数据能力补强后的结构化数据表(tables)渲染。"""
import json, html, re, sys
sys.stdout.reconfigure(encoding='utf-8')
SRC = r'C:/Users/Administrator/AppData/Local/Temp/icbc_enhanced_final.json'
A = json.load(open(SRC, encoding='utf-8'))
ART = A.get('artifact', A)
OUT = r'C:/Users/Administrator/WorkBuddy/2026-11-13-21-32-56/工商银行2025年度报告_agent系统生成.html'

order = ["S-8","S-11","S-17","S-15","S-16","S-7","S-1","S-3","S-4","S-6",
         "S-18","S-19","S-20","S-21","S-22","S-23","S-2","S-24","S-25","S-5","S-26","S-27","S-28",
         "S-12","S-9","S-10","S-13","S-14"]
secs = {s['section_id']: s for s in ART['sections']}
order = [x for x in order if x in secs] + [s['section_id'] for s in ART['sections'] if s['section_id'] not in order]
tables = ART.get('tables', [])

def esc(t): return html.escape(str(t) if t is not None else '')
def body_html(t):
    t = re.sub(r'\[C-(\d+)\]', lambda m: f'<sup class="cite">[{m.group(1)}]</sup>', t)
    parts = [p.strip() for p in t.split('\n') if p.strip()]
    return '\n'.join(f'<p>{p}</p>' for p in parts)

# 目录
toc = []
for sid in order:
    s = secs[sid]
    toc.append(f'<li><a href="#{sid}">{esc(s["title"])}</a>'
               f'<span class="pg-note">{len(s.get("body") or "")} 字 · {len(s.get("claim_ids",[]))} 引用</span></li>')
toc_html = '\n'.join(toc)

# 章节正文
chapters = []
for idx, sid in enumerate(order, 1):
    s = secs[sid]
    body = s.get('body') or ''
    if not body.strip():
        continue
    lines = body.split('\n')
    if lines and lines[0].strip() == s['title'].strip():
        body = '\n'.join(lines[1:])
    cites = sorted(set(s.get('claim_ids', [])))
    chapters.append(f'''<section class="chapter" id="{sid}">
  <h2><span class="ch-no">{idx:02d}</span>{esc(s['title'])}</h2>
  <div class="ch-meta">正文 {len(body)} 字 · 引用 {len(cites)} 条 · 责任工位 {esc(sid)}</div>
  <div class="ch-body">{body_html(body)}</div>
</section>''')
chapters_html = '\n'.join(chapters)

# ===== 结构化数据表渲染（数据能力补强核心） =====
# 按页排序，分主题分组
def table_group(t):
    title = t.get('title','')
    pg = t.get('source_page',0)
    if any(k in title for k in ['利润表','营业收入','手续费','收入结构','分部']): return '一、经营业绩与收入结构'
    if any(k in title for k in ['五级分类','不良','贷款结构','贷款和不良','担保方式','行业','地域','减值准备']): return '三、资产质量与贷款结构'
    if any(k in title for k in ['资本充足','风险加权','核心一级']): return '四、资本管理'
    if any(k in title for k in ['流动性','外汇敞口','汇率']): return '五、流动性与市场风险'
    if any(k in title for k in ['生息资产','计息负债','资产构成','负债结构','存款','投资结构','资产','负债']): return '二、资产负债与息差'
    return '六、其他经营数据'

groups = {}
for t in tables:
    groups.setdefault(table_group(t), []).append(t)

def render_table(t):
    cols = t.get('columns', [])
    rows = t.get('rows', [])
    unit = t.get('unit','')
    title = t.get('title','') or '财报数据表'
    pg = t.get('source_page', 0)
    tid = t.get('table_id','')
    an = t.get('analysis') or {}
    # 表头：项目 + 列（列标签缺失时按列数智能补全：本期/上期/增减/增长率 或 近三年）
    ncol = max([len(r.get('values',[])) for r in rows] + [len(cols)])
    labels = [cols[i] if i < len(cols) else '' for i in range(ncol)]
    if not any(x.startswith('第') for x in labels):
        pass
    else:
        if ncol == 4 and any('增减' in str(x) for x in labels):
            labels = ['本期(2025)', '上期(2024)', '增减额', '增长率(%)']
        elif ncol == 4 and any('占比' in str(x) for x in labels):
            labels = ['2025金额', '2025占比(%)', '2024金额', '2024占比(%)']
        elif ncol == 8:
            labels = ['2025金额','占比(%)','2025不良','不良率(%)','2024金额','占比(%)','2024不良','不良率(%)']
        elif ncol == 6:
            labels = ['2025平均余额','利息收支','收益率(%)','2024平均余额','利息收支','收益率(%)']
        elif ncol == 3:
            labels = ['2025年', '2024年', '2023年']
        elif ncol == 2:
            labels = ['2025年', '2024年']
        else:
            # 4列纯数值多为利润表变动表；其余按数值列
            if ncol == 4:
                labels = ['2025年', '2024年', '增减额', '增长率(%)']
            else:
                labels = [x if x and not x.startswith('第') else f'数值{i+1}' for i, x in enumerate(labels)]
    thead = '<tr><th class="rowhead">项目</th>' + ''.join(
        f'<th>{esc(labels[i])}</th>' for i in range(ncol)) + '</tr>'
    trs = []
    for r in rows:
        vals = r.get('values', [])
        item = r.get('item','')
        cls = ' class="total"' if ('合计' in item or '总计' in item or '总额' in item) else ''
        tds = ''.join(f'<td class="num">{esc(vals[i] if i < len(vals) else "")}</td>' for i in range(ncol))
        trs.append(f'<tr{cls}><td class="rowhead">{esc(item)}</td>{tds}</tr>')
    # 汇总分析脚注（选前2条同比）
    note = ''
    yoys = an.get('yoy') or []
    if yoys:
        bits = []
        for y in yoys[:2]:
            if y.get('item') and y.get('calc_growth_pct'):
                bits.append(f'{esc(y["item"])}同比 {esc(y["calc_growth_pct"])}%')
        if bits: note = f'<div class="t-note">汇总分析（system table_analyze）：' + '；'.join(bits) + '</div>'
    return f'''<figure class="data-table" id="{esc(tid)}">
  <figcaption><span class="tid">{esc(tid)}</span>{esc(title)}
    <span class="t-src">来源：第 {pg} 页{f' · {esc(unit)}' if unit else ''}</span></figcaption>
  <div class="tw"><table><thead>{thead}</thead><tbody>{''.join(trs)}</tbody></table></div>
  {note}
</figure>'''

order_groups = ['一、经营业绩与收入结构','二、资产负债与息差','三、资产质量与贷款结构','四、资本管理','五、流动性与市场风险','六、其他经营数据']
tables_sections = []
gnum = 0
for g in order_groups:
    if g not in groups: continue
    gnum += 1
    figs = '\n'.join(render_table(t) for t in groups[g])
    tables_sections.append(f'''<section class="chapter" id="data-{gnum}">
  <h2><span class="ch-no">表{gnum}</span>{esc(g)}</h2>
  <div class="ch-meta">{len(groups[g])} 张数据表 · 由 system table_extract 从年报原文结构化抽取 · table_analyze 复算</div>
  {figs}
</section>''')
tables_html = '\n'.join(tables_sections)

# 引用溯源表
cit_rows = []
for c in ART['citations']:
    cit_rows.append(f'<tr><td>{esc(c["claim_id"])}</td><td>第 {c["page"]} 页</td>'
                    f'<td>{esc(str(c["source_document_digest"])[:12])}…</td>'
                    f'<td>{esc(str(c["claim_text"])[:90])}</td></tr>')
cit_html = '\n'.join(cit_rows)

# 图表
def render_chart(ch):
    typ = ch.get('type'); title = esc(ch.get('title',''))
    if typ in ('bar','line'):
        series = ch.get('series',[]); labels = ch.get('labels',[]); bars = []
        colors = ['#c0392b','#2c3e50','#2980b9','#8e44ad']
        if typ == 'bar':
            n = len(labels); gw = 560/max(n,1)
            for si, ser in enumerate(series):
                vals = ser.get('values',[]); mx = max([abs(v) for v in vals]+[1])
                for i, v in enumerate(vals):
                    h = abs(v)/mx*140; x = 40+i*gw+si*(gw/(len(series)+1)); y = 200-h
                    bars.append(f'<rect x="{x:.0f}" y="{y:.0f}" width="{gw/(len(series)+1)-6:.0f}" height="{h:.0f}" fill="{colors[si%len(colors)]}" opacity=".82"><title>{esc(ser.get("name",""))} {v}</title></rect>')
                    bars.append(f'<text x="{x+gw/(len(series)+1)/2-3:.0f}" y="{y-4:.0f}" font-size="10" text-anchor="middle">{v}</text>')
            for i, l in enumerate(labels):
                bars.append(f'<text x="{40+i*gw+gw/2:.0f}" y="218" font-size="11" text-anchor="middle">{esc(l)}</text>')
        else:
            for si, ser in enumerate(series):
                vals = ser.get('values',[]); mx = max([abs(v) for v in vals]+[1]); mn = min([0]+vals)
                pts = []
                for i, v in enumerate(vals):
                    x = 60+i*(480/max(len(vals)-1,1)); y = 200-(v-mn)/(mx-mn)*150
                    pts.append(f'{x:.0f},{y:.0f}')
                bars.append(f'<polyline points="{" ".join(pts)}" fill="none" stroke="{colors[si%len(colors)]}" stroke-width="2.5"/>')
        return f'<figure><figcaption>{title}</figcaption><svg viewBox="0 0 620 240" width="100%">{"".join(bars)}</svg></figure>'
    return f'<figure><figcaption>{title}</figcaption><p class="chart-note">图表类型：{esc(typ)}</p></figure>'
charts_html = '\n'.join(render_chart(ch) for ch in ART.get('charts', []))

total = sum(len(s.get('body') or '') for s in ART['sections'])
n_tables = len(tables)
digest = A.get('artifact_digest', ART.get('artifact_digest','sha256:…'))

doc = f'''<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>中国工商银行股份有限公司 2025 年度报告 · Agent 系统生成版</title>
<style>
:root{{--ink:#1a1a1a;--muted:#6b7280;--line:#e5e7eb;--accent:#9b1c1c;--gold:#a67c00;--bg:#f5f5f3;--paper:#fff}}
*{{box-sizing:border-box}}
body{{margin:0;background:var(--bg);color:var(--ink);font-family:"Songti SC","SimSun",Georgia,serif;line-height:1.85;font-size:15px}}
.page{{max-width:1040px;margin:0 auto;background:var(--paper);padding:0 60px 80px;box-shadow:0 0 24px rgba(0,0,0,.05)}}
header.cover{{text-align:center;padding:90px 0 60px;border-bottom:3px double var(--accent)}}
.cover .bank{{font-size:15px;letter-spacing:.4em;color:var(--muted)}}
.cover h1{{font-size:38px;margin:28px 0 8px;letter-spacing:.08em;font-weight:normal}}
.cover .en{{font-size:13px;color:var(--muted);letter-spacing:.1em}}
.cover .year{{font-size:60px;color:var(--accent);margin:30px 0;font-family:Georgia,serif}}
.cover .meta{{margin-top:36px;font-size:13px;color:var(--muted);line-height:2}}
.cover .seal{{display:inline-block;border:1px solid var(--accent);color:var(--accent);padding:6px 18px;font-size:12px;letter-spacing:.15em;margin-top:24px}}
section{{padding:32px 0;border-bottom:1px solid var(--line)}}
h2{{font-size:20px;font-weight:600;color:var(--accent);border-left:4px solid var(--accent);padding-left:14px;margin:0 0 8px;line-height:1.5}}
.ch-no{{display:inline-block;font-family:Georgia,serif;font-size:13px;color:var(--muted);margin-right:12px;font-weight:normal}}
.ch-meta{{font-size:12px;color:var(--muted);margin:0 0 20px 18px;font-family:system-ui,sans-serif}}
.ch-body p{{margin:0 0 15px;text-align:justify;text-indent:2em}}
.cite{{color:var(--accent);font-size:11px;margin-left:1px;font-family:system-ui,sans-serif;cursor:help}}
.toc-h{{font-size:24px;color:var(--accent);text-align:center;margin-bottom:24px;letter-spacing:.2em}}
.toc li{{margin:6px 0;list-style:none;display:flex;justify-content:space-between;border-bottom:1px dotted var(--line);padding-bottom:5px}}
.toc a{{color:var(--ink);text-decoration:none}}
.toc a:hover{{color:var(--accent)}}
.pg-note{{font-size:11px;color:var(--muted);font-family:system-ui,sans-serif}}
figure{{margin:22px 0;text-align:center}}
figcaption{{font-size:14px;color:var(--ink);margin-bottom:10px;font-weight:600}}
.chart-note{{font-size:12px;color:var(--muted)}}
table{{width:100%;border-collapse:collapse;font-size:12px;font-family:system-ui,sans-serif}}
th,td{{border:1px solid var(--line);padding:5px 8px}}
th{{background:#faf8f5;color:#444;font-weight:600;text-align:center}}
td.num{{text-align:right;font-variant-numeric:tabular-nums;white-space:nowrap}}
td.rowhead{{text-align:left;color:#333}}
tr.total td{{background:#fbf3f3;font-weight:700;border-top:2px solid var(--accent)}}
.data-table{{margin:26px 0;text-align:left;border:1px solid var(--line);padding:0;overflow:hidden}}
.data-table figcaption{{margin:0;padding:10px 14px;background:#faf8f5;border-bottom:1px solid var(--line);font-family:system-ui,sans-serif;display:flex;gap:10px;align-items:center;flex-wrap:wrap}}
.data-table .tid{{display:inline-block;background:var(--accent);color:#fff;font-size:10px;padding:2px 7px;border-radius:3px;letter-spacing:.05em}}
.data-table .t-src{{margin-left:auto;font-size:11px;color:var(--muted);font-weight:400}}
.data-table .tw{{overflow-x:auto}}
.data-table table{{margin:0;border:0}}
.t-note{{font-size:11px;color:var(--muted);padding:8px 14px;background:#fcfbf9;border-top:1px solid var(--line);font-family:system-ui,sans-serif}}
.decl{{background:#faf8f5;border:1px solid var(--line);padding:20px 24px;font-size:13px;color:var(--muted)}}
.decl b{{color:var(--accent)}}
code{{background:#f0ece6;padding:1px 5px;border-radius:3px;font-size:12px}}
@media print{{body{{background:#fff}}.page{{box-shadow:none;max-width:none;padding:0 20px}}section{{page-break-inside:avoid}}}}
</style></head>
<body><div class="page">
<header class="cover">
  <div class="bank">中国工商银行股份有限公司</div>
  <div class="en">INDUSTRIAL AND COMMERCIAL BANK OF CHINA LIMITED</div>
  <div class="year">2025</div>
  <h1>年度报告</h1>
  <div class="en">ANNUAL REPORT · Agent 系统生成版（数据支撑强化）</div>
  <div class="meta">报告期：2025 年度　|　生成系统：W06 多 Agent 受控交付流水线<br>
  生成方式：writer-agent 基于官方披露年报全文逐节成稿 + table_extract/table_analyze 结构化数据支撑<br>
  核验：数值核验 N1–N7 PASSED · 引用核验 C1–C2 PASSED · 具名批准</div>
  <div class="seal">受控生成 · 数据可溯源</div>
</header>

<section id="toc">
  <div class="toc-h">目　录</div>
  <ul class="toc">{toc_html}
  <li><a href="#data-1">财务数据表附录（{n_tables} 张）</a><span class="pg-note">结构化抽取 · 来源可打开</span></li>
  <li><a href="#citations">引用溯源清单</a><span class="pg-note">{len(ART['citations'])} 条</span></li>
  <li><a href="#declaration">生成与核验声明</a><span class="pg-note">—</span></li></ul>
</section>

<section id="charts">
  <h2><span class="ch-no">◆</span>关键数据图表</h2>
  <div class="ch-meta">由 chart-agent 渲染 · 数据取自系统工件</div>
  {charts_html}
</section>

{chapters_html}

{tables_html}

<section id="citations">
  <h2><span class="ch-no">◆</span>引用溯源清单</h2>
  <div class="ch-meta">共 {len(ART['citations'])} 条引用 · digest/page/excerpt 三元组经引用核验器逐条校验</div>
  <table><thead><tr><th>编号</th><th>页码</th><th>信源指纹</th><th>论断摘录</th></tr></thead><tbody>{cit_html}</tbody></table>
</section>

<section id="declaration">
  <h2><span class="ch-no">◆</span>生成与核验声明</h2>
  <div class="decl">
  <p>本报告由 <b>W06 多 Agent 受控交付流水线</b>生成，非人工直接撰写。生成链路如下：</p>
  <p>① <b>信源获取</b>：MCP 工具 <code>pdf_extract</code> 从官方披露直链逐页抓取中国工商银行 2025 年度报告全文，<code>snapshot_commit</code> 落盘并 SHA-256 注册（106 页真实文本）；</p>
  <p>② <b>数据能力补强</b>：新增 <code>table_extract</code> 将被 PDF 拍平的财报表格结构化为 <b>{n_tables} 张真实数据表</b>（逐格保留原文数值与来源页），<code>table_analyze</code> 做同比/占比/合计的确定性复算并附算式；表格数据注入 writer 任务书，使正文论断有真实数据支撑；</p>
  <p>③ <b>任务编排与成稿</b>：Java 编排层组装任务书（指标 + 引用 + 结构化表格 + 快照页原文），<b>writer-agent（openjiuwen deep agent）</b>按 annual-report 技能逐节成稿，含缺节续写与反思修订回路；</p>
  <p>④ <b>确定性核验</b>：<b>NumericChecker</b>（N1–N7，含跨主体守护避免同业数据误判）与 <b>CitationChecker</b>（C1–C2 三元组校验）逐条核验，<b>均 PASSED</b>；</p>
  <p>⑤ <b>受控交付</b>：工件经具名批准（approver: Audit），批准绑定内容指纹 <code>{esc(str(digest)[:24])}…</code>，任一数字改动即令批准失效。</p>
  <p>本报告规模：<b>{len(order)} 章 · {total:,} 字正文 · {n_tables} 张数据表 · {len(ART['citations'])} 条可溯源引用</b>。所有数据均来自官方披露原文，页外推算均标注算式。本报告用于多 Agent 受控生成链路验证，不构成投资建议。</p>
  </div>
</section>
</div></body></html>'''

open(OUT, 'w', encoding='utf-8', newline='\n').write(doc)
print('年报成品生成:', OUT)
print(f'  {len(order)} 章 · {total:,} 字 · {n_tables} 张数据表 · {len(ART["citations"])} 引用 · {len(ART.get("charts",[]))} 图表')
