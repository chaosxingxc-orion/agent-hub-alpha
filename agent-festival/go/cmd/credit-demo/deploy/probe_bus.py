#!/usr/bin/env python3
"""Protocol/admission checks against the running real Bus; no model dispatch."""
import copy
import argparse
import json
import subprocess
import time
from pathlib import Path
from smoke_bus import CACHE, PROJECT, Planner, TEAM, redact, request, environment_record
import demo


def inside_gateway(probe):
    code = '''
import json,os,sys,urllib.request,urllib.error
p=json.load(sys.stdin)
headers={"Content-Type":"application/json","X-Tenant-ID":"issue16"}
if p.get("auth",True): headers["Authorization"]="Bearer "+os.environ["GATEWAY_TEST_CREDENTIAL_TOKEN" if p["kind"]=="rpc" else "GATEWAY_RDC_TOKEN"]
url="http://127.0.0.1:8080/a2a" if p["kind"]=="rpc" else "http://rdc:8092"+p["path"]
q=urllib.request.Request(url,data=json.dumps(p["body"]).encode() if p["kind"]=="rpc" else None,headers=headers)
try:
    with urllib.request.urlopen(q,timeout=15) as r: out={"http":r.status,"body":json.load(r)}
except urllib.error.HTTPError as e: out={"http":e.code,"body":json.loads(e.read())}
print(json.dumps(out))
'''
    result = subprocess.run(["docker", "exec", "-i", f"{PROJECT}-gateway-1", "python", "-c", code],
                            input=json.dumps(probe), text=True, capture_output=True, check=True)
    return json.loads(result.stdout)


def rpc(method, params, auth=True):
    return inside_gateway({"kind": "rpc", "auth": auth, "body": {"jsonrpc": "2.0", "id": "acceptance-probe", "method": method, "params": params}})


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--rdc-outage", action="store_true", help="Temporarily stop this project's RDC; always restore it")
    args = parser.parse_args()
    evidence = {"scenario": "protocol-admission", "started_at": time.time(), "cases": {}, "environment": environment_record()}
    target = CACHE / "evidence/agent-bus" / f"protocol-admission-{int(time.time())}.json"
    try:
        for agent in ("credit-material", "credit-analysis", "credit-writer"):
            card = inside_gateway({"kind": "rdc", "path": f"/v2/agents/{agent}/card"})
            instances = inside_gateway({"kind": "rdc", "path": f"/v2/agents/{agent}/instances"})
            assert card["http"] == instances["http"] == 200
            assert any(x["health"] == "ONLINE" for x in instances["body"]["instances"])
            evidence["cases"][agent] = redact({"card": card, "instances": instances})
        unauthorized = rpc("GetTask", {"id": "not-a-task"}, auth=False)
        assert unauthorized["http"] == 401
        unknown = rpc("GetTask", {"id": "not-a-task"})
        assert unknown["http"] == 200 and unknown["body"]["error"]["data"]["code"] == "ROUTE_OWNERSHIP_UNKNOWN"
        invalid = rpc("SendMessage", {})
        assert "error" in invalid["body"] or invalid["http"] >= 400
        missing = rpc("SendMessage", {"message": {"role": "ROLE_USER", "messageId": "rejected-probe", "parts": [{"kind": "text", "text": "no execution"}]},
                                      "configuration": {"returnImmediately": True}, "metadata": {"agentId": "unregistered-agent"}})
        assert "error" in missing["body"] or missing["http"] >= 400
        evidence["cases"].update(unauthorized=unauthorized, unknown_task=unknown, invalid_request=invalid, unknown_capability=missing)
        rejected = rpc("SendMessage", {"message": {"role": "ROLE_USER", "messageId": "invalid-node-scope", "contextId": "unapproved-context",
                         "parts": [{"kind": "text", "text": json.dumps({"run_id": "nonexistent", "step": "report", "token": "invalid-test-token"})}]},
                         "configuration": {"returnImmediately": True}, "metadata": {"agentId": "credit-writer"}})
        task = rejected["body"]["result"].get("task", rejected["body"]["result"])
        for _ in range(30):
            if task["status"]["state"] == "TASK_STATE_FAILED":
                break
            time.sleep(0.2)
            checked = rpc("GetTask", {"id": task["id"], "historyLength": 0})
            task = checked["body"]["result"].get("task", checked["body"]["result"])
        assert task["status"]["state"] == "TASK_STATE_FAILED", "Invalid node scope must fail before model execution"
        evidence["cases"]["runtime_rejected_scope"] = redact(task)
        state, _ = request("/debug/state")
        original_runs = sorted(state["runs"])
        planner = Planner()
        for expert in ("unregistered-agent", "http://attacker.invalid/a2a"):
            agents = copy.deepcopy(TEAM)
            agents[0]["expert"] = expert
            try:
                planner.tool("propose_plan", {"case_id": "C006", "goal": "admission negative test", "agents": agents})
            except RuntimeError as exc:
                evidence["cases"][expert] = {"rejected": True, "reason": str(exc)}
            else:
                raise AssertionError("Unregistered capability/address accepted")
        agents = copy.deepcopy(TEAM)
        agents[3]["expert"] = "bureau-full"
        denied = planner.tool("propose_plan", {"case_id": "C006", "goal": "permission negative test", "agents": agents})
        assert not denied["precheck"]["ok"]
        evidence["cases"]["missing_tool_permission"] = denied["precheck"]
        after, _ = request("/debug/state")
        assert sorted(after["runs"]) == original_runs, "Negative admission created an execution"
        # Use a completed real task for the actual protocol/response sample.
        candidates = [s["remote_call"] for r in after["runs"].values() if r["status"] == "done" for s in r["steps"].values() if "remote_call" in s]
        for call in reversed(candidates):
            result = rpc("GetTask", {"id": call["remote_task_id"], "historyLength": 0})
            if "result" in result["body"]:
                evidence["cases"]["completed_task"] = redact(result)
                break
        if args.rdc_outage:
            demo.BUS = True
            try:
                demo.compose("-p", PROJECT, "stop", "rdc")
                catalog = planner.tool("list_capabilities", {})
                assert all(not c.get("available") for c in catalog["experts"] if c.get("remote"))
                result = rpc("SendMessage", {"message": {"role": "ROLE_USER", "messageId": "rdc-outage", "parts": [{"kind": "text", "text": "no dispatch"}]},
                                             "metadata": {"agentId": "unregistered-during-outage"}})
                assert "error" in result["body"] or result["http"] >= 400
                evidence["cases"]["rdc_unavailable"] = {"catalog": catalog, "gateway": result}
            finally:
                demo.compose("-p", PROJECT, "up", "-d", "--no-build", "--wait", "rdc")
        evidence["passed"] = True
    except Exception as exc:
        evidence.update(passed=False, error=str(exc))
        raise
    finally:
        evidence["finished_at"] = time.time()
        target.write_text(json.dumps(redact(evidence), ensure_ascii=False, indent=2) + "\n")
        print("Evidence:", target)


if __name__ == "__main__":
    main()
