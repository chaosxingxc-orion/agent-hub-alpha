#!/usr/bin/env python3
"""Exercise the real Go/Java interface without approving tasks or calling a model."""
import json
from pathlib import Path
import time
from demo import CACHE, compose


PROBE = r'''
import json, urllib.request

def post(url, payload, headers=None):
    request = urllib.request.Request(url, data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", **(headers or {})})
    with urllib.request.urlopen(request, timeout=30) as response:
        text = response.read().decode()
        if "text/event-stream" in response.headers.get("Content-Type", ""):
            data = next(line[5:].strip() for line in text.splitlines() if line.startswith("data:"))
            return json.loads(data), dict(response.headers)
        return json.loads(text), dict(response.headers)

for url in ["http://credit:8091/healthz", "http://127.0.0.1:8195/healthz"]:
    with urllib.request.urlopen(url, timeout=5) as response:
        assert response.status == 200

agents = [{"id": "extract", "task": "Extract source facts"},
          {"id": "analyse", "task": "Analyse extracted facts", "dependsOn": ["extract"]}]
compiled, _ = post("http://127.0.0.1:8195/compile", {"mode": "graph", "agents": agents})
assert "error" not in compiled, compiled
assert compiled["steps"][1]["depends_on"] == ["extract"], compiled
again, _ = post("http://127.0.0.1:8195/compile", {"mode": "graph", "agents": agents})
assert again["spec_digest"] == compiled["spec_digest"], (compiled, again)
agents[1]["dependsOn"] = ["missing"]
invalid, _ = post("http://127.0.0.1:8195/compile", {"mode": "graph", "agents": agents})
assert invalid.get("error"), invalid

headers = {"Accept": "application/json, text/event-stream"}
init, response_headers = post("http://credit:8091/mcp", {
    "jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
        "protocolVersion": "2025-06-18", "capabilities": {},
        "clientInfo": {"name": "issue16-deploy-smoke", "version": "1"}}}, headers)
assert "result" in init, init
headers["MCP-Protocol-Version"] = init["result"]["protocolVersion"]
for key, value in response_headers.items():
    if key.lower() == "mcp-session-id": headers["Mcp-Session-Id"] = value
notice = urllib.request.Request("http://credit:8091/mcp", data=json.dumps({
    "jsonrpc": "2.0", "method": "notifications/initialized"}).encode(),
    headers={"Content-Type": "application/json", **headers})
with urllib.request.urlopen(notice, timeout=10) as response: response.read()
tools, _ = post("http://credit:8091/mcp", {"jsonrpc": "2.0", "id": 2, "method": "tools/list"}, headers)
names = [tool["name"] for tool in tools["result"]["tools"]]
assert "list_capabilities" in names and "propose_plan" in names, names
assert not any("approve" in name for name in names), names
catalog, _ = post("http://credit:8091/mcp", {"jsonrpc": "2.0", "id": 3, "method": "tools/call",
    "params": {"name": "list_capabilities", "arguments": {}}}, headers)
assert not catalog["result"].get("isError"), catalog
team = [
    {"id": "intake", "expert": "intake", "role": "资料", "task": "核对模拟材料"},
    {"id": "financial", "expert": "financial-analyst", "role": "财务", "task": "核对财务证据", "depends_on": ["intake"]},
    {"id": "industry", "expert": "industry-researcher", "role": "行业", "task": "核对行业证据", "depends_on": ["intake"]},
    {"id": "bureau", "expert": "bureau-summary", "role": "征信", "task": "核对征信摘要", "depends_on": ["intake"]},
    {"id": "related", "expert": "related-party", "role": "关联方", "task": "核对关联方证据", "depends_on": ["intake"]},
    {"id": "judge", "expert": "risk-judge", "role": "判断", "task": "按规约判断", "depends_on": ["financial", "industry", "bureau", "related"]},
    {"id": "report", "expert": "report-writer", "role": "报告", "task": "汇总报告", "depends_on": ["judge"]},
]
proposal, _ = post("http://credit:8091/mcp", {"jsonrpc": "2.0", "id": 4, "method": "tools/call",
    "params": {"name": "propose_plan", "arguments": {"case_id": "C006", "goal": "部署接口验证，保持待批准", "agents": team}}}, headers)
assert not proposal["result"].get("isError"), proposal
plan = json.loads(proposal["result"]["content"][0]["text"])
assert plan["precheck"]["ok"], plan
assert "run_id" not in plan, plan
print(json.dumps({"health": "passed", "graph_compile": "passed", "stable_graph_digest": "passed",
    "invalid_dependency": "rejected", "planner_tools": names, "capability_lookup": "passed",
    "go_java_proposal": "passed", "unapproved_plan_id": plan["plan_id"],
    "probe_model_calls": 0, "business_tasks_approved_by_probe": 0, "scope": "deployment-interface-only"}))
'''


if __name__ == "__main__":
    result = compose("exec", "-T", "orchestrator", "python", "-c", PROBE, capture=True)
    evidence = json.loads(result.stdout)
    evidence["checked_at_unix"] = int(time.time())
    target = CACHE / "evidence" / "core-smoke.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(evidence, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps(evidence, ensure_ascii=False, indent=2))
    print(f"Evidence: {target}")
