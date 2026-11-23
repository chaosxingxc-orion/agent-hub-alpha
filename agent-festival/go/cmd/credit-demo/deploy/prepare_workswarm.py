"""Fetch an upstream PR-only dependency without changing its locked commit."""
from pathlib import Path
import tomllib

commit = "4c5f3e9eb41bd33d34d012b06a3e364e8497ac11"
ref = "refs/merge-requests/391/head"
lock_path = Path("uv.lock")
lock = lock_path.read_text()
packages = tomllib.loads(lock)["package"]
protocol = next(p for p in packages if p["name"] == "openjiuwen-protocol")
assert protocol["source"]["git"].endswith("#" + commit), "Unexpected protocol revision"
# uv still checks out the immutable #commit. The PR ref lets GitCode advertise
# and fetch that object; it does not select a newer dependency revision.
lock_path.write_text(lock.replace("rev=" + commit, "rev=" + ref.replace("/", "%2F")))
project = Path("pyproject.toml")
project.write_text(project.read_text().replace(commit, ref))
assert next(p for p in tomllib.loads(lock_path.read_text())["package"]
            if p["name"] == "openjiuwen-protocol")["source"]["git"].endswith("#" + commit)
print("Protocol fetch ref adapted; locked commit retained: " + commit)
