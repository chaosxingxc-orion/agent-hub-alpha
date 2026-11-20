#!/usr/bin/env python3
"""Test-only transport faults in front of the real Gateway; never mocks Agents."""
import http.server
import json
import os
from pathlib import Path
import threading
import time
import urllib.error
import urllib.request

LOCK = threading.Lock()
MODE = Path("/tmp/bus-fault-mode")
EVENTS = Path("/tmp/bus-fault-events.jsonl")


class Proxy(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_POST(self):
        data = self.rfile.read(int(self.headers.get("Content-Length", "0")))
        body = json.loads(data)
        method = body.get("method")
        with LOCK:
            mode = MODE.read_text().strip() if MODE.exists() else "pass"
            if method == "GetTask" and mode not in ("pass", "hold-completed-once"):
                MODE.write_text("pass")
            elif method != "GetTask":
                mode = "pass"
        status = 200
        if mode == "drop-query-once":
            status = 503
            response = json.dumps({"code": "TEST_CONNECTION_LOST", "message": "Injected lost query response"}).encode()
        else:
            headers = {k: v for k, v in self.headers.items() if k.lower() not in ("host", "connection", "content-length")}
            req = urllib.request.Request(os.environ.get("UPSTREAM", "http://gateway:8080") + self.path, data=data, headers=headers)
            try:
                with urllib.request.urlopen(req, timeout=60) as upstream:
                    status, response = upstream.status, upstream.read()
            except urllib.error.HTTPError as exc:
                status, response = exc.code, exc.read()
            except OSError:
                status, response = 502, b'{"code":"TEST_UPSTREAM_DISCONNECTED"}'
        if mode == "hold-completed-once":
            result = json.loads(response).get("result", {})
            task = result.get("task", result)
            if task.get("status", {}).get("state") == "TASK_STATE_COMPLETED":
                with LOCK:
                    MODE.write_text("pass")
                    Path("/tmp/bus-completed-ready").write_text("ready")
                deadline = time.monotonic() + 55
                while not Path("/tmp/bus-release").exists() and time.monotonic() < deadline:
                    time.sleep(0.1)
            else:
                mode = "pass"
        event = {"at": time.time(), "method": method, "id": body.get("id"),
                 "task_id": body.get("params", {}).get("id"), "status": status, "fault": mode}
        with LOCK, EVENTS.open("a") as stream:
            stream.write(json.dumps(event) + "\n")
        if mode == "late-query-once":
            time.sleep(35)  # Exceeds the caller's 30-second deadline; not a model delay.
        try:
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(response)))
            self.end_headers()
            self.wfile.write(response)
        except (BrokenPipeError, ConnectionResetError):
            pass


http.server.ThreadingHTTPServer(("0.0.0.0", 8099), Proxy).serve_forever()
