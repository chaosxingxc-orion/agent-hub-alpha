#!/usr/bin/env python3
"""Agent Bus acceptance through the real MCP, approval, Graph and A2A path.

Uses the configured model and approves synthetic credit-demo test tasks.
No live lending operation is included in the test graph.
"""
from __future__ import annotations
import argparse
import http.cookiejar
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.parse
import urllib.error
import urllib.request

from demo import CACHE

BASE = os.environ.get("BUS_TEST_BASE", "http://127.0.0.1:5191")
PROJECT = os.environ.get("BUS_TEST_PROJECT", "agent-hub-issue16")


def environment_record():
    root = Path(__file__).resolve().parents[5]
    commit = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    ids = subprocess.check_output(["docker", "ps", "-aq", "--filter", "label=com.docker.compose.project=" + PROJECT], text=True).split()
    containers = json.loads(subprocess.check_output(["docker", "inspect", *ids], text=True)) if ids else []
    return {"project": PROJECT, "base": BASE, "source_commit": commit,
            "images": {x["Name"]: x["Image"] for x in containers},
            "config_sha256": {name: hashlib.sha256((Path(__file__).parent / name).read_bytes()).hexdigest()
                              for name in ("compose.yaml", "compose.bus.yaml", "compose.bus-test.yaml", "bus/rdc.yml", "sources.lock.json")}}
TEAM = [
    {"id": "intake", "expert": "intake", "role": "资料", "task": "使用 get_application 核对本案件申请材料"},
    {"id": "financial", "expert": "financial-analyst", "role": "财务", "task": "使用 get_financials 分析本案件财务数据，明确资产单位", "depends_on": ["intake"]},
    {"id": "industry", "expert": "industry-researcher", "role": "行业", "task": "使用 get_industry_report 分析本案件行业依据", "depends_on": ["intake"]},
    {"id": "bureau", "expert": "bureau-summary", "role": "征信", "task": "使用 query_bureau_summary 核对本案件征信摘要", "depends_on": ["intake"]},
    {"id": "related", "expert": "related-party", "role": "关联方", "task": "逐家核对本案件关联方证据", "depends_on": ["intake"]},
    {"id": "judge", "expert": "risk-judge", "role": "原型规则判断", "task": "按原型规约判断并引用上游步骤", "depends_on": ["financial", "industry", "bureau", "related"]},
    {"id": "report", "expert": "report-writer", "role": "报告", "task": "汇总已提供的结论和证据，缺失内容明确说明", "depends_on": ["judge"]},
]


def request(path, payload=None, headers=None):
    req = urllib.request.Request(BASE + path, data=None if payload is None else json.dumps(payload).encode(),
                                 headers={"Content-Type": "application/json", **(headers or {})})
    with urllib.request.urlopen(req, timeout=45) as response:
        body = response.read().decode()
        if "text/event-stream" in response.headers.get("Content-Type", ""):
            body = next(line[5:].strip() for line in body.splitlines() if line.startswith("data:"))
        return (json.loads(body) if body else None), dict(response.headers)


def redact(value):
    if isinstance(value, dict):
        return {k: redact(v) for k, v in value.items() if k.lower() not in ("token", "capability_token", "authorization", "password", "routehandle")}
    if isinstance(value, list):
        return [redact(x) for x in value]
    if isinstance(value, str) and value.lstrip().startswith("{"):
        try:
            return json.dumps(redact(json.loads(value)), ensure_ascii=False)
        except ValueError:
            pass
    return value


class Planner:
    def __init__(self):
        self.headers = {"Accept": "application/json, text/event-stream"}
        init, headers = request("/mcp", {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
            "protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "bus-acceptance", "version": "1"}}}, self.headers)
        self.headers["MCP-Protocol-Version"] = init["result"]["protocolVersion"]
        for key, value in headers.items():
            if key.lower() == "mcp-session-id":
                self.headers["Mcp-Session-Id"] = value
        request("/mcp", {"jsonrpc": "2.0", "method": "notifications/initialized"}, self.headers)
        self.sequence = 1

    def tool(self, name, arguments):
        self.sequence += 1
        result, _ = request("/mcp", {"jsonrpc": "2.0", "id": self.sequence, "method": "tools/call",
                                     "params": {"name": name, "arguments": arguments}}, self.headers)
        if "error" in result or result.get("result", {}).get("isError"):
            raise RuntimeError(str(redact(result))[:1000])
        return json.loads(result["result"]["content"][0]["text"])


def new_run(label):
    planner = Planner()
    catalog = planner.tool("list_capabilities", {})
    remote = [c for c in catalog["experts"] if c.get("remote")]
    assert len({c["bus_binding"]["agent_id"] for c in remote if c.get("available")}) == 3, "Three real roles must be discoverable"
    plan = planner.tool("propose_plan", {"case_id": "C006", "goal": f"Agent Bus {label} 验证 {int(time.time())}，仅生成模拟授信报告，不提交银行操作", "agents": TEAM})
    assert plan["precheck"]["ok"], "Candidate graph failed precheck"
    graph, _ = request(f'/api/graph?plan={plan["plan_id"]}&v={plan["version"]}')
    approved, _ = request("/api/approve", {"plan": plan["plan_id"], "version": plan["version"], "hash": graph["hash"],
                                          "user": "alice", "password": os.environ.get("CREDIT_DEMO_PASSWORD", "demo")})
    return approved["run"], plan["plan_id"]


def read_run(run_id):
    state, _ = request("/debug/state")
    return state["runs"][run_id]


def docker(service, *args):
    name = f"{PROJECT}-{service}-1"
    return subprocess.run(["docker", args[0], name, *args[1:]], check=True, capture_output=True, text=True).stdout


def resume(run_id):
    # Exercise the same authenticated human action exposed by the existing UI.
    cookies = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookies))
    payload = urllib.parse.urlencode({"user": "alice", "password": os.environ.get("CREDIT_DEMO_PASSWORD", "demo"), "next": "/ui/"}).encode()
    # Redirects use the /credit prefix of the front proxy, so retain cookies even
    # when following this redirect directly on the control port returns 404.
    try:
        opener.open(BASE + "/ui/login", payload, timeout=10).close()
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise
    assert any(c.name == "cd_session" for c in cookies), "Login did not issue a session"
    try:
        opener.open(BASE + f"/ui/run/{run_id}/resume", b"", timeout=10).close()
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise


def call_snapshot(run):
    return {key: {k: c.get(k) for k in ("request_id", "remote_task_id", "artifact_id", "artifact_digest", "model_events", "context_id", "status")}
            for key, step in run["steps"].items() if (c := step.get("remote_call"))}


def recovered(evidence, fault):
    if fault in ("disconnect", "late"):
        mode = "drop-query-once" if fault == "disconnect" else "late-query-once"
        docker("fault-proxy", "exec", "python", "-c", f"from pathlib import Path; Path('/tmp/bus-fault-mode').write_text('{mode}')")
    run_id, plan_id = new_run(fault)
    evidence.update(run_id=run_id, plan_id=plan_id)
    injected = False
    before = None

    def hook(graph):
        nonlocal injected, before
        if fault != "host-kill" or injected:
            return
        nodes = {n["id"]: n for n in graph["nodes"]}
        if nodes["intake"]["status"] == "done" and any(nodes[k].get("remote", {}).get("task_id") for k in ("financial", "industry")):
            before = call_snapshot(read_run(run_id))
            evidence["before_kill"] = before
            docker("orchestrator", "kill", "--signal", "KILL")
            docker("orchestrator", "start")
            injected = True

    graph = wait_run(run_id, evidence, hook)
    if fault in ("disconnect", "late"):
        assert graph["run"]["status"] == "waiting", "Lost response must remain unresolved"
        before = call_snapshot(read_run(run_id))
        assert before["intake"]["remote_task_id"], "Original task receipt was lost"
        evidence["before_resume"] = before
        resume(run_id)
        graph = wait_run(run_id, evidence)
    else:
        assert injected, "Host fault was not actually injected"
    run = read_run(run_id)
    evidence["run"] = redact(run)
    assert run["status"] == "done", run.get("final", run["status"])
    after = call_snapshot(run)
    for key, original in before.items():
        assert after[key]["request_id"] == original["request_id"], "Recovery created a second send"
        if original["remote_task_id"]:
            assert after[key]["remote_task_id"] == original["remote_task_id"]
        if original["artifact_id"]:
            for field in ("artifact_id", "artifact_digest", "model_events"):
                assert after[key][field] == original[field], f"Completed node changed: {key}.{field}"
    if fault in ("disconnect", "late"):
        events = [json.loads(line) for line in docker("fault-proxy", "exec", "cat", "/tmp/bus-fault-events.jsonl").splitlines()]
        ids = {c["request_id"] for c in after.values()}
        sends = [e for e in events if e["method"] == "SendMessage" and e["id"] in ids]
        assert len(sends) == len(after) == 5, "Transport recorded duplicate dispatch"
        evidence["transport_events"] = [e for e in events if e["id"] in ids or e.get("task_id") in {c["remote_task_id"] for c in after.values()}]
    evidence["checks"] = {"same_request_and_task": True, "no_completed_node_model_increment": fault == "host-kill", "completed": True}


def component_restart(evidence, service):
    run_id, plan_id = new_run(service + "-restart")
    evidence.update(run_id=run_id, plan_id=plan_id)
    injected = False
    before = None

    def hook(graph):
        nonlocal injected, before
        intake = next(n for n in graph["nodes"] if n["id"] == "intake")
        if not injected and intake.get("remote", {}).get("task_id") and intake["status"] != "done":
            before = call_snapshot(read_run(run_id))
            evidence["before_restart"] = before
            docker(service, "restart")
            injected = True
    graph = wait_run(run_id, evidence, hook)
    assert injected, "Component restart was not injected"
    assert graph["run"]["status"] == "waiting", "Lost upstream task ownership must remain unresolved"
    # Once healthy, querying again must still use the original identity; an
    # in-memory upstream task store is not magically restored by restarting.
    for _ in range(90):
        status = docker(service, "inspect", "--format", "{{.State.Health.Status}}")
        if status.strip() == "healthy":
            break
        time.sleep(1)
    resume(run_id)
    graph = wait_run(run_id, evidence)
    run = read_run(run_id)
    evidence["run"] = redact(run)
    assert run["status"] == "waiting"
    after = call_snapshot(run)
    assert after["intake"]["request_id"] == before["intake"]["request_id"]
    assert after["intake"]["remote_task_id"] == before["intake"]["remote_task_id"]
    assert not after["intake"]["artifact_id"]
    assert all(s["status"] == "pending" for k, s in run["steps"].items() if k != "intake")
    evidence["checks"] = {"manual_reconciliation_required": True, "no_redispatch": True, "dependents_blocked": True}


def storage_failure(evidence):
    docker("fault-proxy", "exec", "python", "-c", "from pathlib import Path; Path('/tmp/bus-release').unlink(missing_ok=True); Path('/tmp/bus-completed-ready').unlink(missing_ok=True); Path('/tmp/bus-fault-mode').write_text('hold-completed-once')")
    run_id, plan_id = new_run("storage-failure")
    evidence.update(run_id=run_id, plan_id=plan_id)
    injected = False
    restored = False
    before = None
    try:
        deadline = time.monotonic() + 240
        while time.monotonic() < deadline:
            ready = docker("fault-proxy", "exec", "python", "-c", "from pathlib import Path; print(Path('/tmp/bus-completed-ready').exists())").strip()
            if ready == "True":
                break
            time.sleep(0.5)
        assert ready == "True", "No actual completed remote task intercepted"
        before = call_snapshot(read_run(run_id))
        evidence["before_storage_fault"] = before
        # The mounted state is saved intact. A directory at its destination makes
        # the actual atomic rename fail; no storage implementation is mocked.
        docker("credit", "exec", "python", "-c", "from pathlib import Path; p=Path('/data/state.json'); backup=p.with_suffix('.bus-test-backup'); assert not backup.exists(); p.rename(backup); p.mkdir()")
        injected = True
        docker("fault-proxy", "exec", "python", "-c", "from pathlib import Path; Path('/tmp/bus-release').write_text('release')")
        for _ in range(30):
            try:
                request("/healthz")
            except urllib.error.HTTPError as exc:
                assert exc.code == 503
                break
            except ValueError:  # Healthy plain-text response.
                pass
            time.sleep(0.5)
        else:
            raise AssertionError("Storage write failure did not freeze service")
        run = read_run(run_id)
        evidence["during_storage_fault"] = redact(run)
        assert run["steps"]["intake"]["status"] != "done"
        assert not run["steps"]["intake"]["remote_call"].get("artifact_id")
        assert all(s["status"] == "pending" for k, s in run["steps"].items() if k != "intake")
    finally:
        if injected:
            docker("credit", "exec", "python", "-c", "from pathlib import Path; p=Path('/data/state.json'); p.rmdir(); p.with_suffix('.bus-test-backup').rename(p)")
            restored = True
            docker("credit", "restart")
            docker("orchestrator", "restart")
        docker("fault-proxy", "exec", "python", "-c", "from pathlib import Path; Path('/tmp/bus-release').write_text('release'); Path('/tmp/bus-fault-mode').write_text('pass')")
    assert restored
    graph = wait_run(run_id, evidence)
    run = read_run(run_id)
    evidence["run"] = redact(run)
    assert run["status"] == "done", run.get("final")
    after = call_snapshot(run)
    for field in ("request_id", "remote_task_id", "model_events"):
        assert after["intake"][field] == before["intake"][field], "Recovery reran the remote calculation"
    evidence["checks"] = {"real_atomic_rename_failure": True, "no_success_before_durable_commit": True, "dependents_blocked": True, "original_task_reconciled": True}


def wait_run(run_id, evidence, hook=None, timeout=420):
    deadline = time.monotonic() + timeout
    previous = None
    while time.monotonic() < deadline:
        graph, _ = request(f"/api/graph?run={run_id}")
        signature = [(n["id"], n.get("status")) for n in graph["nodes"]]
        if signature != previous:
            print(run_id, graph["run"]["status"], signature, flush=True)
            previous = signature
        evidence["snapshots"].append({"at": time.time(), "graph": graph})
        if hook:
            hook(graph)
        if graph["run"]["status"] in ("done", "failed", "waiting"):
            return graph
        time.sleep(0.5)
    raise TimeoutError("Graph did not reach a terminal or explicit waiting state")


def normal(evidence):
    run_id, plan_id = new_run("normal")
    evidence.update(run_id=run_id, plan_id=plan_id)
    graph = wait_run(run_id, evidence)
    run = read_run(run_id)
    evidence["run"] = redact(run)
    assert run["status"] == "done", run.get("final", run["status"])
    calls = {k: v["remote_call"] for k, v in run["steps"].items() if "remote_call" in v}
    assert len(calls) == 5, "All five ordinary professional nodes must use the three remote services"
    assert len({c["binding"]["agent_id"] for c in calls.values()}) == 3
    assert all(c["remote_task_id"] and c["artifact_id"] and c["model_events"] for c in calls.values())
    assert len({c["context_id"] for c in calls.values()}) == len(calls)
    from datetime import datetime
    start = lambda k: datetime.fromisoformat(calls[k]["agent_started_at"])
    end = lambda k: datetime.fromisoformat(run["steps"][k]["finished_at"])
    assert max(start("financial"), start("industry")) < min(end("financial"), end("industry")), "Analysis nodes did not actually overlap"
    assert "capability_token" not in json.dumps(graph), "Scoped credential leaked into projection"
    evidence["checks"] = {"real_roles": 3, "remote_nodes": len(calls), "parallel_analysis": True,
                          "isolated_contexts": True, "correlated_artifacts": True,
                          "model_attempts": {k: len(v["model_events"]) for k, v in calls.items()}}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("scenario", choices=["normal", "host-kill", "disconnect", "late", "gateway-restart", "runtime-restart", "storage-failure"], default="normal", nargs="?")
    args = parser.parse_args()
    evidence = {"scenario": args.scenario, "started_at": time.time(), "scope": "Agent Bus integration on synthetic credit data; W06 annual-report checks excluded", "snapshots": [], "environment": environment_record()}
    directory = CACHE / "evidence" / "agent-bus"
    directory.mkdir(parents=True, exist_ok=True)
    target = directory / f"{args.scenario}-{int(time.time())}.json"
    try:
        if args.scenario == "normal":
            normal(evidence)
        elif args.scenario in ("host-kill", "disconnect", "late"):
            recovered(evidence, args.scenario)
        elif args.scenario == "storage-failure":
            storage_failure(evidence)
        else:
            component_restart(evidence, "gateway" if args.scenario == "gateway-restart" else "material-agent")
        evidence["passed"] = True
    except Exception as exc:
        evidence["passed"] = False
        evidence["error"] = str(exc)
        raise
    finally:
        evidence["finished_at"] = time.time()
        target.write_text(json.dumps(redact(evidence), ensure_ascii=False, indent=2) + "\n")
        print("Evidence:", target, flush=True)


if __name__ == "__main__":
    main()
