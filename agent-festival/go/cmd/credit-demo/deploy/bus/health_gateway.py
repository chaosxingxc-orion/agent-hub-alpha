"""Check the pinned A2A endpoint and tenant authentication without dispatching."""
import json
import os
import urllib.request

payload = {"jsonrpc": "2.0", "id": "health", "method": "GetTask", "params": {"id": "health-nonexistent-task"}}
request = urllib.request.Request("http://127.0.0.1:8080/a2a", data=json.dumps(payload).encode(),
    headers={"Content-Type": "application/json", "Authorization": "Bearer " + os.environ["GATEWAY_TEST_CREDENTIAL_TOKEN"]})
with urllib.request.urlopen(request, timeout=5) as response:
    result = json.load(response)
assert result.get("error", {}).get("data", {}).get("code") == "ROUTE_OWNERSHIP_UNKNOWN", "A2A authentication or protocol unavailable"
