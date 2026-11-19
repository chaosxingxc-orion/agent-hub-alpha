"""Offline, finite plain-main probes; never launch the application's Main."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import zipfile

FESTIVAL = Path(__file__).resolve().parents[5]
PROBES = {
    "examples/credit-orchestrator": "com.openjiuwen.demo.credit.H03BindingProbe",
    "examples/professional-agent-runtime": "com.openjiuwen.demo.professional.H03EngineeringProbe",
}
SAFE_ENV = {"PATH", "JAVA_HOME", "SystemRoot", "WINDIR", "TEMP", "TMP", "TMPDIR"}
IMAGE_ENV = {
    "PATH", "LANG", "LC_ALL", "JAVA_HOME", "MAVEN_HOME", "M2_HOME", "MAVEN_OPTS",
    "GPG_KEY", "PYTHON_VERSION", "PYTHON_SHA256", "PYTHON_PIP_VERSION",
    "PYTHON_SETUPTOOLS_VERSION", "PYTHON_GET_PIP_URL", "PYTHON_GET_PIP_SHA256",
}


class ProbeBlocked(ValueError):
    """Finite prerequisite failure, never an implicit dependency download."""


def clean_environment(environ=None):
    source = os.environ if environ is None else environ
    return {**{k: v for k, v in source.items() if k in SAFE_ENV},
            "PYTHONDONTWRITEBYTECODE": "1", "PYTHONUTF8": "1"}


def no_links(path):
    if any(p.is_symlink() for p in [path, *path.parents]):
        raise ProbeBlocked("symlink_path_forbidden")


def validate(module, probe, output, festival=FESTIVAL):
    if PROBES.get(module) != probe:
        raise ProbeBlocked("finite_probe_required")
    folder = festival / module
    no_links(folder)
    source = folder / "src/test/java" / (probe.replace(".", "/") + ".java")
    if not (folder / "pom.xml").is_file() or not source.is_file():
        raise ProbeBlocked("original_module_and_probe_required")
    destination = Path(output).absolute()
    if any(c in str(destination) for c in ",\r\n"):
        raise ProbeBlocked("unsafe_mount_path")
    no_links(destination)
    allowed = festival.parent / ".specify/feature/HUB-NEXT/evidence"
    try:
        destination.resolve().relative_to(allowed.resolve())
    except ValueError:
        raise ProbeBlocked("output_outside_owned_evidence") from None
    if destination.resolve() == allowed.resolve() or destination.exists():
        raise ProbeBlocked("fresh_probe_output_required")
    return folder, destination


def call(argv, output, name, env, *, cwd=None, retain=True):
    result = subprocess.run(argv, capture_output=True, text=True, encoding="utf-8",
                            errors="replace", env=env, cwd=cwd, timeout=900)
    if retain:
        (output / (name + ".log")).write_text(result.stdout + result.stderr, encoding="utf-8")
    if result.returncode:
        raise ProbeBlocked(name + "_failed_offline")
    return result.stdout


def extract_libraries(jar, destination):
    """Extract only existing fat-jar libraries, without trusting zip paths."""
    destination.mkdir()
    found = []
    with zipfile.ZipFile(jar) as archive:
        for item in archive.infolist():
            if not item.filename.startswith("BOOT-INF/lib/") or not item.filename.endswith(".jar"):
                continue
            name = item.filename.removeprefix("BOOT-INF/lib/")
            if "/" in name or "\\" in name or name in {"", ".", ".."} or item.file_size > 128 << 20:
                raise ProbeBlocked("invalid_existing_jar_entry")
            target = destination / name
            if target.exists():
                raise ProbeBlocked("duplicate_existing_library")
            target.write_bytes(archive.read(item))
            found.append(target)
    return found


def prepare_source(folder, output, cache):
    cache = Path(cache).absolute()
    no_links(cache)
    if not cache.is_dir() or not any(cache.rglob("*.jar")):
        raise ProbeBlocked("existing_maven_cache_required")
    for base in [folder / "src", cache]:
        if any(p.is_symlink() for p in base.rglob("*")):
            raise ProbeBlocked("symlink_source_or_cache_forbidden")
    no_links(folder / "pom.xml")
    output.mkdir(parents=True)
    work = output / "module"
    work.mkdir()
    shutil.copyfile(folder / "pom.xml", work / "pom.xml")
    shutil.copytree(folder / "src", work / "src")
    shutil.copytree(cache, output / "m2")
    hashes = {str(p.relative_to(work)): hashlib.sha256(p.read_bytes()).hexdigest()
              for p in work.rglob("*") if p.is_file()}
    (output / "source-manifest.json").write_text(json.dumps(hashes, indent=2) + "\n", encoding="utf-8")
    artifacts = {str(p.relative_to(output / "m2")): hashlib.sha256(p.read_bytes()).hexdigest()
                 for p in (output / "m2").rglob("*.jar")}
    (output / "existing-cache-manifest.json").write_text(json.dumps(artifacts, indent=2) + "\n", encoding="utf-8")
    return work


def run(module, probe, output, cache, *, image=None, festival=FESTIVAL):
    folder, destination = validate(module, probe, output, festival)
    lock = festival / "go/cmd/credit-demo/deploy/sources.lock.json"
    if not lock.is_file() or lock.is_symlink():
        raise ProbeBlocked("original_source_lock_required")
    env = clean_environment()
    if image:
        if sys.platform == "win32" or not re.fullmatch(r"sha256:[0-9a-f]{64}", image):
            raise ProbeBlocked("linux_immutable_build_image_required")
        if any(os.environ.get(k) for k in ["DOCKER_HOST", "DOCKER_CONTEXT"]):
            raise ProbeBlocked("docker_override_forbidden")
    elif not shutil.which("mvn") or not shutil.which("java"):
        raise ProbeBlocked("existing_java_and_maven_required")
    work = prepare_source(folder, destination, cache)
    docker = ["docker", "--context", "default"]
    if image:
        context = json.loads(call(docker + ["context", "inspect", "default"], destination,
                                  "docker-context", env))
        if len(context) != 1 or not context[0]["Endpoints"]["docker"]["Host"].startswith("unix://"):
            raise ProbeBlocked("local_linux_daemon_required")
        # Image configuration may contain hostile secret values: inspect in
        # memory and never write the unvalidated Docker response to evidence.
        details = json.loads(call(docker + ["image", "inspect", image], destination,
                                  "image-inspect", env, retain=False))[0]
        if details.get("Id") != image or details.get("Os") != "linux":
            raise ProbeBlocked("image_identity_mismatch")
        for value in details.get("Config", {}).get("Env", []):
            if value.split("=", 1)[0] not in IMAGE_ENV:
                raise ProbeBlocked("unexpected_build_image_environment")
        (destination / "image.json").write_text(json.dumps({"id": image,
            "environment_keys": sorted(v.split("=", 1)[0] for v in details.get("Config", {}).get("Env", []))}) + "\n", encoding="utf-8")
        prefix = docker + ["run", "--rm", "--network", "none", "--read-only",
                           "--cap-drop", "ALL", "--security-opt", "no-new-privileges",
                           "--user", f"{os.getuid()}:{os.getgid()}",
                           "--tmpfs", "/tmp:rw,nosuid,nodev,size=536870912",
                           "--mount", f"type=bind,src={destination},dst=/evidence",
                           "--workdir", "/evidence/module", "--entrypoint", "mvn", image]
        maven = prefix + ["-o", "-B", "-Dmaven.repo.local=/evidence/m2",
                          "-Duser.home=/tmp", "-DskipTests", "package"]
    else:
        maven = ["mvn", "-o", "-B", f"-Dmaven.repo.local={destination / 'm2'}",
                 "-DskipTests", "package"]
    call(maven, destination, "maven", env, cwd=work)
    jars = [p for p in (work / "target").glob("*.jar") if not p.name.startswith("original-")]
    if len(jars) != 1:
        raise ProbeBlocked("one_original_module_jar_required")
    libs = extract_libraries(jars[0], destination / "libraries")
    cp = [work / "target/test-classes", work / "target/classes", jars[0], *libs]
    if image:
        cp_string = ":".join("/evidence/" + str(p.relative_to(destination)) for p in cp)
        argv = prefix[:prefix.index("--entrypoint")] + ["--entrypoint", "java", image, "-cp", cp_string, probe]
    else:
        argv = ["java", "-cp", os.pathsep.join(map(str, cp)), probe]
    text = call(argv, destination, "probe", env, cwd=work)
    # A zero exit without the explicit plain-main assertions is not acceptance.
    if "H03_PROBE_PASS" not in text:
        raise ProbeBlocked("probe_assertion_receipt_missing")
    result = {"status": "PASS", "module": module, "probe": probe,
              "network": "none" if image else "host_probe_only",
              "runtime_acceptance": "NOT_RUN", "jar_sha256": hashlib.sha256(jars[0].read_bytes()).hexdigest(),
              "fat_jar_libraries": {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in libs}}
    result["original_source_lock_sha256"] = hashlib.sha256(lock.read_bytes()).hexdigest()
    (destination / "result.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--module", choices=PROBES, required=True)
    parser.add_argument("--probe", required=True)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--maven-cache", required=True)
    parser.add_argument("--build-image", help="Existing immutable sha256 image ID; Linux local Docker only")
    args = parser.parse_args()
    try:
        result = run(args.module, args.probe, args.output_dir, args.maven_cache, image=args.build_image)
    except (ProbeBlocked, OSError, ValueError, KeyError, subprocess.SubprocessError, zipfile.BadZipFile):
        print(json.dumps({"status": "BLOCKED", "reason": "offline_probe_prerequisite_or_assertion_failed"}))
        return 2
    print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
