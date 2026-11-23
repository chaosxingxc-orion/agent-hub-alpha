"""Isolated adapter tests; SDK stand-in does not prove WorkSwarm compatibility."""
import importlib.util
from pathlib import Path
import sys
import types
import unittest
from unittest.mock import AsyncMock, patch

sdk = types.ModuleType("jiuwenswarm.extensions.sdk")
sdk.ApplicationPluginExtension = object
sdk.FrontendContribution = object
spec = importlib.util.spec_from_file_location("workbench_extension", Path(__file__).with_name("extension.py"))
extension = importlib.util.module_from_spec(spec)
with patch.dict(sys.modules, {"jiuwenswarm.extensions.sdk": sdk}):
    spec.loader.exec_module(extension)


class ResumeAdapterTests(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.methods = {}
        self.channel = types.SimpleNamespace(
            register_method=lambda name, fn, **kw: self.methods.update({name: (fn, kw)}),
            send_response=AsyncMock(),
        )
        self.plugin = extension.CreditWorkbench()
        self.plugin.bind_web_channel(self.channel, None)

    async def invoke(self, params):
        fn, options = self.methods["plugin.credit_workbench.resume"]
        self.assertTrue(options["local_only"])
        await fn(object(), "req-1", params, None)
        return self.channel.send_response.call_args.kwargs

    async def test_forwards_only_original_run_and_demo_credentials(self):
        with patch.object(extension, "_call", return_value={"run": "R1", "status": "queued"}) as call:
            result = await self.invoke({"run": "R1", "password": "fixture", "target": "other-agent", "user": "admin"})
        call.assert_called_once_with(self.plugin._base, "POST", "/api/resume", {"run": "R1", "user": "alice", "password": "fixture"})
        self.assertTrue(result["ok"])
        self.assertEqual(result["payload"]["status"], "queued")

    async def test_invalid_requests_never_reach_service(self):
        with patch.object(extension, "_call") as call:
            for params in (None, {}, {"run": "R1", "password": 1}, {"run": "R1/../R2", "password": "x"}, {"run": "R1\n", "password": "x"}):
                result = await self.invoke(params)
                self.assertFalse(result["ok"])
                self.assertEqual(result["code"], "BAD_REQUEST")
            call.assert_not_called()

    async def test_failure_is_not_success_and_not_retried(self):
        for code in ("REJECTED", "UPSTREAM"):
            with patch.object(extension, "_call", side_effect=extension.OrchestrationError("failed", code)) as call:
                result = await self.invoke({"run": "R1", "password": "fixture"})
            call.assert_called_once()
            self.assertFalse(result["ok"])
            self.assertEqual(result["code"], code)


if __name__ == "__main__":
    unittest.main()
