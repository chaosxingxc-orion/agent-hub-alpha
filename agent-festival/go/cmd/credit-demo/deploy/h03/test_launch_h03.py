"""Reject unsafe deployment configuration before any real Docker operation."""

import copy
import importlib.util
import json
import shlex
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("h03_launcher", Path(__file__).with_name("launch_h03.py"))
launch = importlib.util.module_from_spec(spec)
spec.loader.exec_module(launch)


class LauncherTests(unittest.TestCase):
    def setUp(self):
        home = launch.FESTIVAL.parent / ".specify/feature/HUB-NEXT/evidence/.pytest-tmp/T312-launch"
        home.mkdir(parents=True, exist_ok=True)
        self.temp = tempfile.TemporaryDirectory(dir=home)
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.festival = self.root / "agent-festival"
        self.lock = self.festival / "go/cmd/credit-demo/deploy/sources.lock.json"
        self.lock.parent.mkdir(parents=True)
        self.lock.write_text(json.dumps({"sources": {"workswarm": {"commit": "fixed-original-source"}}}), encoding="utf-8")
        self.output = self.root / ".specify/feature/HUB-NEXT/evidence/deployment"
        self.images = {name: "sha256:" + f"{i + 1:064x}" for i, name in enumerate(launch.BASE)}
        self.manifest = self.root / "images.json"
        self.manifest.write_text(json.dumps({"schema_version": "h03-local-images/v1", "source_lock_sha256": launch.digest(self.lock), "images": self.images}), encoding="utf-8")
        self.calls, self.running, self.containers = [], False, []
        self.image_extra, self.resolved_mutation = [], None

    def image(self, name):
        jars = {"orchestrator": "/app/credit-orchestrator.jar", "engineering-sum": "/app/agent.jar",
                "engineering-reconcile": "/app/agent.jar", "rdc": "/app/rdc.jar", "gateway": "/app/gateway.jar"}
        if name in jars:
            entry, cmd = ["java", "-Xmx192m", "-jar", jars[name]], None
        elif name in ("swarm", "front"):
            entry, cmd = (["/usr/local/bin/aha-entrypoint.sh"], None) if name == "swarm" else (["/docker-entrypoint.sh"], ["nginx", "-g", "daemon off;"])
        else:
            entry, cmd = (["/opt/aha/run/credit/credit-demo"], None) if name == "credit" else (["docker-entrypoint.sh"], ["postgres"])
        return {"Id": self.images[name], "Os": "linux", "Config": {"Env": ["PATH=/usr/bin", *self.image_extra], "Entrypoint": entry, "Cmd": cmd}}

    def fake_docker(self, argv, env, cwd):
        self.calls.append(argv)
        self.assertNotIn("SWARM_API_KEY", env)
        self.assertNotIn("HOME", env)
        self.assertEqual(env["COMPOSE_DISABLE_ENV_FILE"], "true")
        self.assertFalse((cwd / ".env").exists())
        if argv[:3] == ["context", "inspect", "default"]:
            return json.dumps([{"Endpoints": {"docker": {"Host": "unix:///var/run/docker.sock"}}}])
        if argv[:2] == ["image", "inspect"]:
            name = next(k for k, v in self.images.items() if v == argv[2])
            return json.dumps([self.image(name)])
        if argv[0] == "compose" and argv[-3:] == ["config", "--format", "json"]:
            raw = launch.read_json(cwd / "compose.h03.yaml")
            self.expected = launch.expand(raw, env)
            config = copy.deepcopy(self.expected)
            config["name"] = "h03-engineering"
            for service in config["services"].values():
                service.update(command=None, entrypoint=None)
                if service.get("environment") == {}:
                    service.pop("environment")
                service["networks"] = {"h03": None}
                if "depends_on" in service:
                    service["depends_on"] = {k: {"condition": "service_started", "required": True} for k in service["depends_on"]}
            config["networks"]["h03"]["name"] = "h03-engineering_h03"
            for name, value in config["volumes"].items():
                value["name"] = "h03-engineering_" + name
            if self.resolved_mutation:
                self.resolved_mutation(config)
            return json.dumps(config)
        if argv[0] == "compose" and "up" in argv:
            self.assertEqual(argv[-7:], list(launch.BASE))
            self.assertIn("--no-deps", argv)
            self.assertIn("--no-build", argv)
            self.assertEqual(argv[argv.index("--pull") + 1], "never")
            self.running = True
            self.make_containers(env)
            return ""
        if argv[0] == "compose" and "down" in argv:
            self.assertNotIn("-v", argv)
            self.assertNotIn("--volumes", argv)
            self.running = False
            return ""
        if argv[:3] == ["ps", "--all", "--quiet"]:
            return "\n".join(f"{i + 1:012x}" for i in range(7)) if self.running else ""
        if argv[:2] in (["volume", "ls"], ["network", "ls"]):
            return ""
        if argv[0] == "inspect":
            return json.dumps(self.containers)
        if argv[:2] == ["network", "inspect"]:
            return json.dumps([{"Internal": True, "Labels": {"io.agent-hub.h03.owner": env["H03_OWNER"]}}])
        raise AssertionError(argv)

    def make_containers(self, env):
        self.containers = []
        for name, service in self.expected["services"].items():
            labels = {**service["labels"], "com.docker.compose.project": "h03-engineering", "com.docker.compose.service": name}
            values = {"PATH": "/usr/bin", **service["environment"]}
            mounts = [{"Type": m["type"], "Destination": m["target"], "RW": not m.get("read_only", False),
                       **({"Source": m["source"]} if m["type"] == "bind" else {"Name": "h03-engineering_" + m["source"]})}
                      for m in service.get("volumes", [])]
            self.containers.append({"Image": self.images[name], "Config": {"Labels": labels, "Env": [k + "=" + v for k, v in values.items()], "User": service.get("user", "")},
                "HostConfig": {"Privileged": False, "NetworkMode": "h03-engineering_h03", "CapDrop": ["ALL"],
                               "SecurityOpt": ["no-new-privileges:true"],
                               "PortBindings": {str(p["target"]) + "/" + p["protocol"]: [{"HostIp": p["host_ip"], "HostPort": p["published"]}] for p in service.get("ports", [])}},
                "NetworkSettings": {"Networks": {"h03-engineering_h03": {}}}, "Mounts": mounts})

    def prepare(self):
        return launch.prepare("h03-engineering", "none", self.output, self.manifest, festival=self.festival)

    def test_config_has_exact_seven_services_without_start_and_secret_disclosure(self):
        with patch.dict(launch.os.environ, {"PATH": "tools"}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            root, state, env, expected, _ = self.prepare()
        self.assertEqual(set(expected["services"]), set(launch.BASE))
        self.assertFalse(self.running)
        self.assertFalse(any("up" in a for a in self.calls))
        evidence = (root / "configuration-audit.json").read_text(encoding="utf-8")
        self.assertNotIn(env["H03_RUNTIME_CREDENTIAL"], evidence)
        self.assertEqual(expected["services"]["engineering-sum"]["environment"]["H03_ENGINEERING_ROLE"], "sum")
        self.assertNotIn("H03_RUNTIME_CREDENTIAL", expected["services"]["engineering-sum"]["environment"])

    def test_raw_extra_services_inheritance_and_unsafe_keys_rejected(self):
        for key, value in [("env_file", "old.env"), ("network_mode", "host"), ("privileged", True),
                           ("extends", {"service": "old"}), ("build", "."), ("devices", ["/dev/x"] )]:
            raw = launch.read_json(launch.HERE / "compose.h03.yaml")
            raw["services"]["credit"][key] = value
            with self.assertRaises(launch.Blocked):
                launch.validate_template(raw)
        raw["include"] = ["old-compose.yml"]
        with self.assertRaises(launch.Blocked):
            launch.validate_template(raw)

    def test_duplicate_json_and_unknown_interpolation_rejected(self):
        path = self.root / "duplicate.json"
        path.write_text('{"x":1,"x":2}', encoding="utf-8")
        with self.assertRaises(launch.Blocked):
            launch.read_json(path)
        with self.assertRaises(launch.Blocked):
            launch.expand("${HOME}", {})

    def test_ambient_model_even_empty_and_remote_overrides_fail_before_docker(self):
        for values in [{"SWARM_API_KEY": "synthetic"}, {"SWARM_MODEL_NAME": ""}, {"DOCKER_HOST": "ssh://remote"}, {"COMPOSE_FILE": "old.yml"}]:
            with patch.dict(launch.os.environ, values, clear=True), patch.object(launch, "docker") as docker:
                with self.assertRaises(launch.Blocked):
                    self.prepare()
                docker.assert_not_called()

    def test_image_credentials_or_unknown_keys_rejected_without_logging_values(self):
        for extra in [["SWARM_API_KEY=synthetic-private"], ["AWS_SECRET_ACCESS_KEY=synthetic-private"]]:
            self.image_extra = extra
            with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
                with self.assertRaises(launch.Blocked):
                    self.prepare()
            self.assertFalse(any("up" in a for a in self.calls))
            self.assertFalse(any("synthetic-private" in p.read_text(encoding="utf-8") for p in self.output.rglob("*.json")))
            # Each failed configuration requires a separate owned directory.
            self.output = self.output.with_name(self.output.name + "-next")

    def test_original_credit_dockerfile_fixed_defaults_only_apply_to_credit(self):
        dockerfile = (launch.HERE.parent / "Dockerfile.credit").read_text(encoding="utf-8")
        declaration = next(line for line in dockerfile.splitlines() if line.startswith("ENV CREDIT_DATA="))
        defaults = dict(part.split("=", 1) for part in shlex.split(declaration)[1:])
        self.assertEqual(defaults, {"CREDIT_DATA": "/data", "CREDIT_ADDR": ":8091"})
        image = self.image("credit")
        image["Config"]["Env"].extend(k + "=" + v for k, v in defaults.items())
        self.assertEqual(launch.image_audit("credit", image, self.images["credit"]), {"PATH": "/usr/bin", **defaults})
        for key in defaults:
            changed = copy.deepcopy(image)
            changed["Config"]["Env"] = [v for v in changed["Config"]["Env"] if not v.startswith(key + "=")]
            changed["Config"]["Env"].append(key + "=wrong")
            with self.assertRaises(launch.Blocked): launch.image_audit("credit", changed, self.images["credit"])
        other = self.image("orchestrator")
        other["Config"]["Env"].extend(k + "=" + v for k, v in defaults.items())
        with self.assertRaises(launch.Blocked): launch.image_audit("orchestrator", other, self.images["orchestrator"])

    def test_original_pinned_postgres_nonroot_uid_matches_image(self):
        # Original immutable image id@09e4f20... /usr/bin/id postgres receipt:
        # uid=70(postgres) gid=70(postgres), not the Debian-image uid 999.
        raw = launch.read_json(launch.HERE / "compose.h03.yaml")
        self.assertEqual(raw["services"]["bus-db"]["user"], "70:70")

    def test_original_pinned_postgres_package_declaration_is_service_scoped(self):
        image = self.image("bus-db")
        original = "llvm19-dev \t\tclang19"
        image["Config"]["Env"].append("DOCKER_PG_LLVM_DEPS=" + original)
        self.assertEqual(launch.image_audit("bus-db", image, self.images["bus-db"])["DOCKER_PG_LLVM_DEPS"], original)
        image["Config"]["Env"][-1] = "DOCKER_PG_LLVM_DEPS=wrong"
        with self.assertRaises(launch.Blocked): launch.image_audit("bus-db", image, self.images["bus-db"])
        other = self.image("credit")
        other["Config"]["Env"].append("DOCKER_PG_LLVM_DEPS=" + original)
        with self.assertRaises(launch.Blocked): launch.image_audit("credit", other, self.images["credit"])

    def test_resolved_model_environment_unknown_dependency_and_external_network_rejected(self):
        mutations = [lambda c: c["services"]["credit"]["environment"].update(SWARM_API_KEY="synthetic"),
                     lambda c: c["services"]["credit"].update(command=["unexpected"]),
                     lambda c: c["services"]["credit"].update(entrypoint=["unexpected"]),
                     lambda c: c["services"]["credit"].update(depends_on={"foreign": {}}),
                     lambda c: c["networks"]["h03"].update(internal=False),
                     lambda c: c["services"]["credit"].update(network_mode="host"),
                     lambda c: c["services"]["bus-db"].update(user="999:999")]
        for mutation in mutations:
            self.resolved_mutation = mutation
            with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
                with self.assertRaises(launch.Blocked):
                    self.prepare()
            self.assertFalse(any("up" in a for a in self.calls))
            self.output = self.output.with_name(self.output.name + "-next")

    def test_missing_original_ui_is_blocked_not_replaced(self):
        data = launch.read_json(self.manifest)
        data["images"].update({"swarm": "sha256:" + "a" * 64, "front": "sha256:" + "b" * 64})
        self.manifest.write_text(json.dumps(data), encoding="utf-8")
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker") as docker:
            with self.assertRaises((launch.Blocked, OSError)):
                launch.prepare("h03-engineering", "original", self.output, self.manifest, festival=self.festival)
            docker.assert_not_called()

    def test_original_ui_requires_exact_pinned_plugin_and_nine_services(self):
        plugin = self.festival / "go/cmd/credit-demo/workbench-plugin"
        files = ["extension.py", "extension.yaml", "frontend/dist/app.js", "frontend/dist/index.html"]
        for name in files:
            target = plugin / name
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text("synthetic-plugin-fixture", encoding="utf-8")
        self.images.update(swarm="sha256:" + "a" * 64, front="sha256:" + "b" * 64)
        manifest = launch.read_json(self.manifest)
        manifest["images"] = self.images
        manifest["ui"] = {"host_source_revision": "fixed-original-source", "model_free": True,
                          "plugin_sha256": launch.hashlib.sha256(launch.canonical({name: launch.digest(plugin / name) for name in files})).hexdigest()}
        self.manifest.write_text(json.dumps(manifest), encoding="utf-8")
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            _, _, _, expected, _ = launch.prepare("h03-engineering", "original", self.output, self.manifest, festival=self.festival)
        self.assertEqual(set(expected["services"]), set(launch.BASE + launch.UI))
        self.assertFalse(any("up" in a for a in self.calls))

    def test_explicit_start_uses_exact_allowlist_and_audits_actual_containers(self):
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            prepared = self.prepare()
            result = launch.start(prepared)
        self.assertEqual(result["status"], "STARTED_NATIVE_ACCEPTANCE_NOT_RUN")
        self.assertIn("NOT_RUN", result["outbound_observation"])

    def test_mutated_actual_environment_mount_and_foreign_owner_are_not_qualified(self):
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            prepared = self.prepare()
            self.make_containers(prepared[2])
        for kind in ["env", "mount", "owner", "image", "network", "port", "security", "user"]:
            items = copy.deepcopy(self.containers)
            if kind == "env": items[0]["Config"]["Env"].append("SWARM_API_KEY=synthetic")
            if kind == "mount": items[0]["Mounts"].append({"Source": "/private", "Destination": "/old"})
            if kind == "owner": items[0]["Config"]["Labels"]["io.agent-hub.h03.owner"] = "foreign"
            if kind == "image": items[0]["Image"] = "sha256:" + "0" * 64
            if kind == "network": items[0]["HostConfig"]["NetworkMode"] = "host"
            if kind == "port": items[0]["HostConfig"]["PortBindings"]["8091/tcp"][0]["HostIp"] = "0.0.0.0"
            if kind == "security": items[0]["HostConfig"]["SecurityOpt"] = []
            if kind == "user": next(item for item in items if item["Config"]["Labels"]["com.docker.compose.service"] == "bus-db")["Config"]["User"] = "999:999"
            with self.assertRaises(launch.Blocked):
                launch.container_audit(items, prepared[1], prepared[3])

    def test_changed_snapshot_cannot_start(self):
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            prepared = self.prepare()
            (self.output / "config/engineering-material.json").write_text("{}", encoding="utf-8")
            with self.assertRaises(launch.Blocked):
                launch.start(prepared)
        self.assertFalse(any("up" in a for a in self.calls))

    def test_cleanup_refuses_foreign_owner_and_retains_volumes(self):
        with patch.dict(launch.os.environ, {}, clear=True), patch.object(launch, "docker", side_effect=self.fake_docker):
            prepared = self.prepare()
            launch.start(prepared)
            self.containers[0]["Config"]["Labels"]["io.agent-hub.h03.owner"] = "foreign"
            with self.assertRaises(launch.Blocked):
                launch.stop("h03-engineering", self.output, festival=self.festival)
            self.containers[0]["Config"]["Labels"]["io.agent-hub.h03.owner"] = prepared[1]["owner"]
            result = launch.stop("h03-engineering", self.output, festival=self.festival)
        self.assertIn("VOLUMES_RETAINED", result["status"])


if __name__ == "__main__":
    unittest.main()
