#!/usr/bin/env python3
"""X-S3 共译差分·单臂单发启动器（用户本人执行——真 LLM 纪律）。

用法（钥匙经环境变量注入，不落盘不回显；本役实际执行时两臂统一走 XS3_KEY=volcano 网关钥匙）:
  export XS3_KEY=...

  python3 -I xs3_run.py --arm glm   # 我方弱形态臂（glm-5.3-flash）
  python3 -I xs3_run.py --arm qwen  # 第四臂（qwen3.8-flash）

行为: 单次调用（零上下文）· temperature=0 · max_tokens=16384（推理系思考吃额度，4096 实测截断在案）· 无自动重试。
首译绑定（86号 §四.2 拟定）: 首个完整呈回件=五格行齐全且每行含「格ID｜判定｜依据」；
不完整中间态不算呈回——本器只标记 COMPLETE/INCOMPLETE，是否重跑由你决定，
每次尝试（含失败）都写入 attempt 台账，如实留痕。
"""
import argparse, json, os, sys, time, pathlib

ARMS = {
    "glm":  dict(endpoint="https://st8tp3ajl0df3n8b8l8qu.apigateway-cn-beijing.volceapi.com/compatible/chat/completions",
                 env="XS3_KEY", model="glm-5.3-flash",
                 version_pin="GLM-5.3-Flash 上线 2026-08-26 ⇒ 训练截止≤2026-08-26 < 2026-11-10 08:12 ✓（GLM 家族=我方弱形态臂）；端点=volcano 网关（用户钥匙基建；86号「bigmodel 网关」措辞更正；56pp 代理前科披露 XM2 档案）"),
    "qwen": dict(endpoint="https://st8tp3ajl0df3n8b8l8qu.apigateway-cn-beijing.volceapi.com/compatible/chat/completions",
                 env="XS3_KEY", model="qwen3.8-flash",
                 version_pin="qwen3.8-flash 产品线至迟 2026-09-19 在馆（我方 volcano 三闪矩阵实测档案）⇒ 截止≤2026-09-19 ✓；同 id 静默更新盲区披露（家族表保守口径）；qwen 家族跨家族 ✓（vs GLM 臂/vs 夏方 deepseek 臂）；同网关相关性披露"),
}
HERE = pathlib.Path(__file__).parent
OUT = HERE / "results"
PROMPT_FILE = HERE / "arm-input.md"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--arm", choices=ARMS, required=True)
    ap.add_argument("--prompt", default=None, help="输入件路径（默认 arm-input.md）；扰动臂用补正层冻结的对照/扰动件")
    a = ap.parse_args()
    conf = ARMS[a.arm]
    key = os.environ.get(conf["env"])
    if not key:
        sys.exit(f"缺 {conf['env']}——先在你的 shell 里 export（钥匙不经任何文件）")
    prompt_path = PROMPT_FILE if a.prompt is None else __import__("pathlib").Path(a.prompt)
    prompt = prompt_path.read_text(encoding="utf-8")
    body = {
        "model": conf["model"],
        "temperature": 0,
        "max_tokens": 32768,  # 扰动臂六格材料思考更长，16384 实测吃满（attempt-221410 在案）
        "messages": [{"role": "user", "content": prompt}],
    }
    import urllib.request
    req = urllib.request.Request(
        conf["endpoint"], data=json.dumps(body).encode(),
        headers={"Authorization": f"Bearer {key}", "Content-Type": "application/json"})
    t0 = time.time()
    try:
        with urllib.request.urlopen(req, timeout=600) as r:
            raw = json.loads(r.read())
    except Exception as e:
        log_attempt(a.arm, conf, None, f"FAILED: {type(e).__name__}: {e}", t0)
        sys.exit(f"调用失败（已入 attempt 台账，不算呈回）: {e}")
    # 双形解析: OpenAI(choices.message) 或 Anthropic(content blocks)——volcano /compatible 系 Claude 兼容层
    if not isinstance(raw, dict):
        # R2 m2 (landed 2026-11-13 inner-loop IN-1 — the first attempt's
        # replace anchor silently missed, the commit message claimed the
        # defense while the code never landed): a non-dict response body
        # used to crash AFTER transport success with NO ledger row — an
        # attempt that violates the every-attempt-logged claim.
        log_attempt(a.arm, conf, None, f"FAILED: non-dict response body: {type(raw).__name__}", t0)
        sys.exit(f"响应体非 dict（已入 attempt 台账，不算呈回）: {type(raw).__name__}")
    if raw.get("choices"):
        msg = (raw.get("choices") or [{}])[0].get("message", {})
        content = msg.get("content", "") or ""
        reasoning = msg.get("reasoning_content", "") or ""
    else:
        blocks = raw.get("content") or []
        content = "".join(b.get("text", "") for b in blocks if b.get("type") == "text")
        reasoning = "".join(b.get("thinking", "") for b in blocks if b.get("type") == "thinking")
    complete = validate(content, cells=cells_from_prompt(prompt))
    log_attempt(a.arm, conf, raw, "COMPLETE" if complete else "INCOMPLETE", t0,
                content=content, reasoning_len=len(reasoning))
    print(f"arm={a.arm} model={conf['model']} -> {'COMPLETE' if complete else 'INCOMPLETE'}")
    print(f"判定行:\n{content}")


import re as _re

def validate(content: str, cells: str = None) -> bool:
    # R2 m3 tightened the row shape (cell ID anchored at line start, non-empty
    # verdict between the bars). The cell COUNT was then hardcoded — ==5 for
    # the five-cell originals, bumped to ==6 for the perturbation batch —
    # and the hardcode false-REDs any reply whose material carries a
    # different cell count (the five-cell first translations failed under
    # the six-cell constant; full-test regression face, 2026-11-13).
    # Structural fix (消除优于处理): the expected cell set is DERIVED from
    # the prompt's grid table — no count lives in this function at all.
    if cells is None:
        cells = "ABCDEF"
    ok = 0
    for cell in cells:
        for l in content.splitlines():
            m = _re.match(r"\s*" + cell + r"\s*[｜|]\s*([^｜|\n]+?)\s*[｜|]", l)
            if m and m.group(1).strip():
                ok += 1
                break
    return ok == len(cells)


def cells_from_prompt(prompt_text: str) -> str:
    """Derive the expected cell IDs from the prompt's grid table rows."""
    found = []
    for l in prompt_text.splitlines():
        m = _re.match(r"\|\s*([A-Z])\s*\|", l)
        if m and m.group(1) not in found:
            found.append(m.group(1))
    return "".join(found)


def log_attempt(arm, conf, raw, status, t0, content="", reasoning_len=0):
    OUT.mkdir(exist_ok=True)
    ts = time.strftime("%Y%m%d-%H%M%S")
    rec = {
        "ts": ts, "arm": arm, "model": conf["model"], "version_pin": conf["version_pin"],
        "status": status, "elapsed_s": round(time.time() - t0, 1),
        "reasoning_len": reasoning_len,
        "usage": (raw or {}).get("usage"),
    }
    (OUT / f"attempt-{ts}-{arm}.json").write_text(
        json.dumps(rec, ensure_ascii=False, indent=1), encoding="utf-8")
    if raw is not None:
        (OUT / f"raw-{ts}-{arm}.json").write_text(json.dumps(raw, ensure_ascii=False), encoding="utf-8")
    if content:
        (OUT / f"reply-{ts}-{arm}.md").write_text(content, encoding="utf-8")


if __name__ == "__main__":
    main()
