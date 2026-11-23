"""Owned, model-free H03 configuration. Help/config never starts a service."""

from __future__ import annotations

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import subprocess

HERE = Path(__file__).resolve().parent
FESTIVAL = HERE.parents[4]
BASE = ("credit", "orchestrator", "engineering-sum", "engineering-reconcile", "bus-db", "rdc", "gateway")
UI = ("swarm", "front")
TEMPLATE_PIN = "ad51806080a381af29f749c969c6e9ec4829f8e4d6a425dc71fc3b1fef2c2fd1"
IMAGE_ID = re.compile(r"sha256:[0-9a-f]{64}\Z")
PROJECT = re.compile(r"h03-[a-z0-9][a-z0-9-]{0,39}\Z")
VARIABLE = re.compile(r"\$\{([A-Z0-9_]+):\?required\}")
IMAGE_ENV = frozenset("PATH LANG LC_ALL HOME JAVA_HOME JAVA_VERSION MAVEN_OPTS GPG_KEY PYTHON_VERSION PYTHON_SHA256 PYTHON_PIP_VERSION PYTHON_SETUPTOOLS_VERSION PYTHON_GET_PIP_URL PYTHON_GET_PIP_SHA256 PG_MAJOR PG_VERSION PG_SHA256 PGDATA GOSU_VERSION NGINX_VERSION NJS_VERSION PKG_RELEASE DYNPKG_RELEASE NJS_RELEASE UV_PYTHON_DOWNLOADS UV_LINK_MODE PYTHONUNBUFFERED UV_PROJECT_ENVIRONMENT FRONTEND_HOST".split())
PRIVATE_KEYS = ("H03_RUNTIME_CREDENTIAL", "H03_DB_PASSWORD", "H03_REGISTRY_TOKEN", "H03_ROUTE_SECRET", "H03_GATEWAY_TOKEN", "H03_LOGIN_PASSWORD")


class Blocked(ValueError):
    """Only finite errors leave the private operator process."""


def unique(pairs):
    out = {}
    for k, v in pairs:
        if k in out:
            raise Blocked("duplicate_json_key")
        out[k] = v
    return out


def read_json(path):
    path = Path(path)
    if path.is_symlink() or not path.is_file() or path.stat().st_size > 1 << 20:
        raise Blocked("regular_bounded_configuration_required")
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=unique,
                      parse_constant=lambda _: (_ for _ in ()).throw(Blocked("invalid_json_number")))


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def selected(ui):
    if ui not in ("none", "original"):
        raise Blocked("original_ui_or_none_required")
    return BASE + (UI if ui == "original" else ())


def environment(source=None):
    source = os.environ if source is None else source
    if any(k.startswith("SWARM_") for k in source):
        raise Blocked("model_environment_forbidden")
    if any(source.get(k) for k in ("DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "COMPOSE_FILE", "COMPOSE_ENV_FILES")):
        raise Blocked("docker_or_compose_override_forbidden")
    return {k: v for k, v in source.items() if k in {"PATH", "SystemRoot", "WINDIR", "TEMP", "TMP"}}


def owned_path(path, *, fresh=False, festival=FESTIVAL):
    path = Path(path).absolute()
    if any(x.is_symlink() for x in [path, *path.parents]) or any(c in str(path) for c in ",\r\n"):
        raise Blocked("unsafe_owned_path")
    base = festival.parent / ".specify/feature/HUB-NEXT/evidence"
    try:
        path.resolve().relative_to(base.resolve())
    except ValueError:
        raise Blocked("owned_evidence_root_required") from None
    if path.resolve() == base.resolve() or fresh and path.exists():
        raise Blocked("fresh_owned_root_required")
    return path


def validate_template(raw):
    if not isinstance(raw, dict) or hashlib.sha256(canonical(raw)).hexdigest() != TEMPLATE_PIN:
        raise Blocked("finite_original_template_required")
    if set(raw) != {"services", "networks", "volumes"} or set(raw["services"]) != set(BASE + UI):
        raise Blocked("exact_service_allowlist_required")
    for service in raw["services"].values():
        if any(k.startswith("SWARM_") for k in service.get("environment", {})):
            raise Blocked("model_configuration_forbidden")
    return raw


def expand(value, env):
    if isinstance(value, dict):
        return {k: expand(v, env) for k, v in value.items()}
    if isinstance(value, list):
        return [expand(v, env) for v in value]
    if isinstance(value, str):
        def replace(match):
            if match[1] not in env:
                raise Blocked("missing_finite_configuration")
            return env[match[1]]
        result = VARIABLE.sub(replace, value)
        if "${" in result:
            raise Blocked("unsupported_interpolation")
        return result
    return value


def env_map(values):
    result = {}
    for item in values:
        key, mark, val = item.partition("=")
        if not mark or key in result:
            raise Blocked("duplicate_or_invalid_environment")
        result[key] = val
    return result


def image_audit(name, item, expected_id):
    if item.get("Id") != expected_id or item.get("Os") != "linux":
        raise Blocked("pinned_linux_image_required")
    config = item.get("Config", {})
    image_env = env_map(config.get("Env", []))
    service_defaults = {"credit": {"CREDIT_DATA": "/data", "CREDIT_ADDR": ":8091"},
                        "bus-db": {"DOCKER_PG_LLVM_DEPS": "llvm19-dev \t\tclang19"}}.get(name, {})
    if set(image_env) - IMAGE_ENV - set(service_defaults) or any(k.startswith("SWARM_") for k in image_env) or any(
            key in image_env and image_env[key] != value for key, value in service_defaults.items()):
        raise Blocked("image_environment_not_allowlisted")
    entry = config.get("Entrypoint") or []
    java_jars = {"orchestrator": "/app/credit-orchestrator.jar", "engineering-sum": "/app/agent.jar",
                 "engineering-reconcile": "/app/agent.jar", "rdc": "/app/rdc.jar", "gateway": "/app/gateway.jar"}
    if name in java_jars:
        if not entry or entry[0] != "java" or entry[-2:] != ["-jar", java_jars[name]] or any(
                not re.fullmatch(r"-Xmx[0-9]+[mg]", x) and x not in {"-Djdk.httpclient.allowRestrictedHeaders=host", "-XX:MaxRAMPercentage=65"}
                for x in entry[1:-2]):
            raise Blocked("original_java_entrypoint_required")
    else:
        allowed = {"credit": [["/opt/aha/run/credit/credit-demo"]],
                   "bus-db": [["docker-entrypoint.sh"], ["/usr/local/bin/docker-entrypoint.sh"]],
                   "swarm": [["/usr/local/bin/aha-entrypoint.sh"]],
                   "front": [["/docker-entrypoint.sh"]]}
        if entry not in allowed[name]:
            raise Blocked("original_entrypoint_required")
    expected_cmd = {"bus-db": ["postgres"], "front": ["nginx", "-g", "daemon off;"]}.get(name)
    if (config.get("Cmd") or None) != expected_cmd:
        raise Blocked("unexpected_image_command")
    return image_env


def resolved_audit(config, expected, project):
    """Compare security-relevant fields after Compose adds its normal defaults."""
    if set(config) - {"name", "services", "networks", "volumes"} or config.get("name") != project:
        raise Blocked("unexpected_resolved_document")
    if set(config.get("services", {})) != set(expected["services"]):
        raise Blocked("unexpected_resolved_services")
    for name, wanted in expected["services"].items():
        actual = config["services"][name]
        defaults = {key for key in ("command", "entrypoint") if key not in wanted and actual.get(key) is None}
        if set(actual) - set(wanted) - defaults:
            raise Blocked("unexpected_resolved_service_fields")
        for key in ["image", "environment", "labels", "restart", "cap_drop", "security_opt", "user"]:
            if (actual.get(key, {}) if key == "environment" else actual.get(key)) != wanted.get(key):
                raise Blocked("resolved_configuration_drift")
        if set(actual.get("depends_on", {})) != set(wanted.get("depends_on", [])):
            raise Blocked("unexpected_dependencies")
        if set(actual.get("networks", {})) != {"h03"}:
            raise Blocked("unexpected_network")
        for dependency in actual.get("depends_on", {}).values():
            if dependency.get("condition") != "service_started" or dependency.get("restart", False):
                raise Blocked("unexpected_dependency_mode")
        mounts = actual.get("volumes", [])
        if len(mounts) != len(wanted.get("volumes", [])):
            raise Blocked("unexpected_mount")
        for mount, target in zip(mounts, wanted.get("volumes", [])):
            if {k: mount.get(k) for k in target} != target or set(mount) - set(target):
                raise Blocked("resolved_mount_drift")
        ports = actual.get("ports", [])
        if len(ports) != len(wanted.get("ports", [])):
            raise Blocked("unexpected_port")
        for port, target in zip(ports, wanted.get("ports", [])):
            if any(str(port.get(k)) != str(v) for k, v in target.items()) or set(port) - set(target) - {"mode"}:
                raise Blocked("non_loopback_or_changed_port")
            if port.get("mode", "ingress") != "ingress":
                raise Blocked("unexpected_port_mode")
    networks = config.get("networks", {})
    if set(networks) != {"h03"} or networks["h03"].get("internal") is not True or networks["h03"].get("external", False):
        raise Blocked("internal_network_required")
    if networks["h03"].get("ipam") != expected["networks"]["h03"].get("ipam"):
        raise Blocked("network_subnet_drift")
    if set(networks["h03"]) - {"name", "internal", "ipam", "labels"} or networks["h03"].get("name") != project + "_h03" or networks["h03"].get("labels") != expected["networks"]["h03"]["labels"]:
        raise Blocked("network_ownership_drift")
    if set(config.get("volumes", {})) != set(expected["volumes"]):
        raise Blocked("unexpected_volume")
    for name, volume in config["volumes"].items():
        if set(volume) - {"name", "labels", "driver"} or volume.get("name") != project + "_" + name or volume.get("driver", "local") != "local" or volume.get("labels") != expected["volumes"][name]["labels"]:
            raise Blocked("unowned_volume")


def docker(argv, env, cwd):
    p = subprocess.run(["docker", "--context", "default", *argv], env=env,
                       cwd=cwd, capture_output=True, text=True, encoding="utf-8", timeout=120)
    if p.returncode:
        raise Blocked("docker_operation_failed")
    return p.stdout


def private_write(path, value):
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, "w", encoding="utf-8") as stream:
        stream.write(json.dumps(value, indent=2) + "\n")


def prepare(project, ui, output, images_path, *, here=HERE, festival=FESTIVAL):
    if not PROJECT.fullmatch(project):
        raise Blocked("finite_project_name_required")
    env = environment()
    names = selected(ui)
    root = owned_path(output, fresh=True, festival=festival)
    raw = validate_template(read_json(here / "compose.h03.yaml"))
    manifest = read_json(images_path)
    if set(manifest) - {"schema_version", "source_lock_sha256", "images", "ui"} or manifest.get("schema_version") != "h03-local-images/v1":
        raise Blocked("finite_image_manifest_required")
    lock = festival / "go/cmd/credit-demo/deploy/sources.lock.json"
    if manifest.get("source_lock_sha256") != digest(lock):
        raise Blocked("source_lock_mismatch")
    images = manifest.get("images", {})
    if set(images) != set(names) or any(not isinstance(v, str) or not IMAGE_ID.fullmatch(v) for v in images.values()):
        raise Blocked("exact_pinned_image_map_required")
    plugin = festival / "go/cmd/credit-demo/workbench-plugin"
    plugin_files = ["extension.py", "extension.yaml", "frontend/dist/app.js", "frontend/dist/index.html"]
    if ui == "original":
        info = manifest.get("ui", {})
        source_lock = read_json(lock)
        pin = hashlib.sha256(canonical({name: digest(plugin / name) for name in plugin_files})).hexdigest()
        if info != {"host_source_revision": source_lock["sources"]["workswarm"]["commit"], "model_free": True, "plugin_sha256": pin}:
            raise Blocked("original_model_free_ui_receipt_required")
    raw["services"] = {k: v for k, v in raw["services"].items() if k in names}
    if ui == "none":
        raw["volumes"].pop("swarm-data")
    root.mkdir(parents=True)
    (root / "config/scenarios").mkdir(parents=True)
    for source, dest in [("engineering-scenario.json", "scenarios/engineering.json"),
                         ("engineering-profile.json", "engineering-profile.json"),
                         ("engineering-material.json", "engineering-material.json"), ("front.conf", "front.conf")]:
        shutil.copyfile(here / source, root / "config" / dest)
    if ui == "original":
        for name in plugin_files:
            target = root / "config/plugin" / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(plugin / name, target)
    nonce = secrets.token_hex(24)
    generated = {key: secrets.token_hex(32) for key in PRIVATE_KEYS}
    generated.update({"H03_OWNER": nonce, "H03_CONFIG": str(root / "config"),
                      "H03_PLUGIN": str(root / "config/plugin"),
                      "H03_FRONT": str(root / "config/front.conf"),
                      "H03_PROFILE_SHA256": digest(root / "config/engineering-profile.json")})
    generated.update({"H03_IMAGE_" + k.upper().replace("-", "_"): v for k, v in images.items()})
    env.update(generated)
    env["COMPOSE_DISABLE_ENV_FILE"] = "true"
    private_write(root / "private-local.json", generated)
    private_write(root / "compose.h03.yaml", raw)
    expected = expand(raw, env)
    context = json.loads(docker(["context", "inspect", "default"], env, root))
    if len(context) != 1 or not context[0]["Endpoints"]["docker"]["Host"].startswith("unix://"):
        raise Blocked("local_linux_docker_required")
    image_envs = {}
    for name in names:
        item = json.loads(docker(["image", "inspect", images[name]], env, root))[0]
        image_envs[name] = image_audit(name, item, images[name])
    args = ["compose", "--project-name", project, "--project-directory", str(root), "-f", str(root / "compose.h03.yaml")]
    config = json.loads(docker(args + ["config", "--format", "json"], env, root), object_pairs_hook=unique)
    resolved_audit(config, expected, project)
    state = {"schema_version": "h03-owned-deployment/v1", "project": project,
             "owner": nonce, "ui": ui, "images": images, "image_envs": image_envs,
             "config_sha256": digest(root / "compose.h03.yaml"),
             "asset_sha256": {str(p.relative_to(root)): digest(p) for p in (root / "config").rglob("*") if p.is_file()}}
    private_write(root / "ownership.json", state)
    # Only keys and image identities enter public-safe configuration evidence.
    private_write(root / "configuration-audit.json", {"services": list(names), "images": images,
        "environment_keys": {k: sorted(v["environment"]) for k, v in expected["services"].items()},
        "internal_network": True, "runtime_acceptance": "NOT_RUN"})
    return root, state, env, expected, args


def container_audit(items, state, expected):
    services = expected["services"]
    seen = set()
    for item in items:
        cfg = item.get("Config", {})
        labels = cfg.get("Labels", {})
        name = labels.get("com.docker.compose.service")
        if name not in services or name in seen or labels.get("com.docker.compose.project") != state["project"] or labels.get("io.agent-hub.h03.owner") != state["owner"]:
            raise Blocked("foreign_or_duplicate_container")
        seen.add(name)
        if item.get("Image") != state["images"][name]:
            raise Blocked("actual_image_drift")
        wanted_env = {**state["image_envs"][name], **services[name]["environment"]}
        if env_map(cfg.get("Env", [])) != wanted_env:
            raise Blocked("actual_environment_drift")
        if cfg.get("User", "") != services[name].get("user", ""):
            raise Blocked("actual_runtime_user_drift")
        host = item.get("HostConfig", {})
        if host.get("Privileged") or host.get("NetworkMode") != state["project"] + "_h03" or set(item.get("NetworkSettings", {}).get("Networks", {})) != {state["project"] + "_h03"}:
            raise Blocked("actual_network_or_privilege_drift")
        if host.get("CapAdd") or set(host.get("CapDrop", [])) != {"ALL"}:
            raise Blocked("actual_capability_drift")
        if host.get("Devices") or host.get("PidMode") == "host" or host.get("IpcMode") == "host" or set(x.removesuffix(":true") for x in host.get("SecurityOpt", [])) != {"no-new-privileges"}:
            raise Blocked("actual_host_access_drift")
        bindings = {str(p["target"]) + "/" + p["protocol"]: [{"HostIp": p["host_ip"], "HostPort": p["published"]}]
                    for p in services[name].get("ports", [])}
        if (host.get("PortBindings") or {}) != bindings:
            raise Blocked("actual_public_port_drift")
        declared = services[name].get("volumes", [])
        actual = item.get("Mounts", [])
        if len(actual) != len(declared):
            raise Blocked("actual_mount_count_drift")
        for wanted in declared:
            matches = [m for m in actual if m.get("Destination") == wanted["target"]]
            if len(matches) != 1:
                raise Blocked("actual_mount_target_drift")
            mount = matches[0]
            if mount.get("Type") != wanted["type"] or mount.get("RW") != (not wanted.get("read_only", False)):
                raise Blocked("actual_mount_mode_drift")
            if wanted["type"] == "bind" and mount.get("Source") != wanted["source"]:
                raise Blocked("actual_bind_drift")
            if wanted["type"] == "volume" and mount.get("Name") != state["project"] + "_" + wanted["source"]:
                raise Blocked("actual_volume_drift")
    if seen != set(services):
        raise Blocked("actual_service_set_incomplete")


def start(prepared):
    root, state, env, expected, args = prepared
    if digest(root / "compose.h03.yaml") != state["config_sha256"] or any(digest(root / name) != pin for name, pin in state["asset_sha256"].items()):
        raise Blocked("source_snapshot_changed")
    for kind in ("container", "volume", "network"):
        command = ["ps", "--all", "--quiet"] if kind == "container" else [kind, "ls", "--quiet"]
        if docker(command + ["--filter", "label=com.docker.compose.project=" + state["project"]], env, root).strip():
            raise Blocked("existing_project_resources_forbidden")
    docker(args + ["up", "--detach", "--no-deps", "--no-build", "--pull", "never", *selected(state["ui"])], env, root)
    ids = docker(["ps", "--all", "--quiet", "--filter", "label=com.docker.compose.project=" + state["project"]], env, root).split()
    if len(ids) != len(selected(state["ui"])) or any(not re.fullmatch(r"[0-9a-f]{12,64}", value) for value in ids):
        raise Blocked("actual_container_set_incomplete")
    items = json.loads(docker(["inspect", *ids], env, root))
    container_audit(items, state, expected)
    network = json.loads(docker(["network", "inspect", state["project"] + "_h03"], env, root))[0]
    if network.get("Internal") is not True or network.get("Labels", {}).get("io.agent-hub.h03.owner") != state["owner"]:
        raise Blocked("actual_network_not_internal")
    return {"status": "STARTED_NATIVE_ACCEPTANCE_NOT_RUN", "services": list(selected(state["ui"])),
            "outbound_observation": "NOT_RUN; required independently by T313"}


def stop(project, output, *, festival=FESTIVAL):
    root = owned_path(output, festival=festival)
    state = read_json(root / "ownership.json")
    private = read_json(root / "private-local.json")
    if not PROJECT.fullmatch(project) or state.get("project") != project or state.get("owner") != private.get("H03_OWNER") or digest(root / "compose.h03.yaml") != state.get("config_sha256"):
        raise Blocked("owned_cleanup_identity_required")
    env = environment()
    env.update(private)
    env["COMPOSE_DISABLE_ENV_FILE"] = "true"
    context = json.loads(docker(["context", "inspect", "default"], env, root))
    if len(context) != 1 or not context[0]["Endpoints"]["docker"]["Host"].startswith("unix://"):
        raise Blocked("local_linux_docker_required")
    ids = docker(["ps", "--all", "--quiet", "--filter", "label=com.docker.compose.project=" + project], env, root).split()
    if ids:
        if any(not re.fullmatch(r"[0-9a-f]{12,64}", value) for value in ids):
            raise Blocked("invalid_owned_container_identity")
        for item in json.loads(docker(["inspect", *ids], env, root)):
            labels = item.get("Config", {}).get("Labels", {})
            service = labels.get("com.docker.compose.service")
            if labels.get("io.agent-hub.h03.owner") != state["owner"] or labels.get("com.docker.compose.project") != project or service not in selected(state["ui"]) or item.get("Image") != state["images"][service]:
                raise Blocked("foreign_cleanup_forbidden")
    network_ids = docker(["network", "ls", "--quiet", "--filter", "label=com.docker.compose.project=" + project], env, root).split()
    for network_id in network_ids:
        if not re.fullmatch(r"[0-9a-f]{12,64}", network_id):
            raise Blocked("invalid_owned_network_identity")
        network = json.loads(docker(["network", "inspect", network_id], env, root))[0]
        if network.get("Name") != project + "_h03" or network.get("Labels", {}).get("io.agent-hub.h03.owner") != state["owner"]:
            raise Blocked("foreign_cleanup_network_forbidden")
    args = ["compose", "--project-name", project, "--project-directory", str(root), "-f", str(root / "compose.h03.yaml")]
    docker(args + ["down", "--timeout", "10"], env, root)
    if docker(["ps", "--all", "--quiet", "--filter", "label=com.docker.compose.project=" + project], env, root).strip():
        raise Blocked("cleanup_unconfirmed")
    return {"status": "OWNED_SERVICES_STOPPED_VOLUMES_RETAINED", "task_and_unknown_history": "preserved"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["config", "start", "stop"])
    parser.add_argument("--project", required=True)
    parser.add_argument("--ui", choices=["none", "original"], default="none")
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--images", help="Required for config/start: private operator manifest of existing immutable image IDs")
    args = parser.parse_args()
    try:
        if args.command == "stop":
            result = stop(args.project, args.output_dir)
        else:
            if args.images is None:
                raise Blocked("operator_image_manifest_required")
            prepared = prepare(args.project, args.ui, args.output_dir, args.images)
            result = start(prepared) if args.command == "start" else {"status": "CONFIGURED_NOT_STARTED", "services": list(selected(args.ui)), "runtime_acceptance": "NOT_RUN"}
    except (Blocked, OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError):
        print(json.dumps({"status": "BLOCKED", "reason": "owned_model_free_deployment_prerequisite_failed"}))
        return 2
    print(json.dumps(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
