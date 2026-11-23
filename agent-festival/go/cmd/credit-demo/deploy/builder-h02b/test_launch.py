"""Permission, source, descriptor and cleanup boundaries of the local lab."""

import copy
import importlib.util
import json
from pathlib import Path
import socket
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location(
    "launch", Path(__file__).with_name("launch.py")
)
launch = importlib.util.module_from_spec(spec)
spec.loader.exec_module(launch)


def descriptor():
    return {
        "schema_version": "publication-reference/v2",
        "logical_id": "intent-gate",
        "tenant": "h02b",
        "native_agent_id": "candidate-a",
        "artifact_sha256": "a" * 64,
        "capability_version": "1.0.0",
        "deployment_version": "reference-A",
        "instance_id": "owned-A",
        "card_digest": "b" * 64,
        "contract_digest": "c" * 64,
        "source_revision": "d" * 40,
        "manifest_sha256": "e" * 64,
        "wheel_sha256": "f" * 64,
        "acceptance_sha256": "a" * 64,
        "closure_sha256": "b" * 64,
        "original_handoff": {},
        "governance": "unverified",
        "execution_authorized": False,
        "contract_version": "1.0.0",
        "protocol_version": "1.0",
        "address": "127.0.0.1",
        "port": 19001,
        "endpoint_path": "/a2a/jsonrpc/",
        "agent_card": {"name": "candidate-a"},
    }


class LabTests(unittest.TestCase):
    def test_explicit_v2_missing_unsafe_or_old_binding_fails_before_io(self):
        for key, value in (
            ("schema_version", "publication-reference/v1"),
            ("capability_version", ""),
            ("capability_version", "a" * 17),
            ("deployment_version", "a" * 33),
            ("instance_id", "a" * 65),
            ("instance_id", "a/b"),
        ):
            data = descriptor()
            data[key] = value
            with (
                patch.object(launch, "verify_source") as source,
                patch.object(launch, "local_docker") as docker,
            ):
                with self.assertRaises(launch.LabError):
                    launch.prepare("unused", "unused", [data])
                source.assert_not_called()
                docker.assert_not_called()
        data = descriptor()
        del data["instance_id"]
        with self.assertRaises(launch.LabError):
            launch.publication_bindings([data])
        data = descriptor()
        data["unreviewed_default"] = "x"
        with self.assertRaises(launch.LabError):
            launch.publication_bindings([data])

    def test_current_and_explicit_retained_conflicts_fail_before_io(self):
        for mode in ("capability", "instance"):
            current = descriptor()
            old = descriptor()
            if mode == "capability":
                old["artifact_sha256"] = "b" * 64
                old["instance_id"] = "different-instance"
            else:
                old["port"] += 1
            with (
                patch.object(launch, "verify_source") as source,
                patch.object(launch, "local_docker") as docker,
            ):
                with self.assertRaisesRegex(
                    launch.LabError, "publication_binding_conflict"
                ):
                    launch.prepare(
                        "unused", "unused", [current], retained_descriptors=[old]
                    )
                source.assert_not_called()
                docker.assert_not_called()
        current = descriptor()
        launch.publication_bindings([current], [copy.deepcopy(current)])
        with self.assertRaises(launch.LabError):
            launch.publication_bindings([current], [current] * 33)

    def test_shared_v2_fixture_uses_explicit_versions_without_shortening_artifact(self):
        data = json.loads(
            (
                Path(__file__).parents[2] / "testdata/publication-reference-v2.json"
            ).read_text(encoding="utf-8")
        )["descriptor"]
        rows, _ = launch.instances([data], "irrelevant-owner")
        self.assertEqual(rows[0]["capability-version"], data["capability_version"])
        self.assertEqual(rows[0]["deployment-version"], data["deployment_version"])
        self.assertEqual(rows[0]["instance-id"], data["instance_id"])
        self.assertEqual(len(data["artifact_sha256"]), 64)

    def test_candidate_closure_rejects_undeclared_source(self):
        with tempfile.TemporaryDirectory() as tmp:
            with patch.object(
                launch, "run", side_effect=[launch.REVISION, "pom.xml", ""]
            ):
                with self.assertRaisesRegex(launch.LabError, "scope"):
                    launch.source_closure(tmp)

    def test_candidate_closure_covers_new_and_changed_bytes(self):
        relative = "common/agent-bus/registry-discovery-center/src/main/java/com/openjiuwen/rdc/card/CardVersionSource.java"
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / relative
            path.parent.mkdir(parents=True)
            path.write_text("enum first")
            with patch.object(
                launch, "run", side_effect=[launch.REVISION, "", relative]
            ):
                first = launch.source_closure(tmp)
            path.write_text("enum changed")
            with patch.object(
                launch, "run", side_effect=[launch.REVISION, "", relative]
            ):
                second = launch.source_closure(tmp)
            self.assertNotEqual(first["sha256"], second["sha256"])
            self.assertEqual(set(first["files"]), {relative})

    def test_candidate_requires_source_bound_build_and_tests(self):
        with self.assertRaisesRegex(launch.LabError, "required"):
            launch.candidate_receipts(Path("unused"), None, None)
        with tempfile.TemporaryDirectory() as tmp:
            b, t = Path(tmp) / "build.json", Path(tmp) / "tests.json"
            b.write_text(
                json.dumps({"candidate": True, "source_closure": {"sha256": "old"}})
            )
            t.write_text(
                json.dumps({"status": "PASS", "source_closure": {"sha256": "new"}})
            )
            with (
                patch.object(launch, "source_closure", return_value={"sha256": "new"}),
                patch.object(launch, "verify_images") as images,
            ):
                with self.assertRaisesRegex(launch.LabError, "binding"):
                    launch.candidate_receipts(Path(tmp), b, t)
                images.assert_not_called()

    def test_candidate_native_card_unchanged_and_wrong_interface_rejected(self):
        data = descriptor()
        data["agent_card"]["supportedInterfaces"] = [
            {
                "url": "http://127.0.0.1:19001/a2a/jsonrpc/",
                "protocolBinding": "JSONRPC",
                "protocolVersion": "1.0",
            }
        ]
        before = copy.deepcopy(data)
        self.assertEqual(
            launch.candidate_card_check([data])["status"], "SOURCE_PRECHECK_ONLY"
        )
        self.assertEqual(data, before)
        data["agent_card"]["supportedInterfaces"][0]["url"] = "http://evil/a2a"
        with self.assertRaises(launch.LabError):
            launch.candidate_card_check([data])

    def test_critical_suites_cannot_omit_or_skip_governance(self):
        names = [
            "DeploymentVersionSourceConfigurationTest",
            "ReconciliationDeploymentVersionsIntegrationTest",
            "NativeA2aRouteTargetDeriverTest",
            "HttpAgentRuntimeClientTest",
            "AgentCardValidatorTest",
            "AgentCardFetcherTest",
            "AgentCardSignatureVerifierTest",
        ]
        suites = {
            name: {"tests": 1, "failures": 0, "errors": 0, "skipped": 0}
            for name in names
        }
        self.assertTrue(launch.critical_complete(suites))
        suites["AgentCardSignatureVerifierTest"]["skipped"] = 1
        self.assertFalse(launch.critical_complete(suites))
        del suites["AgentCardSignatureVerifierTest"]
        self.assertFalse(launch.critical_complete(suites))

    def test_text_buildx_fallback_still_rejects_remote_and_multiple_nodes(self):
        for endpoint in ["default", "ssh://remote", "default\nEndpoint: ssh://remote"]:
            text = (
                "Name: default\nDriver: docker\nNodes:\nName: default\nEndpoint: "
                + endpoint
                + "\n"
            )
            with (
                patch.dict(launch.os.environ, {}, clear=True),
                patch.object(
                    launch,
                    "local_docker",
                    return_value={
                        "name": "default",
                        "endpoint": "unix:///var/run/docker.sock",
                    },
                ),
                patch.object(
                    launch,
                    "run",
                    side_effect=[launch.LabError("command_failed:docker"), text],
                ),
            ):
                if endpoint == "default":
                    self.assertIn("--builder", launch.local_buildx())
                else:
                    with self.assertRaises(launch.LabError):
                        launch.local_buildx()

    def test_base_transport_uses_only_official_exact_amd64_digest(self):
        import io

        data = {
            "images": [
                {"architecture": "amd64", "os": "linux", "digest": "sha256:" + "a" * 64}
            ]
        }
        with patch.object(
            launch, "urlopen", return_value=io.StringIO(json.dumps(data))
        ) as request:
            pin = launch.original_base_transport()
        self.assertEqual(pin["original_tag"], "python:3.11-slim")
        self.assertEqual(
            pin["mirror_ref"], "docker.m.daocloud.io/library/python@sha256:" + "a" * 64
        )
        self.assertEqual(
            request.call_args.args[0],
            "https://hub.docker.com/v2/repositories/library/python/tags/3.11-slim",
        )
        for images in [
            [],
            data["images"] * 2,
            [{"architecture": "amd64", "os": "linux", "digest": "tag-only"}],
        ]:
            with patch.object(
                launch,
                "urlopen",
                return_value=io.StringIO(json.dumps({"images": images})),
            ):
                with self.assertRaises(launch.LabError):
                    launch.original_base_transport()

    def test_existing_proxy_forwarding_never_puts_values_in_build_argv(self):
        prefix = [
            "docker",
            "--context",
            "default",
            "buildx",
            "build",
            "--builder",
            "default",
        ]
        proxy = "http://127.0.0.1:12345"
        with (
            tempfile.TemporaryDirectory() as tmp,
            patch.dict(launch.os.environ, {"HTTP_PROXY": proxy}, clear=True),
            patch.object(launch, "verify_source", return_value=Path(tmp)),
            patch.object(launch, "local_buildx", return_value=prefix),
            patch.object(launch, "run", return_value="sha256:local"),
            patch.object(
                launch.subprocess,
                "run",
                return_value=launch.subprocess.CompletedProcess([], 0),
            ) as calls,
        ):
            launch.build(tmp, Path(tmp) / "log", use_existing_proxy=True)
            for call in calls.call_args_list:
                args = call.args[0]
                self.assertIn("HTTP_PROXY", args)
                self.assertNotIn(proxy, args)
                self.assertIn("--network", args)
                self.assertIn("host", args)

    def test_existing_proxy_rejects_credentials_and_invalid_values(self):
        for key in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
            for value in (
                "http://user:private@localhost:1234",
                "socks5://localhost:1234",
                "http://localhost:invalid",
                "http://localhost:1234/private?key=value",
                "missing-host",
            ):
                with patch.dict(launch.os.environ, {key: value}, clear=True):
                    with self.assertRaisesRegex(
                        launch.LabError, "^existing_proxy_transport_invalid$"
                    ):
                        launch.existing_proxy_keys()
        with patch.dict(
            launch.os.environ, {"HTTPS_PROXY": "http://localhost:1234"}, clear=True
        ):
            keys, loopback = launch.existing_proxy_keys()
            self.assertEqual({key.upper() for key in keys}, {"HTTPS_PROXY"})
            self.assertTrue(loopback)

    def test_pinned_postgres_mirror_requires_exact_digest_platform_and_id(self):
        valid = {
            "Os": "linux",
            "Architecture": "amd64",
            "Id": "sha256:" + "a" * 64,
            "RepoDigests": [
                "docker.m.daocloud.io/library/postgres@" + launch.POSTGRES_DIGEST
            ],
        }
        with (
            patch.object(launch, "local_docker"),
            patch.object(launch, "run", return_value=json.dumps([valid])),
        ):
            report = launch.postgres_transport()
            self.assertEqual(report["digest"], launch.POSTGRES_DIGEST)
            self.assertEqual(report["mirror_ref"], launch.POSTGRES_MIRROR)
        for key, value in (
            ("Os", "windows"),
            ("Architecture", "arm64"),
            ("Id", "tag-only"),
            (
                "RepoDigests",
                ["docker.m.daocloud.io/library/postgres@sha256:" + "b" * 64],
            ),
        ):
            bad = dict(valid, **{key: value})
            with (
                patch.object(launch, "local_docker"),
                patch.object(launch, "run", return_value=json.dumps([bad])),
            ):
                with self.assertRaisesRegex(
                    launch.LabError, "pinned_postgres_transport_invalid"
                ):
                    launch.postgres_transport()

    def test_pinned_postgres_default_and_opt_in_prepare_are_distinct(self):
        transport = {"mirror_ref": launch.POSTGRES_MIRROR, "id": "sha256:" + "a" * 64}
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            with (
                patch.object(launch, "verify_source", return_value=Path(tmp)),
                patch.object(launch, "ports_free"),
                patch.object(
                    launch, "postgres_transport", return_value=transport
                ) as check,
            ):
                plain = Path(tmp) / "plain"
                launch.prepare(plain, tmp, [descriptor()])
                check.assert_not_called()
                self.assertIn(
                    "H02B_POSTGRES_IMAGE=" + launch.POSTGRES_ORIGINAL,
                    (plain / "private.env").read_text(),
                )
                mirrored = Path(tmp) / "mirrored"
                launch.prepare(
                    mirrored, tmp, [descriptor()], mirror_pinned_postgres=True
                )
                self.assertEqual(check.call_count, 1)
                self.assertIn(
                    launch.POSTGRES_MIRROR, (mirrored / "private.env").read_text()
                )
                self.assertEqual(
                    launch.load_state(mirrored)[1]["postgres_transport"], transport
                )

    def test_start_refuses_changed_pinned_postgres_before_compose(self):
        data = {
            "project": "builder-h02b-" + "a" * 16,
            "source": "unused",
            "postgres_transport": {"id": "old"},
            "reconciliation_compatibility": {"status": "SOURCE_PRECHECK_ONLY"},
        }
        with (
            patch.object(launch, "load_state", return_value=(Path("unused"), data)),
            patch.object(launch, "local_docker"),
            patch.object(launch, "verify_ownership", return_value=[]),
            patch.object(launch, "verify_source"),
            patch.object(launch, "compose_args", return_value=[]),
            patch.object(launch, "postgres_transport", return_value={"id": "changed"}),
            patch.object(launch, "run") as command,
        ):
            with self.assertRaisesRegex(
                launch.LabError, "pinned_postgres_transport_changed"
            ):
                launch.action("up", "unused")
            command.assert_not_called()

    def test_compose_refuses_ambient_image_override(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            (state / "private.env").write_text(
                "H02B_POSTGRES_IMAGE=" + launch.POSTGRES_ORIGINAL
            )
            with patch.dict(
                launch.os.environ, {"H02B_POSTGRES_IMAGE": "unreviewed:latest"}
            ):
                with self.assertRaisesRegex(
                    launch.LabError, "compose_environment_override_forbidden"
                ):
                    launch.compose_args(state, {"project": "owned"})

    def test_failed_maven_report_is_nonzero_cli_exit(self):
        with (
            patch(
                "sys.argv",
                [
                    "launch.py",
                    "test",
                    "--source-dir",
                    "unused",
                    "--state-dir",
                    "unused",
                    "--build-receipt",
                    "unused",
                ],
            ),
            patch.object(launch, "test_candidate", return_value={"status": "FAIL"}),
            patch("builtins.print"),
        ):
            self.assertEqual(launch.main(), 1)

    def test_buildx_remote_override_rejected_before_inspect(self):
        for key in ("BUILDX_BUILDER", "BUILDKIT_HOST"):
            with (
                patch.dict(launch.os.environ, {key: "remote"}, clear=True),
                patch.object(launch, "run") as call,
            ):
                with self.assertRaises(launch.LabError):
                    launch.local_buildx()
                call.assert_not_called()

    def test_buildx_rejects_remote_or_container_nodes(self):
        examples = [
            {
                "Name": "default",
                "Driver": "remote",
                "Nodes": [{"Endpoint": "tcp://remote"}],
            },
            {
                "Name": "default",
                "Driver": "docker-container",
                "Nodes": [{"Endpoint": "default"}],
            },
            {
                "Name": "default",
                "Driver": "docker",
                "Nodes": [{"Endpoint": "other-context"}],
            },
            {
                "Name": "default",
                "Driver": "docker",
                "Nodes": [{"Endpoint": "default"}, {"Endpoint": "ssh://remote"}],
            },
        ]
        for builder in examples:
            with (
                self.subTest(builder=builder),
                patch.dict(launch.os.environ, {}, clear=True),
                patch.object(
                    launch,
                    "local_docker",
                    return_value={
                        "name": "default",
                        "endpoint": "unix:///var/run/docker.sock",
                    },
                ),
                patch.object(launch, "run", return_value=json.dumps(builder)),
            ):
                with self.assertRaises(launch.LabError):
                    launch.local_buildx()

    def test_buildx_pins_inspected_default_without_global_selection(self):
        context = {"name": "default", "endpoint": "unix:///var/run/docker.sock"}
        builder = {
            "Name": "default",
            "Driver": "docker",
            "Nodes": [{"Endpoint": "default"}],
        }
        with (
            patch.dict(launch.os.environ, {}, clear=True),
            patch.object(launch, "local_docker", return_value=context),
            patch.object(launch, "run", return_value=json.dumps(builder)) as call,
        ):
            argv = launch.local_buildx()
        self.assertEqual(
            argv,
            [
                "docker",
                "--context",
                "default",
                "buildx",
                "build",
                "--builder",
                "default",
            ],
        )
        self.assertEqual(
            call.call_args.args[0],
            [
                "docker",
                "--context",
                "default",
                "buildx",
                "inspect",
                "default",
                "--format",
                "{{json .}}",
            ],
        )
        self.assertNotIn("use", call.call_args.args[0])

    def test_buildx_unsupported_context_fails_before_builder_query(self):
        with (
            patch.dict(launch.os.environ, {}, clear=True),
            patch.object(
                launch,
                "local_docker",
                return_value={"name": "desktop-linux", "endpoint": "unix:///socket"},
            ),
            patch.object(launch, "run") as call,
        ):
            with self.assertRaises(launch.LabError):
                launch.local_buildx()
            call.assert_not_called()

    def test_build_rejects_remote_node_before_upload(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            with (
                patch.object(launch, "verify_source", return_value=Path(tmp)),
                patch.object(
                    launch,
                    "local_buildx",
                    side_effect=launch.LabError("local_daemon_buildx_required"),
                ),
                patch.object(launch.subprocess, "run") as upload,
            ):
                with self.assertRaises(launch.LabError):
                    launch.build(tmp, Path(tmp) / "logs")
                upload.assert_not_called()
                self.assertFalse((Path(tmp) / "logs").exists())

    def test_both_image_builds_keep_validated_explicit_builder(self):
        prefix = [
            "docker",
            "--context",
            "default",
            "buildx",
            "build",
            "--builder",
            "default",
        ]
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            with (
                patch.object(launch, "verify_source", return_value=Path(tmp)),
                patch.object(launch, "local_buildx", return_value=prefix),
                patch.object(
                    launch.subprocess,
                    "run",
                    return_value=launch.subprocess.CompletedProcess([], 0),
                ) as upload,
                patch.object(launch, "run", return_value="sha256:local"),
            ):
                launch.build(tmp, Path(tmp) / "logs")
            self.assertEqual(upload.call_count, 2)
            for call in upload.call_args_list:
                self.assertEqual(call.args[0][: len(prefix)], prefix)

    def test_reconciliation_versions_and_card_unchanged(self):
        data = descriptor()
        original = copy.deepcopy(data)
        rows, tenant = launch.instances([data], "builder-h02b-123")
        self.assertEqual(data, original)
        self.assertEqual(rows[0]["capability-version"], "1.0.0")
        self.assertEqual(rows[0]["deployment-version"], "reference-A")
        self.assertEqual(rows[0]["instance-id"], "owned-A")
        self.assertEqual(rows[0]["contract-version"], "1.0.0")
        self.assertEqual(rows[0]["route-key"], "/a2a/jsonrpc/")
        self.assertEqual(rows[0]["service-id"], data["native_agent_id"])
        self.assertEqual(tenant, "h02b")

    def test_rejects_unsafe_descriptor(self):
        for key, value in [
            ("address", "10.0.0.1"),
            ("port", True),
            ("port", 80),
            ("native_agent_id", "a" * 65),
            ("native_agent_id", "a\n"),
            ("artifact_sha256", "a" * 63),
            ("protocol_version", "0.3"),
            ("contract_version", "1.0"),
            ("endpoint_path", "/evil"),
            ("tenant", "bad\nvalue"),
            ("logical_id", "../../unsafe"),
            ("agent_card", {"name": "other"}),
        ]:
            with self.subTest(key=key, value=value):
                data = descriptor()
                data[key] = value
                with self.assertRaises(launch.LabError):
                    launch.instances([data], "owned")

    def test_reject_duplicate_native_and_mixed_tenant(self):
        first = descriptor()
        second = descriptor()
        with self.assertRaises(launch.LabError):
            launch.instances([first, second], "owned")
        second["native_agent_id"] = "candidate-b"
        second["agent_card"]["name"] = "candidate-b"
        second["tenant"] = "other"
        with self.assertRaises(launch.LabError):
            launch.instances([first, second], "owned")

    def test_ports_never_stop_foreign_owner(self):
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
            with self.assertRaises(launch.LabError):
                launch.ports_free([port])
        with self.assertRaises(launch.LabError):
            launch.ports_free([19000, 19000])

    def test_local_docker_only(self):
        with patch.dict(launch.os.environ, {"DOCKER_HOST": "ssh://remote"}):
            with self.assertRaises(launch.LabError):
                launch.local_docker()
        with (
            patch.dict(launch.os.environ, {}, clear=True),
            patch.object(
                launch,
                "run",
                return_value=json.dumps(
                    [{"Endpoints": {"docker": {"Host": "tcp://remote"}}}]
                ),
            ),
        ):
            with self.assertRaises(launch.LabError):
                launch.local_docker()

    def test_source_revision_and_dirty_checks(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            with patch.object(launch, "run", return_value="other"):
                with self.assertRaises(launch.LabError):
                    launch.verify_source(tmp)
            with patch.object(
                launch, "run", side_effect=[launch.REVISION, " M code.java"]
            ):
                with self.assertRaises(launch.LabError):
                    launch.verify_source(tmp)

    def prepared(self, tmp):
        state = Path(tmp) / "state"
        with (
            patch.object(launch, "verify_source", return_value=Path(tmp)),
            patch.object(launch, "ports_free"),
        ):
            result = launch.prepare(state, tmp, [descriptor()])
        self.assertEqual(result["registration"], "NOT_RUN")
        self.assertNotIn("TOKEN", json.dumps(result))
        return state

    def test_prepare_no_subprocess_and_tamper_rejected(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = self.prepared(tmp)
            loaded, data = launch.load_state(state)
            self.assertEqual(loaded, state)
            self.assertIn(
                "registration-paths: [reconciliation]", (state / "rdc.yml").read_text()
            )
            self.assertNotIn("push", (state / "rdc.yml").read_text())
            self.assertIn("${RDC_DB_PASSWORD}", (state / "rdc.yml").read_text())
            with (state / "rdc.yml").open("a") as out:
                out.write("tamper")
            with self.assertRaises(launch.LabError):
                launch.load_state(state)

    def test_gateway_card_public_address_tracks_selected_port(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = Path(tmp) / "state"
            with (
                patch.object(launch, "verify_source", return_value=Path(tmp)),
                patch.object(launch, "ports_free"),
            ):
                launch.prepare(state, tmp, [descriptor()], gateway_port=19477)
            env = dict(
                line.split("=", 1)
                for line in (state / "private.env").read_text().splitlines()
            )
            self.assertEqual(env["H02B_GATEWAY_PORT"], "19477")
            self.assertIn(
                "GATEWAY_PUBLIC_BASE_URL: http://127.0.0.1:${H02B_GATEWAY_PORT:?required}",
                (state / "compose.yaml").read_text(),
            )

    def test_prepare_existing_state_refused(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = self.prepared(tmp)
            with (
                patch.object(launch, "verify_source", return_value=Path(tmp)),
                patch.object(launch, "ports_free"),
            ):
                with self.assertRaises(launch.LabError):
                    launch.prepare(state, tmp, [descriptor()])

    def test_original_native_card_blocked_without_modifying_it(self):
        data = descriptor()
        original = copy.deepcopy(data)
        gate = launch.reconciliation_compatibility([data])
        self.assertEqual(gate["status"], "BLOCKED")
        self.assertIn("ORIGINAL_CARD_VERSION_REQUIRED", gate["reasons"])
        self.assertEqual(data, original)
        # Adding a version alone would still leave the original fetched protocol
        # version overriding the operator's business contract default.
        data["agent_card"]["version"] = "a" * 64
        data["agent_card"]["supportedInterfaces"] = [{"protocolVersion": "1.0"}]
        gate = launch.reconciliation_compatibility([data])
        self.assertEqual(gate["status"], "BLOCKED")
        self.assertEqual(
            gate["reasons"], ["ORIGINAL_FETCHED_CONTRACT_VERSION_MISMATCH"]
        )

    def test_incompatible_prepare_cannot_start_containers(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = self.prepared(tmp)
            with (
                patch.object(launch, "local_docker"),
                patch.object(launch, "verify_ownership", return_value=[]),
                patch.object(launch, "run") as call,
            ):
                with self.assertRaises(launch.LabError):
                    launch.action("up", state)
            call.assert_not_called()

    def test_cleanup_keeps_volumes_and_history(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = self.prepared(tmp)
            with (
                patch.object(launch, "local_docker"),
                patch.object(launch, "verify_ownership", return_value=[]),
                patch.object(launch, "run") as call,
            ):
                result = launch.action("down", state)
            argv = call.call_args.args[0]
            self.assertIn("down", argv)
            self.assertNotIn("-v", argv)
            self.assertNotIn("--volumes", argv)
            self.assertTrue((state / "ownership.json").exists())
            self.assertEqual(result["volumes"], "RETAINED")

    def test_foreign_project_member_rejected(self):
        with patch.object(
            launch,
            "run",
            side_effect=[
                "id1",
                json.dumps(
                    [{"Config": {"Labels": {"io.agent-hub.h02b.owner": "other"}}}]
                ),
            ],
        ):
            with self.assertRaises(launch.LabError):
                launch.verify_ownership({"project": "builder-h02b-123"})

    def test_up_checks_ports_and_does_not_build_or_send(self):
        with tempfile.TemporaryDirectory(dir=Path(__file__).parent) as tmp:
            state = self.prepared(tmp)
            receipt = state / "ownership.json"
            data = json.loads(receipt.read_text())
            # Exercise the post-compatibility lifecycle boundary independently.
            data["reconciliation_compatibility"] = {"status": "SOURCE_PRECHECK_ONLY"}
            receipt.write_text(json.dumps(data))
            with (
                patch.object(launch, "local_docker"),
                patch.object(launch, "verify_ownership", return_value=[]),
                patch.object(launch, "verify_source"),
                patch.object(launch, "ports_free") as ports,
                patch.object(launch, "run") as call,
            ):
                result = launch.action("up", state)
            ports.assert_called_once()
            self.assertEqual(result["status"], "STARTED_REGISTRATION_UNVERIFIED")
            self.assertIn("--no-build", call.call_args.args[0])
            self.assertEqual(call.call_count, 2)


if __name__ == "__main__":
    unittest.main()
