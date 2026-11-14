# 逐字取自邢超 GH PR#22 review 5874479780 正文 python 代码块（主线 gh api 取件 2026-11-12；GitCode 桥件截断，此为全本）
"""Independent guard probes. Only temporary synthetic files/repos are used."""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile

spec = importlib.util.spec_from_file_location('gate', Path(__file__).parent / 'agent-festival/tools/claim_gate.py')
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)
with tempfile.TemporaryDirectory(prefix='aha-gate-review-') as td:
    root = Path(td)
    doc = root / 'docs'
    doc.mkdir()
    outside = root / 'outside.txt'
    outside.write_text('one\ntwo\nthree\n', encoding='utf-8')
    (doc / 'targets').write_bytes(os.fsencode(outside) + b'\0')
    command = 'wc -l --files0-from targets'
    argv, reason = gate._guard(command, str(doc))
    result = gate._run_cmd(argv, str(doc), 5) if argv else None
    print(json.dumps({'case': 'indirect-file-escape', 'accepted': argv is not None, 'reason': reason, 'read_outside_synthetic_file': result is not None and str(outside) in result[1], 'status': result[0] if result else None}))

    repo = root / 'repo'
    repo.mkdir()
    marker = root / 'external-diff-ran'
    driver = root / 'diff-driver.sh'
    driver.write_text('#!/bin/sh\nprintf invoked > "' + str(marker) + '"\n', encoding='utf-8')
    driver.chmod(0o700)
    def git(*args):
        return subprocess.run(['git', *args], cwd=repo, check=True, capture_output=True, text=True)
    git('init', '-q')
    git('config', 'user.name', 'Independent Review')
    git('config', 'user.email', 'review@example.invalid')
    git('config', 'commit.gpgsign', 'false')
    (repo / 'data').write_text('a\n', encoding='utf-8')
    git('add', 'data'); git('commit', '-qm', 'first')
    (repo / 'data').write_text('b\n', encoding='utf-8')
    git('add', 'data'); git('commit', '-qm', 'second')
    git('config', 'diff.external', str(driver))
    command = 'git log -1 --oneline --ext-diff -p'
    argv, reason = gate._guard(command, str(repo))
    result = gate._run_cmd(argv, str(repo), 5) if argv else None
    print(json.dumps({'case': 'external-diff-execution', 'accepted': argv is not None, 'reason': reason, 'synthetic_marker_written': marker.exists(), 'status': result[0] if result else None}))


