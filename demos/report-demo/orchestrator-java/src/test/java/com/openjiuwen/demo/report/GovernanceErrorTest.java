package com.openjiuwen.demo.report;

import static org.junit.jupiter.api.Assertions.*;
import com.sun.net.httpserver.HttpServer;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.atomic.AtomicInteger;
import org.junit.jupiter.api.Test;

class GovernanceErrorTest {
    @Test void knownIdentityFailuresRemainActionableWithoutCopyingMessages() {
        for (String code : new String[]{"AUTH_MISSING", "AUTH_INVALID", "AUTH_EXPIRED"}) {
            var body = Json.MAPPER.createObjectNode().put("code", code).put("message", "private provider diagnostic");
            assertEquals(code, RemoteExpertExecutor.remoteErrorCode(401, body));
            assertEquals("WAITING_AUTH", BoundNodeExecutor.waitingStatus(code));
        }
    }

    @Test void scopeDenialsBlockWhileDependencyFailuresRemainUnknown() {
        assertEquals("CALLER_NOT_AUTHORIZED", RemoteExpertExecutor.remoteErrorCode(403,
                Json.read("{\"error\":{\"code\":\"CALLER_NOT_AUTHORIZED\"}}")));
        for (String code : new String[]{"CALLER_NOT_AUTHORIZED", "SCOPE_VIOLATION", "AUTH_DISABLED"}) {
            assertEquals("BLOCKED", BoundNodeExecutor.waitingStatus(code));
        }
        for (String code : new String[]{"RDC_UNAVAILABLE", "IAM_UNAVAILABLE", "AUDIT_UNAVAILABLE"}) {
            assertEquals(code, RemoteExpertExecutor.remoteErrorCode(503,
                    Json.MAPPER.createObjectNode().put("code", code)));
            assertEquals("UNKNOWN", BoundNodeExecutor.waitingStatus(code));
        }
    }

    @Test void localDependencyWaitDoesNotPostRemoteWaiting() throws Exception {
        var server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        var resolves = new AtomicInteger();
        var waiting = new AtomicInteger();
        server.createContext("/", exchange -> {
            try (exchange) {
                var path = exchange.getRequestURI().getPath();
                int status = 200;
                byte[] body = "{}".getBytes(StandardCharsets.UTF_8);
                if (path.endsWith("/resolve")) {
                    resolves.incrementAndGet();
                    status = 409;
                    body = "{\"error\":{\"code\":\"WAIT\"}}".getBytes(StandardCharsets.UTF_8);
                } else if (path.endsWith("/waiting")) {
                    waiting.incrementAndGet();
                }
                exchange.sendResponseHeaders(status, body.length);
                exchange.getResponseBody().write(body);
            }
        });
        server.start();
        try {
            var client = new ControlClient("http://127.0.0.1:" + server.getAddress().getPort(), "test-worker-token");
            var executor = new BoundNodeExecutor(client, "task", Json.read("{\"id\":\"select\"}"),
                    "worker", "1", () -> true, null, null);
            assertEquals("WAIT", assertThrows(ControlClient.ControlFailure.class, () -> executor.execute("", "")).code);
            assertEquals(1, resolves.get());
            assertEquals(0, waiting.get());
            assertEquals("UNKNOWN", BoundNodeExecutor.waitingStatus("STORAGE_UNAVAILABLE"));
            assertEquals("FAILED", BoundNodeExecutor.waitingStatus("REMOTE_FAILED"));
        } finally {
            server.stop(0);
        }
    }

    @Test void numericRpcErrorsAndUnknownFreeTextCannotSelectControlStates() {
        assertEquals("RDC_UNAVAILABLE", RemoteExpertExecutor.remoteErrorCode(200,
                Json.read("{\"error\":{\"code\":-32000,\"data\":{\"code\":\"RDC_UNAVAILABLE\"}}}")));
        assertEquals("REMOTE_UNKNOWN", RemoteExpertExecutor.remoteErrorCode(502,
                Json.read("{\"code\":\"COMPLETED\",\"message\":\"do not expose\"}")));
        assertEquals("AUTH_REQUIRED", RemoteExpertExecutor.remoteErrorCode(403, Json.read("{}")));
        assertEquals("REMOTE_UNKNOWN", RemoteExpertExecutor.remoteErrorCode(503, Json.read("{}")));
    }
}
