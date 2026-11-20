#!/usr/bin/env python3
"""Build the pinned W02 prototype without an author's private Maven volume."""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import secrets
import subprocess
import sys
import urllib.request

HERE = Path(__file__).resolve().parent
LOCK = json.loads((HERE / "sources.lock.json").read_text())
CACHE = Path(os.environ.get("DEMO_CACHE", str(Path.home() / ".cache/agent-hub-alpha/issue16"))).resolve()
BUS = False


def bus_credentials():
    """Keep generated service credentials outside source and image layers."""
    path = CACHE / "bus-secrets.json"
    if not path.exists():
        path.parent.mkdir(parents=True, exist_ok=True)
        values = {key: secrets.token_hex(32) for key in (
            "RDC_DB_PASSWORD", "RDC_REGISTRY_API_TOKEN", "RDC_ROUTE_HANDLE_SECRET", "CREDIT_BUS_TOKEN")}
        fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, "w") as stream:
            json.dump(values, stream)
    if path.stat().st_mode & 0o077:
        raise RuntimeError("Bus credentials file must have mode 600")
    return json.loads(path.read_text())


def run(args, **kwargs):
    return subprocess.run([str(x) for x in args], check=True, **kwargs)


def prepare():
    for name, source in LOCK["sources"].items():
        dest = CACHE / "sources" / name
        if not (dest / ".git").exists():
            dest.mkdir(parents=True, exist_ok=True)
            run(["git", "init", "-q", dest])
            run(["git", "-C", dest, "remote", "add", "origin", source["url"]])
        status = run(["git", "-C", dest, "status", "--porcelain"], capture_output=True, text=True).stdout
        if status.strip():
            raise RuntimeError(f"Refusing to change modified source cache: {dest}")
        current = subprocess.run(["git", "-C", str(dest), "rev-parse", "HEAD"], capture_output=True, text=True)
        if current.stdout.strip() != source["commit"]:
            run(["git", "-C", dest, "fetch", "--depth", "1", "origin", source["commit"]])
            run(["git", "-C", dest, "checkout", "--detach", source["commit"]])
        print(f"{name}: {source['commit']}", flush=True)
    tools = CACHE / "tools"
    tools.mkdir(parents=True, exist_ok=True)
    archive = tools / "go.tar.gz"
    if not archive.exists() or hashlib.sha256(archive.read_bytes()).hexdigest() != LOCK["go"]["sha256"]:
        temporary = archive.with_suffix(".partial")
        print("Downloading Go toolchain", flush=True)
        with urllib.request.urlopen(LOCK["go"]["url"], timeout=90) as source, temporary.open("wb") as target:
            shutil.copyfileobj(source, target)
        if hashlib.sha256(temporary.read_bytes()).hexdigest() != LOCK["go"]["sha256"]:
            temporary.unlink()
            raise RuntimeError("Go archive checksum mismatch")
        temporary.replace(archive)
    uv = shutil.which("uv")
    if not uv:
        raise RuntimeError(f"Install uv {LOCK['workswarm_uv']} before building WorkSwarm")
    version = run([uv, "--version"], capture_output=True, text=True).stdout.split()[1]
    if version != LOCK["workswarm_uv"]:
        raise RuntimeError(f"Expected uv {LOCK['workswarm_uv']}; found {version}")
    shutil.copy2(uv, tools / "uv")
    print(f"Source and toolchain cache ready: {CACHE}", flush=True)


def compose(*args, capture=False):
    env = os.environ.copy()
    env["DEMO_CACHE"] = str(CACHE)
    env["COMPOSE_PARALLEL_LIMIT"] = "1"
    if BUS:
        for key, value in bus_credentials().items():
            env.setdefault(key, value)
    secret = Path(env.get("AHA_SECRETS_FILE", str(Path.home() / ".config/aha-swarm/secrets.env")))
    # Prefer explicit SWARM_* configuration or a supplied file. Reuse the
    # existing DeepSeek environment only when neither has been configured.
    model_keys = ("SWARM_API_BASE", "SWARM_API_KEY", "SWARM_MODEL_NAME")
    if not secret.is_file() and not any(env.get(key, "").strip() for key in model_keys):
        if env.get("DEEPSEEK_API_KEY", "").strip():
            env["SWARM_API_KEY"] = env["DEEPSEEK_API_KEY"]
            env["SWARM_API_BASE"] = env.get("COEVOSKILLS_DEEPSEEK_BASE_URL") or env.get("DEEPSEEK_BASE_URL") or "https://api.deepseek.com"
            env["SWARM_MODEL_NAME"] = env.get("COEVOSKILLS_DEEPSEEK_MODEL") or env.get("DEEPSEEK_MODEL", "")
            env.setdefault("SWARM_ENDPOINT_PROFILE", "deepseek")
    dotenv = ["--env-file", secret] if secret.is_file() else []
    files = ["-f", HERE / "compose.yaml"]
    if BUS:
        files += ["-f", HERE / "compose.bus.yaml"]
    return run(["docker", "compose", *dotenv, "--profile", "ui", *files, *args],
               env=env, capture_output=capture, text=True)


def model_preflight():
    # Compose resolves both environment variables and optional dotenv files.
    # Never print the resolved configuration: it contains credentials.
    config = json.loads(compose("config", "--format", "json", capture=True).stdout)
    services = ["swarm", "orchestrator"]
    if BUS:
        services += ["material-agent", "analysis-agent", "writer-agent"]
    for service in services:
        env = config["services"][service]["environment"]
        missing = [key for key in ("SWARM_API_BASE", "SWARM_API_KEY", "SWARM_MODEL_NAME")
                   if not str(env.get(key) or "").strip()]
        if missing:
            raise RuntimeError(f"{service} model configuration keys missing: " + ", ".join(missing))


def wait_bus():
    # The RDC process can be healthy before its first reconciliation cycle.
    compose("exec", "-T", "credit", "python", "-c", '''
import json,os,time,urllib.request
deadline=time.monotonic()+90
while True:
    try:
        for agent in ("credit-material","credit-analysis","credit-writer"):
            req=urllib.request.Request(os.environ["CREDIT_RDC_URL"]+"/v2/agents/"+agent+"/instances",headers={"Authorization":"Bearer "+os.environ["RDC_REGISTRY_API_TOKEN"],"X-Tenant-ID":"issue16"})
            with urllib.request.urlopen(req,timeout=5) as response:
                instances=json.load(response).get("instances",[])
            assert any(i.get("health")=="ONLINE" and i.get("contractVersion")=="1.0" for i in instances)
        print("All three Agent Bus roles discovered and online")
        break
    except Exception:
        if time.monotonic()>deadline:
            raise RuntimeError("Agent Bus roles did not become discoverable within 90 seconds") from None
        time.sleep(2)
''')


def build(which):
    prepare()
    compose("config", "--quiet")
    selected = {"go": ["credit"], "java": ["orchestrator"], "ui": ["swarm"],
                "all": ["credit", "orchestrator", "swarm"]}[which]
    if BUS and which in ("java", "all"):
        selected += ["material-agent", "rdc", "gateway"]
    for service in selected:
        compose("--progress", "plain", "build", service)
    manifest = {"sources": LOCK, "images": {}, "base_images": {}}
    names = {"credit": "agent-hub-issue16/credit:46980e8-local",
             "orchestrator": "agent-hub-issue16/orchestrator:46980e8-local",
             "swarm": "agent-hub-issue16/workswarm:1142e5fb3-local"}
    if BUS:
        names.update({"credit": "agent-hub-issue16/credit:bus-v1", "orchestrator": "agent-hub-issue16/orchestrator:bus-v1",
                      "material-agent": "agent-hub-issue16/professional:bus-v1", "rdc": "agent-hub-issue16/rdc:bf702bc",
                      "gateway": "agent-hub-issue16/gateway:bf702bc"})
    for service in selected:
        data = json.loads(run(["docker", "image", "inspect", names[service]], capture_output=True, text=True).stdout)[0]
        manifest["images"][service] = {"tag": names[service], "id": data["Id"], "created": data["Created"]}
    for name in ["python:3.11-slim", "node:22-slim", "nginx:alpine", "postgres:15-alpine"]:
        result = subprocess.run(["docker", "image", "inspect", name], capture_output=True, text=True)
        if result.returncode == 0:
            data = json.loads(result.stdout)[0]
            manifest["base_images"][name] = {"id": data["Id"], "digests": data.get("RepoDigests", [])}
    evidence = CACHE / "evidence" / "agent-bus" if BUS else CACHE / "evidence"
    evidence.mkdir(parents=True, exist_ok=True)
    (evidence / f"build-{which}.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")


def main():
    global BUS
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bus", action="store_true", help="Use real RDC/Gateway DIRECT and professional runtimes")
    parser.add_argument("command", choices=["prepare", "build", "start", "check", "status", "stop"])
    parser.add_argument("component", nargs="?", default="all", choices=["all", "go", "java", "ui", "core"])
    args = parser.parse_args()
    BUS = args.bus
    if args.command == "prepare":
        prepare()
    elif args.command == "build":
        if args.component == "core":
            build("go")
            build("java")
        else:
            build(args.component)
    elif args.command == "start":
        if args.component == "core":
            services = ["credit", "orchestrator"]
            if BUS:
                model_preflight()
                services += ["bus-db", "rdc", "gateway", "material-agent", "analysis-agent", "writer-agent"]
            compose("up", "-d", "--no-build", "--wait", *services)
        else:
            model_preflight()
            compose("up", "-d", "--no-build", "--wait")
        if BUS:
            wait_bus()
    elif args.command == "check":
        compose("config", "--quiet")
        print("Compose configuration valid. This does not validate model calls or #16 acceptance.")
    elif args.command == "status":
        compose("ps")
    else:
        compose("down")  # Retains named volumes and affects this project only.


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, subprocess.CalledProcessError, OSError) as exc:
        print(f"Demo operation failed: {exc}", file=sys.stderr)
        sys.exit(1)
