"""Prepare by default; explicitly verify an already owned original H03 chain."""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import http.cookiejar
import json
from pathlib import Path
import re
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request

import launch_h03 as launch

HASH = re.compile(r"sha256:[0-9a-f]{64}\Z")
QUERY = {"package": "h03-engineering-two-role", "entity": "engineering-entity", "basis": "consolidated",
         "period": "FY2025", "focus": ["reconciliation"], "deliverable": "engineering-reconciliation-v1",
         "materials": ["engineering-material-a"]}


class Blocked(ValueError):
    """Finite operator-facing failure without raw private responses."""


def now():
    return datetime.now(timezone.utc).isoformat()


def sha(*parts):
    d = hashlib.sha256()
    for part in parts:
        d.update(part.encode("utf-8")); d.update(b"\0")
    return "sha256:" + d.hexdigest()


def raw_hash(value):
    return "sha256:" + hashlib.sha256(launch.canonical(value)).hexdigest()


def strict(raw):
    if not raw or len(raw) > 2 << 20:
        raise Blocked("bounded_native_document_required")
    try:
        return json.loads(raw, object_pairs_hook=launch.unique,
                          parse_constant=lambda _: (_ for _ in ()).throw(Blocked("finite_native_json_required")))
    except (ValueError, UnicodeError):
        raise Blocked("strict_native_json_required") from None


def evidence_path(path, deployment=None):
    root = launch.owned_path(path)
    if deployment is not None:
        try:
            root.resolve().relative_to(deployment.resolve())
        except ValueError:
            pass
        else:
            raise Blocked("verification_output_must_not_modify_deployment")
    return root


def receipt_file(parent, relative, pin):
    path = Path(parent) / relative
    if not isinstance(relative, str) or Path(relative).is_absolute() or ".." in Path(relative).parts:
        raise Blocked("bounded_receipt_path_required")
    launch.owned_path(path)
    if any(p.is_symlink() for p in [path, *path.parents]) or not path.is_file() or path.stat().st_size > 2 << 20 or launch.digest(path) != pin:
        raise Blocked("fixed_receipt_unavailable")
    return path


def contains(value, wanted):
    if isinstance(value, dict): return any(contains(v, wanted) for v in value.values())
    if isinstance(value, list): return any(contains(v, wanted) for v in value)
    return value == wanted


def source_receipts(path, images):
    path = launch.owned_path(path)
    manifest = launch.read_json(path)
    if set(manifest) != {"schema_version", "images"} or manifest["schema_version"] != "h03-native-image-receipts/v1" or set(manifest["images"]) != set(images):
        raise Blocked("exact_original_image_receipts_required")
    safe = {}
    for name, image in images.items():
        entry = manifest["images"][name]
        if set(entry) != {"image_id", "receipt", "receipt_sha256", "source_manifest", "source_manifest_sha256"} or entry["image_id"] != image:
            raise Blocked("image_receipt_identity_differs")
        receipt = receipt_file(path.parent, entry["receipt"], entry["receipt_sha256"])
        source = receipt_file(path.parent, entry["source_manifest"], entry["source_manifest_sha256"])
        if not contains(strict(receipt.read_bytes()), image):
            raise Blocked("original_receipt_does_not_name_image")
        strict(source.read_bytes())
        safe[name] = {"image_id": image, "receipt_sha256": launch.digest(receipt), "source_manifest_sha256": launch.digest(source)}
    return safe


def deployed(project, root, receipts):
    root = launch.owned_path(root)
    state = launch.read_json(root / "ownership.json")
    private = launch.read_json(root / "private-local.json")
    if state.get("schema_version") != "h03-owned-deployment/v1" or state.get("project") != project or state.get("owner") != private.get("H03_OWNER") or launch.digest(root / "compose.h03.yaml") != state["config_sha256"]:
        raise Blocked("owned_deployment_identity_required")
    if any(launch.digest(root / p) != pin for p, pin in state["asset_sha256"].items()):
        raise Blocked("owned_configuration_changed")
    env = launch.environment(); env.update(private); env["COMPOSE_DISABLE_ENV_FILE"] = "true"
    context = strict(launch.docker(["context", "inspect", "default"], env, root))
    if len(context) != 1 or not context[0]["Endpoints"]["docker"]["Host"].startswith("unix://"):
        raise Blocked("local_daemon_required")
    raw = launch.read_json(root / "compose.h03.yaml")
    expected = launch.expand(raw, env)
    if set(expected["services"]) != set(launch.selected(state["ui"])):
        raise Blocked("deployment_service_allowlist_differs")
    sources = source_receipts(receipts, state["images"])
    ids = launch.docker(["ps", "--all", "--quiet", "--filter", "label=com.docker.compose.project=" + project], env, root).split()
    if len(ids) != len(expected["services"]) or any(not re.fullmatch(r"[0-9a-f]{12,64}", v) for v in ids):
        raise Blocked("exact_owned_container_set_required")
    items = strict(launch.docker(["inspect", *ids], env, root))
    launch.container_audit(items, state, expected)
    containers = {}
    for item in items:
        name = item["Config"]["Labels"]["com.docker.compose.service"]
        if not item.get("State", {}).get("Running") or item.get("State", {}).get("Paused"):
            raise Blocked("all_original_services_must_be_running")
        fresh = strict(launch.docker(["image", "inspect", state["images"][name]], env, root))[0]
        if launch.image_audit(name, fresh, state["images"][name]) != state["image_envs"][name]:
            raise Blocked("original_image_configuration_changed")
        containers[name] = item["Id"]
    network = strict(launch.docker(["network", "inspect", project + "_h03"], env, root))[0]
    if network.get("Internal") is not True or network.get("Labels", {}).get("io.agent-hub.h03.owner") != state["owner"]:
        raise Blocked("owned_internal_network_required")
    return {"root": root, "state": state, "env": env, "expected": expected, "private": private,
            "containers": containers, "sources": sources}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl): return None


class Client:
    def __init__(self):
        self.base = "http://127.0.0.1:18191"
        self.cookies = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), urllib.request.HTTPCookieProcessor(self.cookies), NoRedirect())
        self.session, self.sequence = None, 0

    def http(self, method, path, value=None, *, form=False, headers=None):
        if not path.startswith("/") or path.startswith("//") or "\r" in path or "\n" in path:
            raise Blocked("finite_local_path_required")
        raw = None if value is None else (urllib.parse.urlencode(value).encode() if form else launch.canonical(value))
        req = urllib.request.Request(self.base + path, data=raw, method=method, headers={
            "Content-Type": "application/x-www-form-urlencoded" if form else "application/json", **(headers or {})})
        try:
            response = self.opener.open(req, timeout=20)
        except urllib.error.HTTPError as error:
            response = error
        except (urllib.error.URLError, TimeoutError):
            raise Blocked("owned_http_unavailable") from None
        with response:
            body = response.read((2 << 20) + 1)
            if len(body) > 2 << 20: raise Blocked("bounded_http_response_required")
            return response.code, dict(response.headers), body

    @staticmethod
    def rpc_body(body):
        if body.lstrip().startswith(b"{"): return strict(body)
        events = []
        for event in body.decode("utf-8").replace("\r\n", "\n").split("\n\n"):
            data = "\n".join(line[5:].lstrip() for line in event.splitlines() if line.startswith("data:"))
            if data: events.append(strict(data))
        if len(events) != 1: raise Blocked("one_original_mcp_result_required")
        return events[0]

    def rpc(self, method, params=None, *, notification=False):
        self.sequence += 1
        value = {"jsonrpc": "2.0", "method": method, "params": params or {}}
        if not notification: value["id"] = self.sequence
        headers = {"Accept": "application/json, text/event-stream", "MCP-Protocol-Version": "2025-03-26"}
        if self.session: headers["Mcp-Session-Id"] = self.session
        code, received, body = self.http("POST", "/mcp", value, headers=headers)
        if code not in (200, 202): raise Blocked("original_mcp_rejected")
        for key, v in received.items():
            if key.lower() == "mcp-session-id": self.session = v
        if notification: return None
        result = self.rpc_body(body)
        if result.get("jsonrpc") != "2.0" or result.get("id") != self.sequence or "error" in result:
            raise Blocked("original_mcp_identity_or_error")
        return result["result"]

    def initialize(self):
        self.rpc("initialize", {"protocolVersion": "2025-03-26", "capabilities": {}, "clientInfo": {"name": "h03-owned-verifier", "version": "0.1.0"}})
        self.rpc("notifications/initialized", notification=True)

    def tool(self, name, arguments):
        if name not in {"select_capabilities", "propose_plan", "run_status", "get_plan", "verify_run"}:
            raise Blocked("original_planner_tool_allowlist_required")
        value = self.rpc("tools/call", {"name": name, "arguments": arguments})
        content = value.get("content", [])
        if value.get("isError") or len(content) != 1 or content[0].get("type") != "text": raise Blocked("original_tool_rejected")
        return strict(content[0]["text"])

    def login(self, password):
        code, headers, _ = self.http("POST", "/ui/login", {"user": "alice", "password": password, "next": "/ui/"}, form=True)
        location = headers.get("Location", headers.get("location", ""))
        if code != 303 or location != "/ui/" or not any(c.name == "cd_session" for c in self.cookies):
            raise Blocked("original_login_session_required")

    def approve(self, plan):
        if not re.fullmatch(r"P[0-9]{1,6}", str(plan.get("plan_id", ""))) or not isinstance(plan.get("version"), int) or not HASH.fullmatch(plan.get("hash", "")):
            raise Blocked("original_frozen_plan_identity_required")
        path = "/ui/plan/" + plan["plan_id"] + "/v" + str(plan["version"]) + "/approve"
        code, headers, _ = self.http("POST", path, {"hash": plan["hash"]}, form=True)
        location = headers.get("Location", headers.get("location", ""))
        if code != 303 or not re.fullmatch(r"/ui/run/R[0-9]{1,6}", location):
            raise Blocked("original_logged_approve_required")
        return location.rsplit("/", 1)[1]


    def graph(self, run_id):
        code, _, body = self.http("GET", "/api/graph?run=" + run_id + "&verify=1")
        if code != 200: raise Blocked("original_graph_unavailable")
        return strict(body)

    def reauthenticate(self, password):
        self.session = None
        self.cookies.clear()
        self.initialize()
        self.login(password)


def state_read(deployment):
    # Raw private state stays in memory; no debug endpoint or public copy.
    return strict(launch.docker(["exec", deployment["containers"]["credit"], "/bin/cat", "/data/state.json"], deployment["env"], deployment["root"]))


def original_run(state, run):
    value = state.get("runs", {}).get(run)
    if not isinstance(value, dict): raise Blocked("original_private_run_missing")
    plans = state.get("plans", {}).get(value.get("plan_id"), [])
    version = value.get("plan_version", 0)
    if not isinstance(version, int) or version < 1 or version > len(plans): raise Blocked("adopted_plan_missing")
    return value, plans[version - 1]


def formal_receipt(state, run_id, expected_a):
    run, plan = original_run(state, run_id)
    if run.get("status") != "done" or run.get("approved_by") != "alice" or run.get("approved_via", {}).get("channel") != "confirm-page" or run.get("plan_hash") != plan.get("hash"):
        raise Blocked("original_logged_approval_or_completion_missing")
    binding = plan.get("execution_binding")
    if not isinstance(binding, dict) or binding.get("profile") != "two-role-identity/v1": raise Blocked("original_adopted_profile_required")
    contract_map = binding["node_contracts"]
    steps = run.get("steps", {})
    if len(steps) != 2 or set(run.get("order", [])) != set(steps): raise Blocked("two_original_steps_required")
    receipts = {}; calls = {}; services = set(); model_events = 0
    for node, step in steps.items():
        call = step.get("remote_call", {}); role = step.get("expert")
        if role not in ("eng.sum", "eng.reconcile") or role in receipts or step.get("status") != "done" or call.get("status") != "completed": raise Blocked("two_completed_roles_required")
        service = "h03-engineering-" + role.split(".")[1]
        if call.get("binding", {}).get("agent_id") != service or call.get("binding", {}).get("version") != "0.1.0": raise Blocked("original_role_identity_differs")
        services.add(service)
        if not all(isinstance(call.get(k), str) and call[k] for k in ("context_id", "remote_task_id", "agent_started_at", "artifact_id", "request_id")): raise Blocked("original_role_task_receipt_missing")
        raw_input = call["input"]; payload = strict(raw_input); output = step.get("output", "")
        prefix = "[step:" + node + "]\n"
        if not output.startswith(prefix): raise Blocked("original_formal_step_identity_differs")
        raw_output = output[len(prefix):]; artifact = strict(raw_output)
        if launch.canonical(payload).decode() != raw_input or launch.canonical(artifact).decode() != raw_output: raise Blocked("canonical_formal_bytes_required")
        if payload.get("run_id") != run_id or payload.get("node_path") != node or payload.get("attempt") != 1 or call.get("attempt") != 1 or payload.get("node_execution_id") != call.get("node_execution_id") or sha(run_id, call["graph_revision"], node, "iteration:0") != call["node_execution_id"]:
            raise Blocked("original_node_identity_differs")
        if "context_id" in payload or payload.get("plan_hash") != plan["hash"] or payload.get("execution_binding_sha256") != raw_hash(binding) or payload.get("node_contract_sha256") != raw_hash(contract_map[node]): raise Blocked("adopted_input_closure_differs")
        preimage = {"input": raw_input, "goal": call["goal"], "binding": call["binding"], "model": run.get("exec_binding", {}).get("model", "")}
        if sha(launch.canonical(preimage).decode()) != call.get("input_digest") or sha(raw_output) != call.get("artifact_digest"):
            raise Blocked("original_nul_digest_differs")
        if artifact.get("schema_version") != "hub-node-output/v1" or artifact.get("input_digest") != call["input_digest"] or artifact.get("execution_binding_sha256") != payload["execution_binding_sha256"] or artifact.get("node_contract_sha256") != payload["node_contract_sha256"]:
            raise Blocked("formal_output_binding_differs")
        model_events += len(call.get("model_events", {}))
        events = list(call.get("gateway_exchanges", {}).values())
        sends = [e for e in events if e.get("method") == "SendMessage"]
        queries = [e for e in events if e.get("method") == "GetTask"]
        if len(sends) != 1 or not any(e.get("outcome") == "task_received" and e.get("http_status") == 200 for e in events): raise Blocked("original_gateway_receipt_missing")
        receipts[role] = {"node": node, "input": raw_input, "output": raw_output, "input_digest": call["input_digest"], "artifact_digest": call["artifact_digest"],
                          "context_id": call["context_id"], "remote_task_id": call["remote_task_id"], "node_execution_id": call["node_execution_id"], "agent_started_at": call["agent_started_at"],
                          "service": service, "send_count": len(sends), "query_count": len(queries)}
        calls[role] = call
    if len(services) != 2 or calls["eng.sum"]["context_id"] == calls["eng.reconcile"]["context_id"] or calls["eng.sum"]["remote_task_id"] == calls["eng.reconcile"]["remote_task_id"]:
        raise Blocked("two_independent_role_tasks_required")
    def values(raw):
        fields = strict(raw).get("fields", [])
        if not isinstance(fields, list) or len({f.get("name") for f in fields}) != len(fields): raise Blocked("unique_consumed_fields_required")
        if any((f.get("type"), f.get("unit"), f.get("period"), f.get("basis")) != ("decimal", "USD-million", "FY2025", "consolidated") for f in fields): raise Blocked("consumed_field_semantics_differ")
        return {f["name"]: f.get("value") for f in fields}
    sum_input = values(receipts["eng.sum"]["input"]); sum_output = values(receipts["eng.sum"]["output"])
    expected_total, expected_difference = str(expected_a + 60), str(expected_a + 60 - 95)
    if sum_input != {"component_a": str(expected_a), "component_b": "60", "reference": "95"} or sum_output != {"reference": "95", "total": expected_total} or values(receipts["eng.reconcile"]["input"]) != sum_output or values(receipts["eng.reconcile"]["output"]) != {"difference": expected_difference}:
        raise Blocked("actual_values_do_not_match_fixed_consumption_case")
    materials = binding.get("materials", [])
    if len(materials) != 1 or strict(receipts["eng.sum"]["input"])["dependencies"] != [{"kind": "material", "id": materials[0]["id"], "sha256": "sha256:" + materials[0]["sha256"]}]:
        raise Blocked("adopted_material_dependency_differs")
    deps = strict(receipts["eng.reconcile"]["input"])["dependencies"]
    if deps != [{"kind": "step", "id": receipts["eng.sum"]["node"], "node_execution_id": receipts["eng.sum"]["node_execution_id"], "artifact_digest": receipts["eng.sum"]["artifact_digest"]}]: raise Blocked("formal_upstream_dependency_differs")
    if model_events: raise Blocked("framework_model_events_are_not_zero")
    return {"run_id": run_id, "plan_id": run["plan_id"], "plan_hash": plan["hash"], "approval_channel": "confirm-page", "roles": receipts,
            "observed_framework_model_events": model_events, "actual_role_model_bank_egress_counters": "BLOCKED_REQUIRES_EXTERNAL_NATIVE_OBSERVER"}


def graph_receipt(graph, receipt):
    if graph.get("run", {}).get("id") != receipt["run_id"] or graph.get("run", {}).get("status") != "done" or graph.get("hash") != receipt["plan_hash"] or not graph.get("verify", {}).get("ok"):
        raise Blocked("original_graph_and_formal_run_differ")
    nodes = {n["id"]: n for n in graph.get("nodes", [])}
    for role in receipt["roles"].values():
        node = nodes.get(role["node"], {}); remote = node.get("remote", {}); consumption = node.get("consumption", {})
        if node.get("status") not in ("done", "reused") or remote.get("task_id") != role["remote_task_id"] or remote.get("context_id") != role["context_id"] or consumption.get("status") != "committed" or consumption.get("input_digest") != role["input_digest"] or consumption.get("artifact_digest") != role["artifact_digest"]:
            raise Blocked("original_page_projection_differs")
    return {"status": "PASS_ORIGINAL_GO_GRAPH_PROJECTION", "run_id": receipt["run_id"], "revision": graph.get("rev"), "workswarm_actual_page": "BLOCKED_REQUIRES_ORIGINAL_HOST_AND_BROWSER_OBSERVATION"}


def native_negatives(deployment, client, state, run_id):
    """Denied/control probes are separate from normal business counters."""
    run, _ = original_run(state, run_id)
    record = {"interval": "TOOL_VALIDATION_SEPARATE_FROM_BUSINESS", "started_at": now(), "checks": {}}
    for path, expected in [("/exec/lease", 401), ("/agent/model", 404), ("/member-mcp/eng.sum", 404), ("/debug/state", 404)]:
        code, headers, response = client.http("POST" if path != "/debug/state" else "GET", path, {"worker": "denied-probe"} if path == "/exec/lease" else None)
        if code != expected: raise Blocked("private_surface_was_not_denied")
        if path == "/debug/state":
            if response != b"404 page not found\n":
                raise Blocked("profile_outer_debug_denial_was_not_static")
            record["debug_denial_route"] = "profile_outer_handler_404; inner build/history 403 is a separate gate"
        record["checks"][path] = code
    node, step = next(iter(run["steps"].items())); call = step["remote_call"]
    if run.get("token"): raise Blocked("finished_original_lease_not_closed")
    body = {"run_id": run_id, "token": "expired-native-probe-token", "step": node, "node_execution_id": call["node_execution_id"], "input_digest": call["input_digest"],
            "attempt": call["attempt"], "remote_task_id": call["remote_task_id"], "context_id": call["context_id"], "output": step["output"].removeprefix("[step:" + node + "]\n")}
    header = {"X-Hub-Runtime-Credential": deployment["private"]["H03_RUNTIME_CREDENTIAL"]}
    code, _, _ = client.http("POST", "/exec/call/commit", body, headers=header)
    if code != 409: raise Blocked("closed_original_lease_was_not_fenced")
    record["checks"]["identical_commit_after_done_fenced"] = code
    for key, value in [("remote_task_id", "foreign-task"), ("context_id", "foreign-context"), ("input_digest", "sha256:" + "0" * 64), ("output", "{}")]:
        changed = dict(body); changed[key] = value
        code, _, _ = client.http("POST", "/exec/call/commit", changed, headers=header)
        if code not in (400, 409): raise Blocked("conflicting_formal_commit_not_rejected")
        record["checks"]["closed_lease_wrong_" + key] = code
    after, _ = original_run(state_read(deployment), run_id)
    if after["steps"] != run["steps"] or after["status"] != run["status"]:
        raise Blocked("rejected_probe_changed_original_formal_state")
    record["original_formal_state_unchanged"] = True
    record["finished_at"] = now()
    record["live_lease_idempotency_and_field_validation"] = "BLOCKED_REQUIRES_ACTIVE_OWNED_LEASE_WINDOW"
    return record


def wait_state(deployment, run, target, timeout):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        state = state_read(deployment); current, _ = original_run(state, run)
        if current.get("status") in target: return state
        if current.get("status") == "failed": raise Blocked("original_run_failed")
        time.sleep(.5)
    raise Blocked("original_run_deadline_without_redispatch")


def propose(client):
    selection = client.tool("select_capabilities", QUERY)
    if selection.get("outcome") != "graph" or set(selection.get("selection", [])) != {"eng.sum", "eng.reconcile"}:
        raise Blocked("original_selector_did_not_select_complete_pair")
    plan = client.tool("propose_plan", {"goal": "controlled two-role engineering consumption; no business acceptance", "selection_id": selection["id"]})
    if not plan.get("precheck", {}).get("ok"): raise Blocked("original_proposal_precheck_failed")
    return plan, {"selection_id": selection["id"], "selection_digest": selection["digest"], "selected": selection["selection"], "plan_id": plan["plan_id"], "plan_hash": plan["hash"]}


def planning_negatives(client):
    results = []
    for key, value in [("period", "FY2024"), ("basis", "standalone")]:
        query = dict(QUERY); query[key] = value
        record = client.tool("select_capabilities", query)
        if record.get("outcome") != "no_graph" or not record.get("eligibility") or any(e.get("result") == "pass" for e in record["eligibility"]):
            raise Blocked("inapplicable_original_selection_was_not_excluded")
        results.append({"query_field": key, "selection_id": record["id"], "outcome": record["outcome"],
                        "candidate_results": [{"capability": e["capability"], "result": e["result"], "reasons": e.get("reasons", [])} for e in record["eligibility"]]})
    return results


def restart_owned_control(deployment, client, password, timeout):
    credit = deployment["containers"]["credit"]
    def inspect():
        item = strict(launch.docker(["inspect", credit], deployment["env"], deployment["root"]))[0]
        labels = item.get("Config", {}).get("Labels", {})
        state = deployment["state"]
        if item.get("Id") != credit or item.get("Image") != state["images"]["credit"] or labels.get("io.agent-hub.h03.owner") != state["owner"] or labels.get("com.docker.compose.project") != state["project"] or labels.get("com.docker.compose.service") != "credit":
            raise Blocked("owned_control_restart_identity_differs")
        return item
    before = inspect()
    launch.docker(["restart", "--time", "10", credit], deployment["env"], deployment["root"])
    after = inspect()
    if not after.get("State", {}).get("Running") or before.get("State", {}).get("StartedAt") == after.get("State", {}).get("StartedAt"):
        raise Blocked("actual_owned_control_restart_not_observed")
    deadline = time.monotonic() + min(timeout, 60)
    while time.monotonic() < deadline:
        try:
            code, _, _ = client.http("GET", "/healthz")
            if code == 200:
                client.reauthenticate(password)
                return {"container_id": credit, "image_id": after["Image"], "started_before": before["State"]["StartedAt"], "started_after": after["State"]["StartedAt"]}
        except Blocked:
            pass
        time.sleep(.5)
    raise Blocked("restarted_original_control_not_ready")


def unknown_check(deployment, client, password, timeout):
    plan, _ = propose(client)
    gateway = deployment["containers"]["gateway"]
    paused = False
    try:
        launch.docker(["pause", gateway], deployment["env"], deployment["root"]); paused = True
        run_id = client.approve(plan)
        before = wait_state(deployment, run_id, {"waiting"}, timeout)
    finally:
        if paused: launch.docker(["unpause", gateway], deployment["env"], deployment["root"])
    run, _ = original_run(before, run_id)
    if run.get("token"): raise Blocked("unknown_original_lease_must_be_closed_before_restart")
    calls = {k: v["remote_call"] for k, v in run["steps"].items() if v.get("remote_call")}
    if len(calls) != 1 or next(iter(calls.values())).get("status") != "unknown": raise Blocked("real_unknown_call_not_observed")
    snapshot = {k: {f: c.get(f) for f in ("node_execution_id", "request_id", "context_id", "input_digest", "remote_task_id")} for k, c in calls.items()}
    sends_before = sum(sum(e.get("method") == "SendMessage" for e in c.get("gateway_exchanges", {}).values()) for c in calls.values())
    restarted = restart_owned_control(deployment, client, password, timeout)
    restored, _ = original_run(state_read(deployment), run_id)
    restored_calls = {k: v["remote_call"] for k, v in restored["steps"].items() if v.get("remote_call")}
    if restored.get("status") != "waiting" or restored.get("token") or any(c.get("status") != "unknown" for c in restored_calls.values()) or {k: {f: c.get(f) for f in ("node_execution_id", "request_id", "context_id", "input_digest", "remote_task_id")} for k, c in restored_calls.items()} != snapshot:
        raise Blocked("restart_did_not_preserve_original_unknown_identity")
    code, _, _ = client.http("POST", "/api/resume", {"run": run_id, "user": "alice", "password": password})
    if code != 200: raise Blocked("original_unknown_resume_rejected")
    after = wait_state(deployment, run_id, {"waiting"}, timeout)
    latest, _ = original_run(after, run_id)
    latest_calls = {k: v["remote_call"] for k, v in latest["steps"].items() if v.get("remote_call")}
    latest_snapshot = {k: {f: c.get(f) for f in ("node_execution_id", "request_id", "context_id", "input_digest", "remote_task_id")} for k, c in latest_calls.items()}
    sends_after = sum(sum(e.get("method") == "SendMessage" for e in c.get("gateway_exchanges", {}).values()) for c in latest_calls.values())
    if snapshot != latest_snapshot or sends_before != 1 or sends_after != sends_before or any(c.get("status") != "unknown" for c in latest_calls.values()) or any(v.get("status") == "done" for v in latest["steps"].values()): raise Blocked("unknown_identity_or_no_redispatch_invariant_failed")
    return {"status": "PASS_ORIGINAL_UNKNOWN_RESTART_NO_REDISPATCH", "run_id": run_id, "send_count_before": sends_before, "send_count_after": sends_after, "identity": snapshot, "control_restart": restarted, "history_preserved": True}


def prepare(output, project, case):
    if not launch.PROJECT.fullmatch(project): raise Blocked("finite_owned_project_required")
    root = evidence_path(output)
    if root.exists(): raise Blocked("fresh_verification_output_required")
    root.mkdir(parents=True)
    value = {"schema_version": "h03-native-verification/v1", "project": project, "case": case, "mode": "PREPARED_NOT_RUN", "source_files": {p.name: launch.digest(p) for p in [Path(__file__), Path(__file__).with_name("launch_h03.py")]},
             "requires": ["Frozen original source/image receipts", "Already started T312 owned deployment", "Original MCP and demo login approval", "Original Java/Runtime/Bus roles", "Independent native outbound observer", "Original WorkSwarm host/browser for UI item"],
             "expected_component_a": 41 if case == "changed" else 40, "core": "NOT_RUN", "unknown": "NOT_RUN", "actual_workswarm_ui": "BLOCKED", "actual_outbound_observation": "BLOCKED", "business_Gold_D2_production": "NOT_RUN"}
    launch.private_write(root / "verification.json", value)
    return root, value


def run(output, project, case, deployment_root, receipts, timeout, *, unknown=False):
    deployment = deployed(project, deployment_root, receipts)
    root = evidence_path(output, deployment["root"])
    root, result = prepare(root, project, case)
    result.update(mode="EXPLICIT_NATIVE_RUN", started_at=now(), fixed_sources=deployment["sources"], container_ids=deployment["containers"])
    launch.private_write(root / "started.json", {k: v for k, v in result.items() if k != "requires"})
    client = Client(); client.initialize(); client.login(deployment["private"]["H03_LOGIN_PASSWORD"])
    result["planning_negatives"] = planning_negatives(client)
    plan, selection = propose(client)
    code, _, _ = client.http("POST", "/api/approve", {"plan": plan["plan_id"], "version": plan["version"], "hash": plan["hash"], "user": "alice", "password": "not-the-generated-password"})
    if code != 401: raise Blocked("wrong_demo_approval_credential_not_rejected")
    run_id = client.approve(plan)
    state = wait_state(deployment, run_id, {"done"}, timeout)
    receipt = formal_receipt(state, run_id, 41 if case == "changed" else 40)
    page = graph_receipt(client.graph(run_id), receipt)
    code, _, body = client.http("GET", "/ui/run/" + run_id)
    if code != 200 or ('/api/graph?run=' + run_id).encode() not in body: raise Blocked("original_run_page_does_not_reference_run")
    result.update(core="PASS_ORIGINAL_ENGINEERING_CHAIN", formal=receipt, selection=selection, original_go_page=page)
    result["native_control_negatives"] = native_negatives(deployment, client, state, run_id)
    (root / "verification.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    if unknown: result["unknown"] = unknown_check(deployment, client, deployment["private"]["H03_LOGIN_PASSWORD"], timeout)
    result.update(finished_at=now(), status="PARTIAL_CORE_PASS_OBSERVER_AND_ORIGINAL_UI_BLOCKED")
    # The started record survives any failure. Only whitelisted formal fields
    # enter the successful report, never raw state/passwords/lease credentials.
    (root / "verification.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    return result


def compare(baseline, changed):
    baseline, changed = launch.owned_path(baseline), launch.owned_path(changed)
    a, b = launch.read_json(baseline), launch.read_json(changed)
    if a.get("case") != "baseline" or b.get("case") != "changed" or a.get("core") != "PASS_ORIGINAL_ENGINEERING_CHAIN" or b.get("core") != "PASS_ORIGINAL_ENGINEERING_CHAIN": raise Blocked("two_actual_original_chain_receipts_required")
    if a["formal"]["plan_hash"] == b["formal"]["plan_hash"] or a["formal"]["roles"]["eng.sum"]["input_digest"] == b["formal"]["roles"]["eng.sum"]["input_digest"] or a["formal"]["roles"]["eng.reconcile"]["artifact_digest"] == b["formal"]["roles"]["eng.reconcile"]["artifact_digest"]:
        raise Blocked("value_change_did_not_change_adopted_consumption")
    return {"status": "PASS_TWO_FIXED_ORIGINAL_CONSUMPTION_CASES", "baseline_receipt_sha256": launch.digest(baseline), "changed_receipt_sha256": launch.digest(changed), "difference": ["5", "6"], "original_ui_and_outbound": "REQUIRE_SEPARATE_NATIVE_EVIDENCE"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project", required=True)
    parser.add_argument("--scenario", choices=["engineering-two-role"], default="engineering-two-role")
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--case", choices=["baseline", "changed"], default="baseline")
    parser.add_argument("--run", action="store_true", help="Explicitly operate an already started owned deployment; never starts services")
    parser.add_argument("--deployment-root")
    parser.add_argument("--image-receipts")
    parser.add_argument("--unknown", action="store_true", help="Explicit owned gateway fault, durable credit restart and original-run resume after the normal run")
    parser.add_argument("--timeout", type=int, default=180)
    parser.add_argument("--compare-baseline", help="Existing owned baseline verification.json; comparison performs no native calls")
    parser.add_argument("--compare-changed", help="Existing owned changed verification.json")
    args = parser.parse_args()
    try:
        if not 40 <= args.timeout <= 300: raise Blocked("bounded_native_timeout_required")
        if args.compare_baseline or args.compare_changed:
            if not args.compare_baseline or not args.compare_changed or args.run or args.unknown: raise Blocked("separate_two_receipt_comparison_required")
            root, value = prepare(args.output_dir, args.project, args.case)
            value["comparison"] = compare(args.compare_baseline, args.compare_changed)
            value["status"] = value["comparison"]["status"]
            (root / "verification.json").write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")
        elif args.run:
            if not args.deployment_root or not args.image_receipts: raise Blocked("existing_owned_deployment_and_original_receipts_required")
            value = run(args.output_dir, args.project, args.case, args.deployment_root, args.image_receipts, args.timeout, unknown=args.unknown)
        else:
            if args.unknown: raise Blocked("unknown_negative_requires_explicit_run")
            _, value = prepare(args.output_dir, args.project, args.case)
    except (Blocked, launch.Blocked, OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        reason = str(error) if isinstance(error, (Blocked, launch.Blocked)) else "invalid_owned_native_prerequisite"
        print(json.dumps({"status": "BLOCKED", "reason": reason, "services_started_by_driver": False}))
        return 2
    print(json.dumps({"status": value.get("status", value["mode"]), "core": value.get("core"), "actual_workswarm_ui": "BLOCKED", "actual_outbound_observation": "BLOCKED", "services_started_by_driver": False}))
    return 0


if __name__ == "__main__": raise SystemExit(main())
