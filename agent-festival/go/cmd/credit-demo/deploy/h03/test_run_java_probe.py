"""Finite offline runner boundaries; no Maven, Java or Docker process runs."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

spec = importlib.util.spec_from_file_location("h03_java_probe", Path(__file__).with_name("run_java_probe.py"))
probe = importlib.util.module_from_spec(spec)
spec.loader.exec_module(probe)


class JavaProbeTests(unittest.TestCase):
    def setUp(self):
        evidence = probe.FESTIVAL.parent / ".specify/feature/HUB-NEXT/evidence/.pytest-tmp/T312-probe"
        evidence.mkdir(parents=True, exist_ok=True)
        self.tmp = tempfile.TemporaryDirectory(dir=evidence)
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.festival = self.root / "agent-festival"
        lock = self.festival / "go/cmd/credit-demo/deploy/sources.lock.json"
        lock.parent.mkdir(parents=True)
        lock.write_text("{}", encoding="utf-8")
        self.module, self.name = next(iter(probe.PROBES.items()))
        self.folder = self.festival / self.module
        source = self.folder / "src/test/java" / (self.name.replace(".", "/") + ".java")
        source.parent.mkdir(parents=True)
        source.write_text("// finite probe fixture", encoding="utf-8")
        (self.folder / "pom.xml").write_text("<project/>", encoding="utf-8")
        self.cache = self.root / "old-cache"
        self.cache.mkdir()
        (self.cache / "original.jar").write_bytes(b"existing artifact")
        self.output = self.root / ".specify/feature/HUB-NEXT/evidence/fresh"

    def test_rejects_main_other_module_and_other_probe(self):
        for module, name in [("../outside", self.name), (self.module, "com.openjiuwen.demo.credit.Main"),
                             (self.module, "com.openjiuwen.demo.professional.H03EngineeringProbe")]:
            with self.assertRaises(probe.ProbeBlocked):
                probe.validate(module, name, self.output, self.festival)

    def test_output_must_be_fresh_and_owned(self):
        for output in [self.root / "outside", self.output.parent, self.output.parent / "bad,mount"]:
            if output == self.output.parent:
                output.mkdir(parents=True)
            with self.assertRaises(probe.ProbeBlocked):
                probe.validate(self.module, self.name, output, self.festival)

    def test_missing_probe_and_cache_are_blocked_before_process(self):
        with patch.object(probe, "call") as call:
            with self.assertRaises(probe.ProbeBlocked):
                probe.run(self.module, self.name, self.output, self.root / "missing", festival=self.festival)
            call.assert_not_called()

    def test_environment_drops_model_bank_and_java_injection(self):
        env = probe.clean_environment({"PATH": "tools", "SWARM_API_KEY": "private",
              "BANK_PASSWORD": "private", "JAVA_TOOL_OPTIONS": "injected", "HOME": "private-home"})
        self.assertEqual(set(env), {"PATH", "PYTHONUTF8", "PYTHONDONTWRITEBYTECODE"})

    def fake_call(self, argv, output, name, env, **kwargs):
        self.commands.append(argv)
        if name == "maven":
            self.assertIn("-o", argv)
            self.assertIn("-DskipTests", argv)
            target = output / "module/target"
            target.mkdir()
            with zipfile.ZipFile(target / "original.jar", "w") as z:
                z.writestr("BOOT-INF/lib/original-lib.jar", b"existing-library")
            return "package complete"
        if name == "probe":
            self.assertEqual(argv[-1], self.name)
            self.assertNotIn("Main", argv[-1])
            return "H03_PROBE_PASS" if self.marker else "no assertions"
        raise AssertionError(name)

    def test_host_offline_probe_runs_only_exact_class_from_copied_source(self):
        self.commands, self.marker = [], True
        with patch.object(probe.shutil, "which", return_value="existing-tool"), patch.object(probe, "call", side_effect=self.fake_call):
            result = probe.run(self.module, self.name, self.output, self.cache, festival=self.festival)
        self.assertEqual(result["status"], "PASS")
        self.assertEqual(result["runtime_acceptance"], "NOT_RUN")
        self.assertEqual(len(self.commands), 2)
        self.assertFalse((self.folder / "target").exists())
        self.assertTrue((self.output / "source-manifest.json").is_file())
        self.assertTrue((self.output / "m2/original.jar").is_file())

    def test_zero_exit_without_assertion_marker_is_not_pass(self):
        self.commands, self.marker = [], False
        with patch.object(probe.shutil, "which", return_value="existing-tool"), patch.object(probe, "call", side_effect=self.fake_call):
            with self.assertRaisesRegex(probe.ProbeBlocked, "assertion_receipt"):
                probe.run(self.module, self.name, self.output, self.cache, festival=self.festival)
        self.assertFalse((self.output / "result.json").exists())

    def test_extract_rejects_zip_path_escape(self):
        jar = self.root / "bad.jar"
        with zipfile.ZipFile(jar, "w") as z:
            z.writestr("BOOT-INF/lib/../../escape.jar", "bad")
        with self.assertRaises(probe.ProbeBlocked):
            probe.extract_libraries(jar, self.root / "libs")
        self.assertFalse((self.root / "escape.jar").exists())

    def test_mutable_image_tag_is_not_accepted(self):
        with patch.object(probe, "call") as call:
            with self.assertRaises(probe.ProbeBlocked):
                probe.run(self.module, self.name, self.output, self.cache,
                          image="some-image:latest", festival=self.festival)
            call.assert_not_called()

    def docker_call(self, argv, output, name, env, **kwargs):
        if name == "docker-context":
            return json.dumps([{"Endpoints": {"docker": {"Host": "unix:///var/run/docker.sock"}}}])
        if name == "image-inspect":
            self.assertFalse(kwargs["retain"])
            return json.dumps([{"Id": "sha256:" + "a" * 64, "Os": "linux", "Config": {"Env": self.image_env}}])
        self.assertIn("--network", argv)
        self.assertEqual(argv[argv.index("--network") + 1], "none")
        self.assertIn("--read-only", argv)
        self.assertIn("--cap-drop", argv)
        return self.fake_call(argv, output, name, env, **kwargs)

    def test_image_runner_uses_offline_network_and_exact_probe(self):
        self.commands, self.marker, self.image_env = [], True, ["PATH=/usr/bin", "MAVEN_OPTS=-Xmx768m"]
        with patch.object(probe.sys, "platform", "linux"), patch.object(probe.os, "getuid", return_value=1000, create=True), patch.object(probe.os, "getgid", return_value=1000, create=True), patch.dict(probe.os.environ, {}, clear=True), patch.object(probe, "call", side_effect=self.docker_call):
            result = probe.run(self.module, self.name, self.output, self.cache,
                               image="sha256:" + "a" * 64, festival=self.festival)
        self.assertEqual(result["network"], "none")
        self.assertEqual(self.commands[1][-1], self.name)

    def test_image_secret_is_rejected_before_any_container_without_logging_value(self):
        self.commands, self.marker, self.image_env = [], True, ["SWARM_API_KEY=synthetic-private"]
        with patch.object(probe.sys, "platform", "linux"), patch.dict(probe.os.environ, {}, clear=True), patch.object(probe, "call", side_effect=self.docker_call):
            with self.assertRaisesRegex(probe.ProbeBlocked, "unexpected_build_image"):
                probe.run(self.module, self.name, self.output, self.cache,
                           image="sha256:" + "a" * 64, festival=self.festival)
        self.assertEqual(self.commands, [])
        self.assertFalse((self.output / "image-inspect.log").exists())


if __name__ == "__main__":
    unittest.main()
