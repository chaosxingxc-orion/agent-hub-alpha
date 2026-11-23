#!/usr/bin/env python3
"""Check installed WorkSwarm assets and protocol imports with networking disabled."""
import json
import time
from demo import CACHE, run

PROBE = r'''
import importlib.metadata as metadata
import json
from pathlib import Path
import shutil
import jiuwenswarm
from openjiuwen_protocol.message import ReqMethod
from jiuwenswarm.channels.web.app_web import _default_dist_dir

dist = _default_dist_dir()
assert (dist / "index.html").is_file(), dist
assert any((dist / "assets").glob("*.js")), dist
for command in ("jiuwenswarm-init", "jiuwenswarm-start"):
    assert shutil.which(command), command
methods = ("MCP_APP_LIST_TOOLS", "MCP_APP_READ_RESOURCE", "MCP_APP_CALL_TOOL")
for method in methods:
    assert getattr(ReqMethod, method).value.startswith("mcp_app."), method
assert Path(jiuwenswarm.__file__).is_relative_to("/opt/venv"), jiuwenswarm.__file__
print(json.dumps({"workswarm_version": metadata.version("workswarm"),
    "frontend_assets": "passed", "installed_commands": "passed", "protocol_methods": list(methods),
    "network": "disabled", "probe_model_calls": 0, "scope": "image-packaging-only"}))
'''

if __name__ == "__main__":
    result = run(["docker", "run", "--rm", "--network", "none", "--entrypoint", "python",
                  "agent-hub-issue16/workswarm:1142e5fb3-local", "-c", PROBE],
                 capture_output=True, text=True)
    evidence = json.loads(result.stdout)
    evidence["checked_at_unix"] = int(time.time())
    target = CACHE / "evidence" / "workswarm-image-smoke.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(evidence, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps(evidence, ensure_ascii=False, indent=2))
    print(f"Evidence: {target}")
