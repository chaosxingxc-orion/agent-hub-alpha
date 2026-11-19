"""Own local reference lab; no action without an explicit subcommand.

Run under Linux/WSL with a local Docker daemon. Private state is operator input,
not an authentication receipt or evidence of registration/production eligibility.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import socket
from string import Template
import subprocess
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

HERE = Path(__file__).resolve().parent
DEPLOY = HERE.parent
REVISION = "bf702bc95064d3491cf8b374f54d271363d84230"
IMAGE = "agent-hub-h02b/{target}:" + REVISION
SAFE_ID = re.compile(r"[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}\Z")
HEX64 = re.compile(r"[a-f0-9]{64}\Z")
POSTGRES_DIGEST = (
    "sha256:09e4f20b14ddb3dfe3a0c825b206032aaf8f28300ba2070c0b60fc1c10c6abc7"
)
POSTGRES_ORIGINAL = "postgres:15-alpine@" + POSTGRES_DIGEST
POSTGRES_MIRROR = "docker.m.daocloud.io/library/postgres:15-alpine@" + POSTGRES_DIGEST
PUBLICATION_KEYS = set(
    "schema_version logical_id native_agent_id artifact_sha256 capability_version deployment_version instance_id contract_version protocol_version card_digest contract_digest tenant source_revision manifest_sha256 wheel_sha256 acceptance_sha256 closure_sha256 original_handoff address port endpoint_path agent_card governance execution_authorized".split()
)


def publication_bindings(descriptors, retained=None):
    """Reject malformed/conflicting explicit v2 bindings before any external action."""
    if not isinstance(descriptors, list) or not 1 <= len(descriptors) <= 2:
        raise LabError("one_or_two_descriptors_required")
    retained = [] if retained is None else retained
    if not isinstance(retained, list) or len(retained) > 32:
        raise LabError("retained_descriptors_invalid")
    capabilities, instances_seen = {}, {}
    fields = "native_agent_id artifact_sha256 capability_version deployment_version contract_version protocol_version address port endpoint_path card_digest contract_digest".split()
    for item in descriptors + retained:
        if (
            not isinstance(item, dict)
            or set(item) != PUBLICATION_KEYS
            or item.get("schema_version") != "publication-reference/v2"
        ):
            raise LabError("explicit_v2_descriptor_required")
        for key, length in (
            ("capability_version", 16),
            ("deployment_version", 32),
            ("instance_id", 64),
        ):
            value = item.get(key)
            if (
                not isinstance(value, str)
                or len(value) > length
                or not SAFE_ID.fullmatch(value)
            ):
                raise LabError("explicit_version_or_instance_invalid")
        if any(
            not isinstance(item[k], str) for k in fields if k != "port"
        ) or not isinstance(item["tenant"], str):
            raise LabError("descriptor_binding_type_invalid")
        port_number(item["port"])
        if not isinstance(item.get("artifact_sha256"), str) or not HEX64.fullmatch(
            item["artifact_sha256"]
        ):
            raise LabError("invalid_artifact_sha256")
        capkey = tuple(
            item[k] for k in ("tenant", "native_agent_id", "capability_version")
        )
        instancekey = (item["tenant"], item["instance_id"])
        identity = tuple(item[k] for k in fields)
        if (
            capkey in capabilities and capabilities[capkey] != item["artifact_sha256"]
        ) or (
            instancekey in instances_seen and instances_seen[instancekey] != identity
        ):
            raise LabError("publication_binding_conflict")
        capabilities[capkey] = item["artifact_sha256"]
        instances_seen[instancekey] = identity


class LabError(ValueError):
    """Finite operator error; subprocess output is not disclosed."""


def run(argv, *, timeout=60):
    """Never invoke a shell, print private output, or inherit remote Docker overrides."""
    result = subprocess.run(argv, capture_output=True, text=True, timeout=timeout)
    if result.returncode:
        raise LabError("command_failed:" + Path(argv[0]).name)
    return result.stdout.strip()


def local_docker():
    if os.environ.get("DOCKER_HOST") or os.environ.get("DOCKER_CONTEXT"):
        raise LabError("docker_override_forbidden")
    data = json.loads(run(["docker", "context", "inspect"]))
    endpoint = data[0]["Endpoints"]["docker"]["Host"]
    if not endpoint.startswith("unix://"):
        raise LabError("local_linux_docker_required")
    return {"name": data[0].get("Name"), "endpoint": endpoint}


def local_buildx():
    """Pin the daemon-backed default builder, independently of selected Buildx.

    Checking the Docker context alone is insufficient: the selected Buildx
    builder may use a remote node. No builder is created or selected globally.
    """
    if os.environ.get("BUILDX_BUILDER") or os.environ.get("BUILDKIT_HOST"):
        raise LabError("buildx_override_forbidden")
    context = local_docker()
    if context["name"] != "default":
        raise LabError("local_default_context_required_for_build")
    inspect = ["docker", "--context", "default", "buildx", "inspect", "default"]
    try:
        builder = json.loads(run(inspect + ["--format", "{{json .}}"]))
    except LabError:
        # Buildx releases without inspect --format still expose the same local
        # facts. Parse anchored fields, never infer endpoint locality from name.
        text = run(inspect)
        parts = re.split(r"(?m)^Nodes:\s*$", text)
        if len(parts) != 2:
            raise LabError("buildx_inspect_shape_invalid")
        names = re.findall(r"(?m)^Name:\s+(\S+)\s*$", parts[0])
        drivers = re.findall(r"(?m)^Driver:\s+(\S+)\s*$", parts[0])
        endpoints = re.findall(r"(?m)^Endpoint:\s+(\S+)\s*$", parts[1])
        if len(names) != 1 or len(drivers) != 1:
            raise LabError("buildx_inspect_shape_invalid")
        builder = {
            "Name": names[0],
            "Driver": drivers[0],
            "Nodes": [{"Endpoint": e} for e in endpoints],
        }
    nodes = builder.get("Nodes", [])
    if (
        builder.get("Name") != "default"
        or builder.get("Driver") != "docker"
        or len(nodes) != 1
        or nodes[0].get("Endpoint") not in {"default", context["endpoint"]}
    ):
        raise LabError("local_daemon_buildx_required")
    return ["docker", "--context", "default", "buildx", "build", "--builder", "default"]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_git(source):
    """Use real Windows Git for an NTFS checkout when WSL interop is available."""
    executable = Path("/mnt/c/Program Files/Git/cmd/git.exe")
    if os.name != "nt" and str(source).startswith("/mnt/") and executable.is_file():
        winpath = run(["wslpath", "-w", str(source)])
        return [str(executable), "-c", "core.fileMode=false", "-C", winpath]
    return ["git", "-c", "core.fileMode=false", "-C", str(source)]


def verify_source(source, *, candidate=False):
    source = Path(source).resolve(strict=True)
    if run(source_git(source) + ["rev-parse", "HEAD"]) != REVISION:
        raise LabError("source_revision_mismatch")
    if not candidate and run(
        [
            "git",
            "-c",
            "core.fileMode=false",
            "-C",
            str(source),
            "status",
            "--porcelain",
            "--untracked-files=all",
        ],
        timeout=300,
    ):
        raise LabError("source_not_clean")
    pins = json.loads((DEPLOY / "sources.lock.json").read_text())
    if pins["sources"]["agent-solution"]["commit"] != REVISION:
        raise LabError("source_lock_mismatch")
    if candidate:
        source_closure(source)
    return source


def port_number(value):
    if type(value) is not int or not 1024 <= value <= 65535:
        raise LabError("invalid_port")
    return value


def ports_free(ports):
    if len(set(ports)) != len(ports):
        raise LabError("duplicate_port")
    for port in ports:
        port_number(port)
        with socket.socket() as sock:
            try:
                sock.bind(("127.0.0.1", port))
            except OSError:
                raise LabError("port_occupied") from None


def instances(descriptors, owner):
    publication_bindings(descriptors)
    if not isinstance(descriptors, list) or not 1 <= len(descriptors) <= 2:
        raise LabError("one_or_two_descriptors_required")
    result, ids, tenants = [], set(), set()
    for item in descriptors:
        for key in ("logical_id", "tenant"):
            if not SAFE_ID.fullmatch(item.get(key, "")):
                raise LabError("invalid_descriptor_identity")
        native = item.get("native_agent_id", "")
        if not 1 <= len(native) <= 64 or any(not 33 <= ord(c) <= 126 for c in native):
            raise LabError("invalid_native_id")
        if native in ids:
            raise LabError("duplicate_native_id")
        ids.add(native)
        tenants.add(item["tenant"])
        if not HEX64.fullmatch(item.get("artifact_sha256", "")):
            raise LabError("invalid_artifact_sha256")
        if (
            item.get("contract_version") != "1.0.0"
            or item.get("protocol_version") != "1.0"
        ):
            raise LabError("unsupported_contract_or_protocol")
        if (
            item.get("address") != "127.0.0.1"
            or item.get("endpoint_path") != "/a2a/jsonrpc/"
        ):
            raise LabError("owned_loopback_endpoint_required")
        card = item.get("agent_card", {})
        if card.get("name") != native:
            raise LabError("card_name_mismatch")
        port = port_number(item.get("port"))
        # Original AgentIdCodec derives agentId from serviceId, not card.name.
        # The stable logical alias belongs in Hub's private mapping only.
        result.append(
            {
                "tenant-id": item["tenant"],
                "service-id": native,
                "instance-id": item["instance_id"],
                "base-url": f"http://127.0.0.1:{port}",
                "route-key": item["endpoint_path"],
                "contract-version": "1.0.0",
                "capability-version": item["capability_version"],
                "deployment-version": item["deployment_version"],
                "readiness": "READY",
            }
        )
    if len(tenants) != 1:
        raise LabError("single_tenant_required")
    return result, tenants.pop()


def reconciliation_compatibility(descriptors):
    """Check original validator semantics without amending an original AgentCard.

    AgentCardValidator requires card.version and returns it as capabilityVersion.
    ReconciliationService prefers those fetched versions to binding defaults.
    An operator declaration cannot override either fact.
    """
    reasons = set()
    for item in descriptors:
        card = item.get("agent_card", {})
        if not isinstance(card.get("version"), str) or not card.get("version"):
            reasons.add("ORIGINAL_CARD_VERSION_REQUIRED")
        elif card["version"] != item.get("artifact_sha256"):
            reasons.add("ORIGINAL_CARD_CAPABILITY_VERSION_MISMATCH")
        interfaces = card.get("supportedInterfaces", [])
        versions = {
            entry.get("protocolVersion", entry.get("version"))
            for entry in interfaces
            if isinstance(entry, dict)
        }
        if versions != {item.get("protocol_version")}:
            reasons.add("ORIGINAL_CARD_PROTOCOL_VERSION_MISMATCH")
        # This reference accepts one explicitly pinned original JSON-RPC
        # interface only. It must not choose a second interface just to make
        # the original fetcher's business/protocol conflation appear valid.
        if item.get("contract_version") != item.get("protocol_version"):
            reasons.add("ORIGINAL_FETCHED_CONTRACT_VERSION_MISMATCH")
    return {
        "status": "BLOCKED" if reasons else "SOURCE_PRECHECK_ONLY",
        "reasons": sorted(reasons),
        "registration_verified": False,
    }


def private_write(path, text):
    with path.open("x", encoding="utf-8") as stream:
        stream.write(text)
    path.chmod(0o600)


def postgres_transport():
    """Verify a previously pulled, exact pinned database image; never pull or start."""
    local_docker()
    info = json.loads(
        run(["docker", "--context", "default", "image", "inspect", POSTGRES_MIRROR])
    )[0]
    if (
        info.get("Os") != "linux"
        or info.get("Architecture") != "amd64"
        or "docker.m.daocloud.io/library/postgres@" + POSTGRES_DIGEST
        not in info.get("RepoDigests", [])
        or not re.fullmatch(r"sha256:[a-f0-9]{64}", info.get("Id", ""))
    ):
        raise LabError("pinned_postgres_transport_invalid")
    return {
        "original_ref": POSTGRES_ORIGINAL,
        "mirror_ref": POSTGRES_MIRROR,
        "digest": POSTGRES_DIGEST,
        "platform": "linux/amd64",
        "id": info["Id"],
    }


def prepare(
    state_dir,
    source_dir,
    descriptors,
    *,
    db_port=15432,
    rdc_port=18092,
    gateway_port=18080,
    candidate=False,
    build_receipt=None,
    test_report=None,
    mirror_pinned_postgres=False,
    retained_descriptors=None,
):
    """Prepare only. The descriptor is a deployment declaration, not verified evidence."""
    publication_bindings(descriptors, retained_descriptors)
    owner = "builder-h02b-" + secrets.token_hex(8)
    rows, tenant = instances(descriptors, owner)
    source = (
        verify_source(source_dir, candidate=True)
        if candidate
        else verify_source(source_dir)
    )
    qualified = (
        candidate_receipts(source, build_receipt, test_report) if candidate else None
    )
    database_transport = postgres_transport() if mirror_pinned_postgres else None
    ports_free([db_port, rdc_port, gateway_port])
    if candidate:
        for row in rows:
            row["version-source"] = "deployment-binding"
    state = Path(state_dir).absolute()
    if state.exists() or any(p.is_symlink() for p in [state, *state.parents]):
        raise LabError("new_unlinked_state_required")
    if any(c in str(state) for c in "\r\n$#'\" "):
        raise LabError("simple_private_state_path_required")
    state.mkdir(parents=True, mode=0o700)
    env = {
        "H02B_OWNER": owner,
        "H02B_STATE": str(state),
        "H02B_TENANT": tenant,
        "H02B_DB_PORT": str(db_port),
        "H02B_RDC_PORT": str(rdc_port),
        "H02B_GATEWAY_PORT": str(gateway_port),
        "H02B_POSTGRES_IMAGE": POSTGRES_ORIGINAL,
    }
    if database_transport:
        env["H02B_POSTGRES_IMAGE"] = database_transport["mirror_ref"]
    env.update(
        H02B_RDC_IMAGE=qualified["images"]["rdc"]["tag"]
        if qualified
        else IMAGE.format(target="rdc"),
        H02B_GATEWAY_IMAGE=qualified["images"]["gateway"]["tag"]
        if qualified
        else IMAGE.format(target="gateway"),
        H02B_PROTOCOL_VERSION="1.0" if candidate else "",
    )
    for name in (
        "RDC_DB_PASSWORD",
        "RDC_ROUTE_HANDLE_SECRET",
        "RDC_REGISTRY_API_TOKEN",
        "H02B_GATEWAY_TOKEN",
    ):
        env[name] = secrets.token_hex(32)
    private_write(
        state / "private.env", "".join(f"{key}={value}\n" for key, value in env.items())
    )
    template = Template((HERE / "rdc.yml.template").read_text())
    private_write(
        state / "rdc.yml",
        template.substitute(
            rdc_port=rdc_port,
            db_port=db_port,
            instances=json.dumps(rows, ensure_ascii=True),
        ),
    )
    private_write(state / "compose.yaml", (HERE / "compose.yaml").read_text())
    receipt = {
        "schema": "h02b-lab/v1",
        "project": owner,
        "source": str(source),
        "revision": REVISION,
        "ports": [db_port, rdc_port, gateway_port],
        "tenant": tenant,
        "reconciliation_compatibility": candidate_card_check(descriptors)
        if candidate
        else reconciliation_compatibility(descriptors),
        "candidate": candidate,
        "source_closure": qualified["source_closure"] if qualified else None,
        "candidate_build": qualified,
        "postgres_transport": database_transport,
        "files": {
            name: sha(state / name)
            for name in ["private.env", "rdc.yml", "compose.yaml"]
        },
        "tooling": {
            name: sha(HERE / name)
            for name in ["launch.py", "compose.yaml", "rdc.yml.template"]
        },
    }
    private_write(state / "ownership.json", json.dumps(receipt, indent=2))
    return {
        "status": "PREPARED_NOT_STARTED",
        "project": owner,
        "registration": "NOT_RUN",
        "reconciliation_compatibility": receipt["reconciliation_compatibility"],
    }


def load_state(state_dir):
    state = Path(state_dir).resolve(strict=True)
    if any(
        p.is_symlink()
        for p in [Path(state_dir).absolute(), *Path(state_dir).absolute().parents]
    ):
        raise LabError("linked_state_forbidden")
    data = json.loads((state / "ownership.json").read_text())
    if data.get("schema") != "h02b-lab/v1" or not re.fullmatch(
        r"builder-h02b-[a-f0-9]{16}", data.get("project", "")
    ):
        raise LabError("ownership_receipt_invalid")
    for name in ["private.env", "rdc.yml", "compose.yaml"]:
        if (state / name).is_symlink() or sha(state / name) != data.get(
            "files", {}
        ).get(name):
            raise LabError("state_file_changed")
    for name in ["launch.py", "compose.yaml", "rdc.yml.template"]:
        if sha(HERE / name) != data.get("tooling", {}).get(name):
            raise LabError("launcher_source_changed")
    return state, data


def compose_args(state, data):
    # Compose gives inherited shell variables precedence over --env-file.
    # Refuse that ambiguity rather than let ambient values replace pinned images.
    keys = [
        line.split("=", 1)[0]
        for line in (state / "private.env").read_text().splitlines()
    ]
    if any(key in os.environ for key in keys):
        raise LabError("compose_environment_override_forbidden")
    return [
        "docker",
        "compose",
        "--project-name",
        data["project"],
        "--env-file",
        str(state / "private.env"),
        "-f",
        str(state / "compose.yaml"),
    ]


def verify_ownership(data):
    ids = run(
        [
            "docker",
            "ps",
            "-aq",
            "--filter",
            "label=com.docker.compose.project=" + data["project"],
        ]
    ).split()
    for ident in ids:
        obj = json.loads(run(["docker", "inspect", ident]))[0]
        labels = obj.get("Config", {}).get("Labels", {})
        if labels.get("io.agent-hub.h02b.owner") != data["project"]:
            raise LabError("foreign_resource_in_project")
    return ids


def original_base_transport():
    """Only relocate exact original amd64 manifest bytes, never change base version."""
    api = "https://hub.docker.com/v2/repositories/library/python/tags/3.11-slim"
    with urlopen(api, timeout=20) as response:
        value = json.load(response)
    images = [
        entry
        for entry in value.get("images", [])
        if entry.get("architecture") == "amd64" and entry.get("os") == "linux"
    ]
    if len(images) != 1 or not re.fullmatch(
        r"sha256:[a-f0-9]{64}", images[0].get("digest", "")
    ):
        raise LabError("original_base_metadata_invalid")
    digest = images[0]["digest"]
    return {
        "official_metadata": api,
        "original_tag": "python:3.11-slim",
        "platform": "linux/amd64",
        "manifest_digest": digest,
        "mirror_ref": "docker.m.daocloud.io/library/python@" + digest,
    }


def existing_proxy_keys():
    """Validate opt-in proxy transport without exposing configuration values."""
    from urllib.parse import urlsplit

    keys, loopback = [], False
    for key in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
        value = os.environ.get(key)
        if not value:
            continue
        try:
            parsed = urlsplit(value)
            valid = (
                parsed.scheme in ("http", "https")
                and parsed.hostname
                and parsed.port is not None
                and parsed.username is None
                and parsed.password is None
                and parsed.path in ("", "/")
                and not parsed.query
                and not parsed.fragment
                and not any(char.isspace() for char in value)
            )
        except ValueError:
            valid = False
        if not valid:
            raise LabError("existing_proxy_transport_invalid")
        keys.append(key)
        loopback |= parsed.hostname in ("localhost", "127.0.0.1", "::1")
    return keys, loopback


def build(
    source_dir,
    log_dir,
    *,
    candidate=False,
    mirror_original_base=False,
    use_existing_proxy=False,
):
    """Build original images only; never starts services or sends candidate requests."""
    source = (
        verify_source(source_dir, candidate=True)
        if candidate
        else verify_source(source_dir)
    )
    closure = source_closure(source) if candidate else None
    build_argv = local_buildx()
    logs = Path(log_dir).absolute()
    logs.mkdir(parents=True, exist_ok=True)
    context = DEPLOY.parents[3]
    proxy_args = []
    if use_existing_proxy:
        keys, loopback = existing_proxy_keys()
        for key in keys:
            proxy_args += ["--build-arg", key]
        if not proxy_args:
            raise LabError("existing_build_proxy_required")
        if loopback:
            proxy_args += ["--network", "host"]
    base_transport = original_base_transport() if mirror_original_base else None
    tags = {
        t: candidate_tag(t, closure["sha256"]) if candidate else IMAGE.format(target=t)
        for t in ("build", "rdc", "gateway")
    }
    for target in ("build", "rdc", "gateway") if candidate else ("rdc", "gateway"):
        argv = build_argv + [
            "--load",
            "--target",
            target,
            "--tag",
            tags[target],
            "--build-context",
            "solution-src=" + str(source),
            "-f",
            str(DEPLOY / "Dockerfile.bus"),
            str(context),
        ]
        argv += proxy_args
        if base_transport:
            argv += ["--build-arg", "PYTHON_IMAGE=" + base_transport["mirror_ref"]]
        with (logs / (target + "-build.log")).open("x", encoding="utf-8") as output:
            result = subprocess.run(
                argv, stdout=output, stderr=subprocess.STDOUT, timeout=1800
            )
        if result.returncode:
            raise LabError("original_image_build_failed:" + target)
    receipt = {
        "revision": REVISION,
        "source_clean": not candidate,
        "base_image_transport": base_transport,
        "existing_proxy_forwarded": use_existing_proxy,
        "source_closure": closure,
        "candidate": candidate,
        "dockerfile_sha256": sha(DEPLOY / "Dockerfile.bus"),
        "images": {
            t: run(
                [
                    "docker",
                    "--context",
                    "default",
                    "image",
                    "inspect",
                    IMAGE.format(target=t),
                    "--format",
                    "{{.Id}}",
                ]
            )
            for t in (() if candidate else ("rdc", "gateway"))
        },
    }
    if candidate:
        if source_closure(source) != closure:
            raise LabError("source_changed_during_build")
        receipt["images"] = {
            t: {
                "tag": tags[t],
                "id": run(
                    [
                        "docker",
                        "--context",
                        "default",
                        "image",
                        "inspect",
                        tags[t],
                        "--format",
                        "{{.Id}}",
                    ]
                ),
            }
            for t in tags
        }
    private_write(logs / "image-build.json", json.dumps(receipt, indent=2))
    return {"status": "IMAGES_BUILT_NOT_STARTED", "source_revision": REVISION}


def action(command, state_dir):
    state, data = load_state(state_dir)
    local_docker()
    ids = verify_ownership(data)
    argv = compose_args(state, data)
    if command == "up":
        if (
            data.get("reconciliation_compatibility", {}).get("status")
            != "SOURCE_PRECHECK_ONLY"
        ):
            raise LabError("original_reconciliation_incompatible")
        if ids:
            raise LabError("existing_project_query_before_restart")
        if data.get("candidate"):
            if source_closure(Path(data["source"])) != data["source_closure"]:
                raise LabError("candidate_source_changed")
            verify_images(data["candidate_build"])
        else:
            verify_source(data["source"])
        if data.get("postgres_transport") is not None:
            if postgres_transport() != data["postgres_transport"]:
                raise LabError("pinned_postgres_transport_changed")
        ports_free(data["ports"])
        run(argv + ["config", "--quiet"])
        run(argv + ["up", "--detach", "--no-build"], timeout=180)
        # Started is not registered. Actual registry facts are read separately.
        return {"status": "STARTED_REGISTRATION_UNVERIFIED", "project": data["project"]}
    if command == "down":
        run(argv + ["down", "--timeout", "20"], timeout=60)
        return {
            "status": "OWNED_CONTAINERS_STOPPED",
            "volumes": "RETAINED",
            "project": data["project"],
        }
    env = dict(
        line.split("=", 1) for line in (state / "private.env").read_text().splitlines()
    )
    try:
        request = Request(
            f"http://127.0.0.1:{data['ports'][1]}/v2/agents",
            headers={
                "Authorization": "Bearer " + env["RDC_REGISTRY_API_TOKEN"],
                "X-Tenant-ID": data["tenant"],
            },
        )
        with urlopen(request, timeout=5) as response:
            ready = response.status == 200 and len(ids) == 3
        with socket.create_connection(("127.0.0.1", data["ports"][2]), timeout=3):
            pass
    except (HTTPError, URLError, OSError):
        ready = False
    return {
        "status": "SERVICES_REACHABLE" if ready else "NOT_READY",
        "owned_containers": len(ids),
        "registration": "CHECK_ORIGINAL_CARD_AND_INSTANCES",
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("prepare")
    p.add_argument("--state-dir", required=True)
    p.add_argument("--source-dir", required=True)
    p.add_argument("--descriptors", required=True)
    p.add_argument("--retained-descriptors")
    p.add_argument("--candidate", action="store_true")
    p.add_argument("--build-receipt")
    p.add_argument("--test-report")
    p.add_argument("--mirror-pinned-postgres", action="store_true")
    for name, default in [("db", 15432), ("rdc", 18092), ("gateway", 18080)]:
        p.add_argument("--" + name + "-port", type=int, default=default)
    p = sub.add_parser("build")
    p.add_argument("--state-dir", required=True)
    p.add_argument("--source-dir", required=True)
    p.add_argument("--candidate", action="store_true")
    p.add_argument("--mirror-original-base", action="store_true")
    p.add_argument("--use-existing-proxy", action="store_true")
    p = sub.add_parser("test")
    p.add_argument("--source-dir", required=True)
    p.add_argument("--state-dir", required=True)
    p.add_argument("--build-receipt", required=True)
    for command in ["up", "status", "down"]:
        p = sub.add_parser(command)
        p.add_argument("--state-dir", required=True)
    args = parser.parse_args()
    try:
        if args.command == "prepare":
            result = prepare(
                args.state_dir,
                args.source_dir,
                json.loads(Path(args.descriptors).read_text()),
                db_port=args.db_port,
                rdc_port=args.rdc_port,
                gateway_port=args.gateway_port,
                candidate=args.candidate,
                build_receipt=args.build_receipt,
                test_report=args.test_report,
                mirror_pinned_postgres=args.mirror_pinned_postgres,
                retained_descriptors=json.loads(
                    Path(args.retained_descriptors).read_text()
                )
                if args.retained_descriptors
                else None,
            )
        elif args.command == "build":
            result = build(
                args.source_dir,
                args.state_dir,
                candidate=args.candidate,
                mirror_original_base=args.mirror_original_base,
                use_existing_proxy=args.use_existing_proxy,
            )
        elif args.command == "test":
            result = test_candidate(args.source_dir, args.state_dir, args.build_receipt)
        else:
            result = action(args.command, args.state_dir)
    except LabError as error:
        # LabError messages are finite codes constructed by this module, never
        # raw subprocess output or configuration text containing credentials.
        print(json.dumps({"status": "FAILED_OR_UNKNOWN", "code": str(error)}))
        return 1
    except (
        OSError,
        KeyError,
        TypeError,
        json.JSONDecodeError,
        subprocess.TimeoutExpired,
    ):
        print(
            json.dumps(
                {
                    "status": "FAILED_OR_UNKNOWN",
                    "detail": "Inspect private local evidence; do not redispatch business commands.",
                }
            )
        )
        return 1
    print(json.dumps(result))
    return 1 if result.get("status") == "FAIL" else 0


_RDC = "common/agent-bus/registry-discovery-center/"
_GW = "common/agent-bus/agent-gateway/"
ALLOWED_SOURCE = (
    {
        _RDC + "src/main/java/com/openjiuwen/rdc/" + n
        for n in [
            "card/CardVersionSource.java",
            "deployment/DeploymentDiscoveryProperties.java",
            "reconcile/ReconciliationScheduler.java",
            "reconcile/ReconciliationService.java",
            "card/AgentCardValidator.java",
            "card/AgentCardFetcher.java",
            "card/RouteTargetDeriver.java",
        ]
    }
    | {
        _RDC + "src/test/java/com/openjiuwen/rdc/" + n
        for n in [
            "deployment/DeploymentVersionSourceConfigurationTest.java",
            "card/AgentCardValidatorTest.java",
            "card/AgentCardFetcherTest.java",
            "reconcile/ReconciliationDeploymentVersionsIntegrationTest.java",
            "card/NativeA2aRouteTargetDeriverTest.java",
        ]
    }
    | {
        _GW + "src/main/java/com/openjiuwen/gateway/direct/HttpAgentRuntimeClient.java",
        _GW
        + "src/test/java/com/openjiuwen/gateway/direct/HttpAgentRuntimeClientTest.java",
        _GW + "README.md",
        _RDC + "README.md",
        "common/agent-bus/docs/native-a2a-reference.md",
    }
)


def source_closure(source):
    """Hash actual allowed changed/new source; reject every undeclared product path."""
    source = Path(source).resolve(strict=True)
    prefix = source_git(source)
    if run(prefix + ["rev-parse", "HEAD"]) != REVISION:
        raise LabError("candidate_base_revision_mismatch")
    paths = set(
        run(prefix + ["diff", "--name-only", REVISION, "--"], timeout=300).splitlines()
    )
    paths.update(
        run(
            prefix + ["ls-files", "--others", "--exclude-standard"], timeout=300
        ).splitlines()
    )
    paths = {p for p in paths if not p.startswith(".codegraph/")}
    if not paths or not paths <= ALLOWED_SOURCE:
        raise LabError("candidate_source_scope_invalid")
    files = {}
    for relative in sorted(paths):
        path = source / relative
        if (
            not path.is_file()
            or path.is_symlink()
            or any(p.is_symlink() for p in path.parents)
        ):
            raise LabError("candidate_source_file_invalid")
        files[relative] = sha(path)
    raw = json.dumps(files, sort_keys=True, separators=(",", ":")).encode()
    return {
        "base_revision": REVISION,
        "files": files,
        "sha256": hashlib.sha256(raw).hexdigest(),
    }


def candidate_tag(target, digest):
    return "agent-hub-h02b/" + target + ":candidate-" + digest[:24]


def verify_images(receipt):
    local_docker()
    for info in receipt["images"].values():
        if (
            run(
                [
                    "docker",
                    "--context",
                    "default",
                    "image",
                    "inspect",
                    info["tag"],
                    "--format",
                    "{{.Id}}",
                ]
            )
            != info["id"]
        ):
            raise LabError("candidate_image_changed")


def candidate_receipts(source, build_path, tests_path):
    if not build_path or not tests_path:
        raise LabError("candidate_build_and_tests_required")
    build = json.loads(Path(build_path).read_text())
    tests = json.loads(Path(tests_path).read_text())
    closure = source_closure(source)
    if (
        build.get("candidate") is not True
        or build.get("source_closure") != closure
        or tests.get("source_closure") != closure
        or tests.get("status") != "PASS"
        or tests.get("build_receipt_sha256") != sha(Path(build_path))
    ):
        raise LabError("candidate_receipt_binding_invalid")
    if build.get("dockerfile_sha256") != sha(DEPLOY / "Dockerfile.bus"):
        raise LabError("candidate_dockerfile_changed")
    report_root = Path(tests_path).resolve().parent
    if not tests.get("raw_reports") or tests.get("critical_tests_complete") is not True:
        raise LabError("candidate_test_evidence_missing")
    for relative, expected in tests["raw_reports"].items():
        path = report_root / relative
        if (
            Path(relative).is_absolute()
            or ".." in Path(relative).parts
            or path.is_symlink()
            or not path.is_file()
            or sha(path) != expected
        ):
            raise LabError("candidate_test_evidence_changed")
    actual = read_suites([report_root / name for name in tests["raw_reports"]])
    if (
        actual != tests.get("suites")
        or tests.get("returncode") != 0
        or not critical_complete(actual)
    ):
        raise LabError("candidate_test_results_invalid")
    verify_images(build)
    return build


def candidate_card_check(descriptors):
    for item in descriptors:
        interfaces = item["agent_card"].get("supportedInterfaces", [])
        expected = f"http://127.0.0.1:{item['port']}/a2a/jsonrpc/"
        if (
            len(interfaces) != 1
            or interfaces[0].get("protocolVersion") != "1.0"
            or interfaces[0].get("protocolBinding") != "JSONRPC"
            or interfaces[0].get("url") != expected
        ):
            raise LabError("candidate_interface_invalid")
    return {
        "status": "SOURCE_PRECHECK_ONLY",
        "reasons": [],
        "registration_verified": False,
    }


def read_suites(paths):
    import xml.etree.ElementTree as ET

    suites = {}
    for path in paths:
        suite = ET.parse(path).getroot()
        if suite.tag != "testsuite" or suite.attrib.get("name") in suites:
            raise LabError("candidate_test_xml_invalid")
        suites[suite.attrib["name"]] = {
            key: int(suite.attrib.get(key, "0"))
            for key in ["tests", "failures", "errors", "skipped"]
        }
    return suites


def critical_complete(suites):
    required = [
        "DeploymentVersionSourceConfigurationTest",
        "ReconciliationDeploymentVersionsIntegrationTest",
        "NativeA2aRouteTargetDeriverTest",
        "HttpAgentRuntimeClientTest",
        "AgentCardValidatorTest",
        "AgentCardFetcherTest",
        "AgentCardSignatureVerifierTest",
    ]
    return all(
        item["errors"] == item["failures"] == 0 for item in suites.values()
    ) and all(
        any(
            name.endswith(required_name)
            and item["tests"] > 0
            and item["skipped"] == item["errors"] == item["failures"] == 0
            for name, item in suites.items()
        )
        for required_name in required
    )


def test_candidate(source_dir, log_dir, build_path):
    """Run original Maven under nonroot nobody; retain raw Surefire XML for review."""
    source = verify_source(source_dir, candidate=True)
    receipt = json.loads(Path(build_path).read_text())
    closure = source_closure(source)
    if receipt.get("candidate") is not True or receipt.get("source_closure") != closure:
        raise LabError("candidate_build_binding_invalid")
    verify_images(receipt)
    logs = Path(log_dir).absolute()
    if logs.exists():
        raise LabError("new_test_output_required")
    logs.mkdir(parents=True)
    # /tmp/run is entirely inside this ephemeral container; host source is never mounted.
    script = """set -eu
mkdir -p /tmp/check/home /tmp/check/m2
cp -R /src/solution /tmp/check/src
cp /tmp/maven_settings.py /tmp/check/settings.py
cat > /tmp/h02b-copy-reports.py <<'H02B_REPORTS'
import pathlib, shutil
out = pathlib.Path("/evidence/surefire")
out.mkdir(exist_ok=True)
for path in pathlib.Path("/tmp/check/src/common/agent-bus").glob("*/target/surefire-reports/TEST-*.xml"):
    shutil.copyfile(path, out / (path.parent.parent.parent.name + "-" + path.name))
H02B_REPORTS
trap 'python /tmp/h02b-copy-reports.py' EXIT
python /tmp/check/settings.py
mvn -B -s /tmp/maven-settings.xml -Dmaven.repo.local=/tmp/check/m2 -f /tmp/check/src/pom.xml -N install
mvn -B -s /tmp/maven-settings.xml -Dmaven.repo.local=/tmp/check/m2 -f /tmp/check/src/common/agent-bus/registry-discovery-center/pom.xml install
mvn -B -s /tmp/maven-settings.xml -Dmaven.repo.local=/tmp/check/m2 -f /tmp/check/src/common/agent-bus/event-bus/pom.xml -pl event-bus-spi -am -Dmaven.test.skip=true install
mvn -B -s /tmp/maven-settings.xml -Dmaven.repo.local=/tmp/check/m2 -f /tmp/check/src/common/agent-bus/agent-gateway/pom.xml test
python -c 'import pathlib,shutil; out=pathlib.Path("/evidence/surefire"); out.mkdir(); [(shutil.copyfile(p,out/(p.parent.parent.parent.name+"-"+p.name))) for p in pathlib.Path("/tmp/check/src/common/agent-bus").glob("*/target/surefire-reports/TEST-*.xml")]'
"""
    # Container uid nobody cannot modify host data except the dedicated new evidence directory.
    logs.chmod(0o777)
    name = "builder-h02b-test-" + secrets.token_hex(8)
    argv = [
        "docker",
        "--context",
        "default",
        "run",
        "--rm",
        "--name",
        name,
        "--label",
        "io.agent-hub.h02b.owner=" + name,
        "--user",
        "nobody",
        "--mount",
        "type=bind,src=" + str(logs) + ",dst=/evidence",
        "--env",
        "MAVEN_OPTS=-Xmx768m -Duser.home=/tmp/check/home",
        receipt["images"]["build"]["tag"],
        "sh",
        "-c",
        script,
    ]
    if receipt.get("existing_proxy_forwarded"):
        keys, loopback = existing_proxy_keys()
        if not keys:
            raise LabError("existing_test_proxy_required")
        options = [part for key in keys for part in ("--env", key)]
        if loopback:
            options += ["--network", "host"]
        image_index = argv.index(receipt["images"]["build"]["tag"])
        argv[image_index:image_index] = options
    with (logs / "maven.log").open("x") as stream:
        result = subprocess.run(
            argv, stdout=stream, stderr=subprocess.STDOUT, timeout=3600
        )
    reports = read_suites(list((logs / "surefire").glob("*.xml")))
    complete = critical_complete(reports)
    good = result.returncode == 0 and complete and source_closure(source) == closure
    report = {
        "status": "PASS" if good else "FAIL",
        "source_closure": closure,
        "build_receipt_sha256": sha(Path(build_path)),
        "returncode": result.returncode,
        "suites": reports,
        "critical_tests_complete": complete,
        "nonroot_user": "nobody",
        "raw_reports": {str(q.relative_to(logs)): sha(q) for q in logs.rglob("*.xml")},
    }
    private_write(logs / "test-report.json", json.dumps(report, indent=2))
    return report


if __name__ == "__main__":
    raise SystemExit(main())
