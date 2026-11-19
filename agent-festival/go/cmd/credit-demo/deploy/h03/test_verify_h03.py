"""Pure stdlib verifier boundaries; fixtures never qualify a native runtime."""
import copy
import importlib.util
import json
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
spec = importlib.util.spec_from_file_location("h03_verifier", HERE / "verify_h03.py")
verify = importlib.util.module_from_spec(spec); spec.loader.exec_module(verify)


def fixture(component=40):
    profile = verify.launch.read_json(HERE / "engineering-profile.json")
    binding = {"schema_version": "hub-node-binding/v1", "profile": "two-role-identity/v1", "bus_tenant": "fixture-tenant",
               "materials": [{"id": "engineering-material-a", "sha256": "d" * 64}],
               "node_contracts": {r: profile["candidate_contracts"]["eng." + r + "@0.1.0"] for r in ("sum", "reconcile")}}
    plan_hash = "sha256:" + ("1" if component == 40 else "2") * 64
    plan = {"id": "P1", "version": 1, "hash": plan_hash, "execution_binding": binding}
    run = {"id": "R1", "plan_id": "P1", "plan_version": 1, "plan_hash": plan_hash, "status": "done", "approved_by": "alice",
           "approved_via": {"channel": "confirm-page"}, "exec_binding": {"model": "engineering-fixture"}, "order": ["sum", "reconcile"], "steps": {}, "token": "never_publish_lease"}
    def fields(specs, values): return [{**s, "value": values[s["name"]]} for s in sorted(specs, key=lambda s: s["name"])]
    for role in ("sum", "reconcile"):
        contract = binding["node_contracts"][role]
        node = verify.sha("R1", "P1-v1", role, "iteration:0")
        inputs = {"component_a": str(component), "component_b": "60", "reference": "95"} if role == "sum" else {"reference": "95", "total": str(component + 60)}
        deps = [{"kind": "material", "id": "engineering-material-a", "sha256": "sha256:" + "d" * 64}] if role == "sum" else [{"kind": "step", "id": "sum", "node_execution_id": run["steps"]["sum"]["remote_call"]["node_execution_id"], "artifact_digest": run["steps"]["sum"]["remote_call"]["artifact_digest"]}]
        payload = {"schema_version": "hub-bound-input/v1", "execution_binding_sha256": verify.raw_hash(binding), "node_contract_sha256": verify.raw_hash(contract), "plan_hash": plan_hash,
                   "run_id": "R1", "node_execution_id": node, "node_path": role, "attempt": 1, "fields": fields(contract["input_fields"], inputs), "dependencies": deps}
        remote = {"agent_id": "h03-engineering-" + role, "version": "0.1.0", "card_digest": "sha256:" + "c" * 64, "contract_digest": "sha256:" + "e" * 64}
        raw_input = verify.launch.canonical(payload).decode()
        digest = verify.sha(verify.launch.canonical({"input": raw_input, "goal": "工程。范围：" + role, "binding": remote, "model": "engineering-fixture"}).decode())
        outputs = {"reference": "95", "total": str(component + 60)} if role == "sum" else {"difference": str(component - 35)}
        output = verify.launch.canonical({"schema_version": "hub-node-output/v1", "execution_binding_sha256": payload["execution_binding_sha256"], "node_contract_sha256": payload["node_contract_sha256"], "input_digest": digest, "fields": fields(contract["output_fields"], outputs)}).decode()
        call = {"status": "completed", "attempt": 1, "input": raw_input, "input_digest": digest, "goal": "工程。范围：" + role, "graph_revision": "P1-v1", "node_execution_id": node,
                "binding": remote, "context_id": "fixture-context-" + role, "remote_task_id": "fixture-task-" + role, "agent_started_at": "fixture-start", "request_id": "fixture-request-" + role,
                "artifact_id": "fixture-artifact-" + role, "artifact_digest": verify.sha(output), "model_events": {}, "capability_token": "never_publish_capability",
                "gateway_exchanges": {"fixture-request-" + role: {"method": "SendMessage", "outcome": "task_received", "http_status": 200}}}
        run["steps"][role] = {"id": role, "expert": "eng." + role, "status": "done", "output": "[step:" + role + "]\n" + output, "remote_call": call}
    return {"plans": {"P1": [plan]}, "runs": {"R1": run}}


class VerifierTests(unittest.TestCase):
    def setUp(self):
        base = verify.launch.FESTIVAL.parent / ".specify/feature/HUB-NEXT/evidence/.pytest-tmp/T313"
        base.mkdir(parents=True, exist_ok=True)
        self.temp = tempfile.TemporaryDirectory(dir=base); self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def test_default_prepare_never_constructs_client_or_calls_docker(self):
        with patch.object(verify, "Client") as client, patch.object(verify.launch, "docker") as docker:
            output, value = verify.prepare(self.root / "new", "h03-fixture", "baseline")
        client.assert_not_called(); docker.assert_not_called()
        self.assertEqual(value["core"], "NOT_RUN")
        self.assertEqual(value["actual_outbound_observation"], "BLOCKED")
        with self.assertRaises(verify.Blocked): verify.prepare(output, "h03-fixture", "baseline")

    def test_explicit_run_missing_receipts_blocks_before_any_docker(self):
        args = ["verify_h03.py", "--run", "--project", "h03-fixture", "--output-dir", str(self.root / "new")]
        with patch.object(sys, "argv", args), patch.object(verify.launch, "docker") as docker, patch("builtins.print"):
            self.assertEqual(verify.main(), 2)
        docker.assert_not_called()

    def test_documented_default_cli_is_prepare_only(self):
        args = ["verify_h03.py", "--project", "h03-engineering", "--scenario", "engineering-two-role", "--output-dir", str(self.root / "cli")]
        with patch.object(sys, "argv", args), patch.object(verify.launch, "docker") as docker, patch.object(verify, "Client") as client, patch("builtins.print"):
            self.assertEqual(verify.main(), 0)
        docker.assert_not_called(); client.assert_not_called()
        self.assertEqual(json.loads((self.root / "cli/verification.json").read_text(encoding="utf-8"))["mode"], "PREPARED_NOT_RUN")

    def test_strict_json_and_cross_root_output_rejected(self):
        for raw in [b'{"x":1,"x":2}', b'{} {}', b'{"x":NaN}']:
            with self.assertRaises((verify.Blocked, verify.launch.Blocked)): verify.strict(raw)
        with self.assertRaises(verify.launch.Blocked): verify.evidence_path(Path.home() / "outside")
        with self.assertRaises(verify.Blocked): verify.evidence_path(self.root / "inside", self.root)

    def test_original_mcp_json_and_sse_identity_and_session(self):
        client = verify.Client(); calls = []
        def http(method, path, value, headers):
            calls.append((method, path, value, headers))
            result = {"jsonrpc": "2.0", "id": value.get("id"), "result": {"ok": True}}
            if "id" not in value: return 202, {}, b""
            return 200, {"Mcp-Session-Id": "fixture-session"}, ("event: message\ndata: " + json.dumps(result) + "\n\n").encode()
        with patch.object(client, "http", side_effect=http): client.initialize(); self.assertEqual(client.rpc("tools/list"), {"ok": True})
        self.assertEqual(calls[-1][3]["Mcp-Session-Id"], "fixture-session")
        self.assertEqual(calls[1][2]["method"], "notifications/initialized")
        with self.assertRaises(verify.Blocked): client.rpc_body(b'data: {}\n\ndata: {}\n\n')
        with patch.object(client, "http", return_value=(200, {}, b'{"jsonrpc":"2.0","id":999,"result":{}}')):
            with self.assertRaises(verify.Blocked): client.rpc("tools/list")

    def test_login_and_original_form_approval_never_follows_remote_redirect(self):
        client = verify.Client(); client.cookies = [SimpleNamespace(name="cd_session")]
        plan = {"plan_id": "P1", "version": 1, "hash": "sha256:" + "1" * 64}
        with patch.object(client, "http", side_effect=[(303, {"Location": "/ui/"}, b""), (303, {"Location": "/ui/run/R1"}, b"")]) as http:
            client.login("synthetic-only"); self.assertEqual(client.approve(plan), "R1")
        self.assertEqual(http.call_args.args[1], "/ui/plan/P1/v1/approve")
        with patch.object(client, "http", return_value=(303, {"Location": "https://foreign.invalid"}, b"")):
            with self.assertRaises(verify.Blocked): client.approve(plan)

    def test_formal_two_role_consumption_and_value_change_without_secret_publication(self):
        for a in (40, 41):
            receipt = verify.formal_receipt(fixture(a), "R1", a)
            self.assertEqual(receipt["roles"]["eng.sum"]["send_count"], 1)
            self.assertNotIn("never_publish", json.dumps(receipt))
            self.assertEqual(receipt["observed_framework_model_events"], 0)
            self.assertIn("BLOCKED", receipt["actual_role_model_bank_egress_counters"])
        with self.assertRaises(verify.Blocked): verify.formal_receipt(fixture(41), "R1", 40)

    def test_persisted_step_uses_original_service_tag_and_rejects_http_call_alias(self):
        source = (HERE.parents[1] / "service.go").read_text(encoding="utf-8")
        self.assertRegex(source, r'Call\s+\*RemoteCall\s+`json:"remote_call,omitempty"`')
        state = fixture()
        verify.formal_receipt(state, "R1", 40)
        for step in state["runs"]["R1"]["steps"].values():
            step["call"] = step.pop("remote_call")
        with self.assertRaises(verify.Blocked):
            verify.formal_receipt(state, "R1", 40)

    def test_original_nul_digest_closure_task_context_and_duplicate_send_negatives(self):
        mutations = [lambda s: s["runs"]["R1"].update(approved_via={"channel": "fake"}),
            lambda s: s["runs"]["R1"]["steps"]["sum"]["remote_call"].update(input_digest="sha256:"+"0"*64),
            lambda s: s["runs"]["R1"]["steps"]["reconcile"]["remote_call"].update(context_id="fixture-context-sum"),
            lambda s: s["runs"]["R1"]["steps"]["reconcile"]["remote_call"].update(remote_task_id="fixture-task-sum"),
            lambda s: s["runs"]["R1"]["steps"]["sum"]["remote_call"]["gateway_exchanges"].update(duplicate={"method":"SendMessage"}),
            lambda s: s["runs"]["R1"]["steps"]["sum"]["remote_call"]["model_events"].update(unexpected="fixture"),
            lambda s: s["plans"]["P1"][0]["execution_binding"]["materials"][0].update(sha256="0"*64),
            lambda s: s["runs"]["R1"]["steps"]["sum"]["remote_call"].update(input=s["runs"]["R1"]["steps"]["sum"]["remote_call"]["input"]+"\n")]
        for mutation in mutations:
            state = fixture(); mutation(state)
            with self.assertRaises(verify.Blocked): verify.formal_receipt(state, "R1", 40)

    def test_original_graph_must_bind_same_task_context_and_committed_consumption(self):
        receipt = verify.formal_receipt(fixture(), "R1", 40)
        graph = {"run": {"id": "R1", "status": "done"}, "hash": receipt["plan_hash"], "verify": {"ok": True}, "rev": 9, "nodes": []}
        for role in receipt["roles"].values():
            graph["nodes"].append({"id": role["node"], "status": "done", "remote": {"task_id": role["remote_task_id"], "context_id": role["context_id"]}, "consumption": {"status": "committed", "input_digest": role["input_digest"], "artifact_digest": role["artifact_digest"]}})
        self.assertIn("BLOCKED", verify.graph_receipt(graph, receipt)["workswarm_actual_page"])
        graph["nodes"][0]["remote"]["task_id"] = "other"
        with self.assertRaises(verify.Blocked): verify.graph_receipt(graph, receipt)

    def test_unknown_queries_preserve_identity_and_send_count_and_always_unpause(self):
        state = fixture(); run = state["runs"]["R1"]; run["status"] = "waiting"; run["steps"].pop("reconcile")
        run.pop("token")
        run["steps"]["sum"]["status"] = "pending"; run["steps"]["sum"]["remote_call"]["status"] = "unknown"
        deployment = {"containers": {"gateway": "a"*64}, "root": self.root, "env": {}}
        client = SimpleNamespace(approve=lambda p: "R1", http=lambda *a, **k: (200, {}, b"{}"))
        with patch.object(verify, "propose", return_value=({}, {})), patch.object(verify, "wait_state", return_value=state), patch.object(verify,"restart_owned_control",return_value={"actual":"restart-fixture-only"}) as restart, patch.object(verify,"state_read",return_value=state), patch.object(verify.launch, "docker", return_value="") as docker:
            result = verify.unknown_check(deployment, client, "synthetic", 40)
        restart.assert_called_once()
        self.assertEqual((result["send_count_before"], result["send_count_after"]), (1, 1))
        self.assertEqual([c.args[0][0] for c in docker.call_args_list], ["pause", "unpause"])
        with patch.object(verify, "propose", return_value=({}, {})), patch.object(verify, "wait_state", side_effect=verify.Blocked("fixture")), patch.object(verify.launch, "docker", return_value="") as docker:
            with self.assertRaises(verify.Blocked): verify.unknown_check(deployment, client, "synthetic", 40)
        self.assertEqual(docker.call_args.args[0][0], "unpause")
        before = copy.deepcopy(state); after = copy.deepcopy(state)
        after["runs"]["R1"]["steps"]["sum"]["remote_call"]["input_digest"] = "sha256:" + "0" * 64
        with patch.object(verify, "propose", return_value=({}, {})), patch.object(verify, "wait_state", side_effect=[before, after]), patch.object(verify,"restart_owned_control",return_value={}), patch.object(verify,"state_read",return_value=before), patch.object(verify.launch, "docker", return_value=""):
            with self.assertRaises(verify.Blocked): verify.unknown_check(deployment, client, "synthetic", 40)

    def test_failed_or_prepared_call_is_never_reported_as_unknown_after_restart(self):
        state = fixture(); run = state["runs"]["R1"]; run["status"]="waiting";run.pop("token");run["steps"].pop("reconcile")
        run["steps"]["sum"]["status"]="pending";run["steps"]["sum"]["remote_call"]["status"]="unknown"
        deployment={"containers":{"gateway":"a"*64},"root":self.root,"env":{}}
        client=SimpleNamespace(approve=lambda p:"R1",http=lambda *a,**k:(200,{},b"{}"))
        for stage,bad_status in [("restored","failed"),("latest","dispatching")]:
            changed=copy.deepcopy(state);changed["runs"]["R1"]["steps"]["sum"]["remote_call"]["status"]=bad_status
            with patch.object(verify,"propose",return_value=({},{})),patch.object(verify,"wait_state",side_effect=[state,changed] if stage=="latest" else [state]),patch.object(verify,"restart_owned_control",return_value={}),patch.object(verify,"state_read",return_value=changed if stage=="restored" else state),patch.object(verify.launch,"docker",return_value=""):
                with self.assertRaises(verify.Blocked):verify.unknown_check(deployment,client,"synthetic",40)

    def test_restart_is_actual_owned_container_and_refreshes_original_sessions(self):
        owner,project,image="fixture-owner","h03-fixture","sha256:"+"c"*64
        credit="b"*64
        deployment={"containers":{"credit":credit},"state":{"owner":owner,"project":project,"images":{"credit":image}},"env":{},"root":self.root}
        def item(start):return {"Id":credit,"Image":image,"Config":{"Labels":{"io.agent-hub.h03.owner":owner,"com.docker.compose.project":project,"com.docker.compose.service":"credit"}},"State":{"Running":True,"StartedAt":start}}
        calls=[]
        def docker(args,*unused):
            calls.append(args)
            if args[0]=="restart":return credit
            return json.dumps([item("before" if len(calls)==1 else "after")])
        client=SimpleNamespace(http=lambda *a:(200,{},b"{}"),reauthenticate=lambda password:calls.append(["reauthenticated"]))
        with patch.object(verify.launch,"docker",side_effect=docker):result=verify.restart_owned_control(deployment,client,"synthetic",40)
        self.assertEqual([c[0] for c in calls],["inspect","restart","inspect","reauthenticated"])
        self.assertNotEqual(result["started_before"],result["started_after"])
        wrong=item("wrong");wrong["Config"]["Labels"]["io.agent-hub.h03.owner"]="foreign"
        with patch.object(verify.launch,"docker",return_value=json.dumps([wrong])) as docker:
            with self.assertRaises(verify.Blocked):verify.restart_owned_control(deployment,client,"synthetic",40)
        self.assertEqual(docker.call_count,1)

    def test_native_control_negatives_are_separate_and_preserve_formal_state(self):
        state = fixture(); state["runs"]["R1"].pop("token")
        responses = [401, 404, 404, 404, 409, 409, 409, 409, 409]
        client = SimpleNamespace(http=lambda method,path,*a, **k: (responses.pop(0), {}, b"404 page not found\n" if path=="/debug/state" else b"denied"))
        with patch.object(verify, "state_read", return_value=state):
            result = verify.native_negatives({"private": {"H03_RUNTIME_CREDENTIAL": "never_publish_runtime"}}, client, state, "R1")
        self.assertTrue(result["original_formal_state_unchanged"])
        self.assertIn("SEPARATE", result["interval"])
        self.assertNotIn("never_publish", json.dumps(result))
        self.assertIn("BLOCKED", result["live_lease_idempotency_and_field_validation"])
        denied = [401,404,404,404]
        leaking = SimpleNamespace(http=lambda *a, **k: (denied.pop(0), {}, b'{"token":"must-not-publish"}'))
        with self.assertRaises(verify.Blocked): verify.native_negatives({"private":{}},leaking,state,"R1")

    def test_planning_unknown_or_excluded_never_becomes_an_adopted_graph(self):
        client = SimpleNamespace(tool=lambda *a: {"id": "S1", "outcome": "no_graph", "eligibility": [{"capability": "eng.sum", "result": "excluded", "reasons": ["wrong period"]}]})
        self.assertEqual(len(verify.planning_negatives(client)), 2)
        client.tool = lambda *a: {"id": "S1", "outcome": "graph", "eligibility": [{"capability": "eng.sum", "result": "pass"}]}
        with self.assertRaises(verify.Blocked): verify.planning_negatives(client)

    def test_fixed_original_receipts_require_matching_image_and_source_file_hash(self):
        image = "sha256:"+"a"*64
        receipt = self.root / "image.json"; source = self.root / "source.json"
        receipt.write_text(json.dumps({"image_id":image}),encoding="utf-8"); source.write_text('{"source":"fixture-only"}',encoding="utf-8")
        manifest = self.root / "receipts.json"
        manifest.write_text(json.dumps({"schema_version":"h03-native-image-receipts/v1","images":{"credit":{"image_id":image,"receipt":"image.json","receipt_sha256":verify.launch.digest(receipt),"source_manifest":"source.json","source_manifest_sha256":verify.launch.digest(source)}}}),encoding="utf-8")
        self.assertEqual(verify.source_receipts(manifest,{"credit":image})["credit"]["image_id"],image)
        source.write_text('{"changed":true}',encoding="utf-8")
        with self.assertRaises(verify.Blocked): verify.source_receipts(manifest,{"credit":image})


if __name__ == "__main__": unittest.main()
