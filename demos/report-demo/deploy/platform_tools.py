"""Native tool and source paths shared by Windows and Linux launch/verification."""
from __future__ import annotations

import os
from pathlib import Path
import shutil
import sys

REPO = Path(__file__).resolve().parents[3]


def executable(name: str) -> str:
    return name + (".exe" if os.name == "nt" else "")


def java() -> Path:
    home = os.environ.get("JAVA_HOME")
    if home:
        path = Path(home) / "bin" / executable("java")
    else:
        found = shutil.which(executable("java"))
        if not found:
            raise SystemExit("Set JAVA_HOME to a native JDK 21 or put java on PATH")
        path = Path(found)
    if not path.is_file():
        raise SystemExit("Native Java executable missing: " + str(path))
    return path.resolve()


def go() -> Path:
    isolated = REPO / ".local/tools/go/bin" / executable("go")
    found = os.environ.get("GO_BINARY") or (str(isolated) if isolated.is_file() else shutil.which(executable("go")))
    if not found or not Path(found).is_file():
        raise SystemExit("Set GO_BINARY to a native Go executable or put go on PATH")
    return Path(found).resolve()


def maven() -> Path:
    found=shutil.which("mvn.cmd" if os.name == "nt" else "mvn")
    if not found:
        raise SystemExit("Put native Maven on PATH")
    return Path(found).resolve()


def workswarm() -> Path:
    supplied = os.environ.get("WORKSWARM_SOURCE")
    path = Path(supplied) if supplied else REPO / ".local/sources/jiuwenswarm"
    if not path.is_dir():
        raise SystemExit("Set WORKSWARM_SOURCE to the pinned WorkSwarm source directory")
    return path.resolve()


def keycloak() -> Path:
    return Path(os.environ.get("KEYCLOAK_HOME", str(REPO / ".local/tools/keycloak-26.8.2"))).resolve()


def _linux_process_identity(pid: int) -> tuple[str, int]:
    boot_id = Path("/proc/sys/kernel/random/boot_id").read_text(encoding="ascii").strip()
    stat = Path(f"/proc/{pid}/stat").read_text(encoding="ascii", errors="replace")
    # The command name is parenthesized and may itself contain spaces or ')'.
    fields = stat.rsplit(")", 1)[1].split()
    return boot_id, int(fields[19])  # starttime is field 22; fields starts at 3


def record_process(process) -> dict:
    record = {"pid": process.pid, "created": process.create_time()}
    if sys.platform.startswith("linux"):
        record["boot_id"], record["start_ticks"] = _linux_process_identity(process.pid)
    return record


def matches_process(process, record: dict) -> bool:
    if process.pid != record.get("pid"):
        return False
    if "boot_id" in record or "start_ticks" in record:
        if not sys.platform.startswith("linux"):
            return False
        try:
            return _linux_process_identity(process.pid) == (record.get("boot_id"), record.get("start_ticks"))
        except (OSError, ValueError, IndexError):
            return False
    # Compatibility for Windows and PID records written before this change.
    return abs(process.create_time() - record.get("created", 0)) < .01
