package com.openjiuwen.demo.professional;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.openjiuwen.service.spec.dto.QueryResponse;
import com.openjiuwen.service.spec.dto.ServeRequest;
import com.openjiuwen.service.spec.spi.AgentHandler;
import com.openjiuwen.service.spec.spi.QueryStreamObserver;
import java.math.BigDecimal;
import java.net.URI;
import java.net.http.*;
import java.time.Duration;
import java.util.*;

/** Model/tool-free arithmetic behind the original Runtime's AgentHandler. */
final class H03EngineeringHandler implements AgentHandler {
    @FunctionalInterface interface ContextReader { JsonNode read(Map<String, Object> auth); }
    private final String role, service, version, base;
    private final ContextReader reader;

    H03EngineeringHandler(Map<String, String> env) { this(env, null); }
    H03EngineeringHandler(Map<String, String> env, ContextReader reader) {
        role = env.getOrDefault("H03_ENGINEERING_ROLE", "");
        service = env.getOrDefault("AGENT_SERVICE_ID", "");
        version = env.getOrDefault("OPENJIUWEN_SERVICE_VERSION", "");
        base = env.getOrDefault("CREDIT_URL", "").replaceAll("/$", "");
        if (!Set.of("sum", "reconcile").contains(role) || !service.equals("h03-engineering-" + role) || !version.equals("0.1.0")
                || env.containsKey("H03_RUNTIME_CREDENTIAL") || env.keySet().stream().anyMatch(k -> k.startsWith("SWARM_")))
            throw new IllegalArgumentException("Independent model-free engineering identity required");
        URI origin = URI.create(base);
        if (!Set.of("http", "https").contains(origin.getScheme()) || origin.getHost() == null || origin.getUserInfo() != null
                || origin.getQuery() != null || origin.getFragment() != null || !origin.getPath().isEmpty())
            throw new IllegalArgumentException("Control origin required");
        this.reader = reader == null ? this::context : reader;
    }
    @Override public void start() { }
    @Override public void stop() { }
    @Override public QueryResponse query(ServeRequest request) {
        return new QueryResponse(Map.of("role", "assistant", "content", perform(request.lastUserQuery(), request.getConversationId())), request.getConversationId());
    }
    @Override public void streamQuery(ServeRequest request, QueryStreamObserver observer) {
        observer.onError(new UnsupportedOperationException("Engineering capability is non-streaming"));
    }
    private JsonNode context(Map<String, Object> auth) {
        try {
            HttpRequest req = HttpRequest.newBuilder(URI.create(base + "/agent/context")).timeout(Duration.ofSeconds(15))
                    .header("Content-Type", "application/json").POST(HttpRequest.BodyPublishers.ofString(H03BoundNodeCodec.JSON.writeValueAsString(auth))).build();
            HttpResponse<String> response = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(5)).build().send(req, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() != 200) throw new IllegalStateException("Scoped context rejected");
            return H03BoundNodeCodec.parse(response.body());
        } catch (InterruptedException e) { Thread.currentThread().interrupt(); throw new IllegalStateException("Scoped context interrupted"); }
        catch (Exception e) { throw new IllegalStateException("Scoped context unavailable"); }
    }
    String perform(String raw, String conversation) {
        if (conversation == null || conversation.isBlank() || conversation.length() > 256) throw new IllegalArgumentException("Outer context required");
        JsonNode envelope = H03BoundNodeCodec.parse(raw);
        H03BoundNodeCodec.keys(envelope, "run_id", "step", "token");
        String run = H03BoundNodeCodec.text(envelope, "run_id"), step = H03BoundNodeCodec.text(envelope, "step");
        String capability = H03BoundNodeCodec.text(envelope, "token");
        if (run.length() > 256 || step.length() > 256 || capability.length() > 512 || capability.chars().anyMatch(c -> c < 33 || c > 126))
            throw new IllegalArgumentException("Scoped node envelope required");
        Map<String, Object> auth = new LinkedHashMap<>();
        auth.put("run_id", run); auth.put("step", step); auth.put("token", capability);
        auth.put("service_id", service); auth.put("service_version", version); auth.put("context_id", conversation);
        JsonNode safe = reader.read(auth);
        H03BoundNodeCodec.keys(safe, "expert", "tools", "node_contract", "call");
        JsonNode call = safe.path("call"), contract = safe.path("node_contract");
        H03BoundNodeCodec.keys(call, "node_execution_id", "graph_revision", "node_path", "attempt", "binding", "input_digest", "input", "goal", "request_id", "context_id");
        if (!safe.path("tools").isArray() || !safe.path("tools").isEmpty()
                || !H03BoundNodeCodec.text(safe, "expert").equals("eng." + role)
                || !H03BoundNodeCodec.text(call, "context_id").equals(conversation)
                || !H03BoundNodeCodec.text(call.path("binding"), "agent_id").equals(service)
                || !H03BoundNodeCodec.text(call.path("binding"), "version").equals(version))
            throw new IllegalArgumentException("Scoped role/context differs");
        JsonNode input = H03BoundNodeCodec.input(call, contract, run, step);
        String output = calculate(input, call, contract, role);
        H03BoundNodeCodec.output(output, call, input, contract);
        return H03BoundNodeCodec.canonical(Map.of("node_execution_id", H03BoundNodeCodec.text(call, "node_execution_id"),
                "input_digest", H03BoundNodeCodec.text(call, "input_digest"), "attempt", 1, "output", output));
    }
    static String calculate(JsonNode input, JsonNode call, JsonNode contract, String role) {
        H03BoundNodeCodec.contract(contract, "h03-engineering-" + role);
        Map<String, JsonNode> fields = H03BoundNodeCodec.specs(input.path("fields"), true);
        BigDecimal value;
        if (role.equals("sum")) value = H03BoundNodeCodec.decimal(H03BoundNodeCodec.text(fields.get("component_a"), "value"))
                .add(H03BoundNodeCodec.decimal(H03BoundNodeCodec.text(fields.get("component_b"), "value")));
        else value = H03BoundNodeCodec.decimal(H03BoundNodeCodec.text(fields.get("total"), "value"))
                .subtract(H03BoundNodeCodec.decimal(H03BoundNodeCodec.text(fields.get("reference"), "value")));
        String computed = value.stripTrailingZeros().toPlainString(); H03BoundNodeCodec.decimal(computed);
        ArrayNode result = H03BoundNodeCodec.JSON.createArrayNode();
        Map<String, JsonNode> output = H03BoundNodeCodec.specs(contract.path("output_fields"), false);
        for (String name : output.keySet()) {
            ObjectNode f = output.get(name).deepCopy();
            f.put("value", name.equals("reference") ? H03BoundNodeCodec.text(fields.get("reference"), "value") : computed); result.add(f);
        }
        return H03BoundNodeCodec.canonical(Map.of("schema_version", "hub-node-output/v1",
                "execution_binding_sha256", H03BoundNodeCodec.text(input, "execution_binding_sha256"),
                "node_contract_sha256", H03BoundNodeCodec.text(input, "node_contract_sha256"),
                "input_digest", H03BoundNodeCodec.text(call, "input_digest"), "fields", result));
    }
}
