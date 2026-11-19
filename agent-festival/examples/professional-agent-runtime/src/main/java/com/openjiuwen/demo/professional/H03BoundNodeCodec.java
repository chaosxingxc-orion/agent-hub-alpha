package com.openjiuwen.demo.professional;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.math.BigDecimal;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

/** Private finite codec. Original Go Call digests remain authoritative. */
final class H03BoundNodeCodec {
    static final String PROFILE = "two-role-identity/v1";
    static final ObjectMapper JSON = new ObjectMapper().enable(JsonParser.Feature.STRICT_DUPLICATE_DETECTION)
            .enable(DeserializationFeature.FAIL_ON_TRAILING_TOKENS);
    private H03BoundNodeCodec() { }

    static JsonNode parse(String raw) {
        if (raw == null || raw.isEmpty() || raw.getBytes(StandardCharsets.UTF_8).length > 65536)
            throw new IllegalArgumentException("Bounded node JSON required");
        try { JsonNode n = JSON.readTree(raw); finite(n, 0); return n; }
        catch (Exception e) { throw new IllegalArgumentException("Invalid finite node JSON"); }
    }
    private static void finite(JsonNode n, int depth) {
        if (n == null || depth > 16 || n.size() > 64 || n.isFloatingPointNumber() || n.isIntegralNumber() && !n.canConvertToInt())
            throw new IllegalArgumentException("Finite node value required");
        if (n.isTextual()) {
            String s = n.textValue();
            for (int i = 0; i < s.length(); i++) {
                char c = s.charAt(i);
                if (Character.isHighSurrogate(c)) {
                    if (++i >= s.length() || !Character.isLowSurrogate(s.charAt(i))) throw new IllegalArgumentException("Invalid Unicode");
                } else if (Character.isLowSurrogate(c)) throw new IllegalArgumentException("Invalid Unicode");
            }
        }
        if (n.isContainerNode()) n.forEach(v -> finite(v, depth + 1));
    }
    static void keys(JsonNode n, String... expected) {
        if (n == null || !n.isObject()) throw new IllegalArgumentException("Node object required");
        Set<String> actual = new HashSet<>(); n.fieldNames().forEachRemaining(actual::add);
        if (!actual.equals(Set.of(expected))) throw new IllegalArgumentException("Node field set differs");
    }
    static String text(JsonNode n, String key) {
        JsonNode v = n.get(key);
        if (v == null || !v.isTextual() || v.textValue().isBlank()) throw new IllegalArgumentException("Node text required");
        return v.textValue();
    }
    static String canonical(JsonNode n) {
        finite(n, 0);
        try {
            if (n.isObject()) {
                List<String> names = new ArrayList<>(); n.fieldNames().forEachRemaining(names::add); Collections.sort(names);
                List<String> parts = new ArrayList<>();
                for (String name : names) parts.add(JSON.writeValueAsString(name) + ":" + canonical(n.get(name)));
                return "{" + String.join(",", parts) + "}";
            }
            if (n.isArray()) { List<String> parts = new ArrayList<>(); n.forEach(v -> parts.add(canonical(v))); return "[" + String.join(",", parts) + "]"; }
            return JSON.writeValueAsString(n);
        } catch (Exception e) { throw new IllegalArgumentException("Node canonicalization failed"); }
    }
    static String canonical(Object value) { return canonical(JSON.valueToTree(value)); }
    static String sha(String... parts) {
        try {
            MessageDigest d = MessageDigest.getInstance("SHA-256");
            for (String s : parts) { d.update(s.getBytes(StandardCharsets.UTF_8)); d.update((byte) 0); }
            return "sha256:" + HexFormat.of().formatHex(d.digest());
        } catch (Exception e) { throw new IllegalStateException(e); }
    }
    static String rawHash(JsonNode value) {
        try { return "sha256:" + HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(canonical(value).getBytes(StandardCharsets.UTF_8))); }
        catch (Exception e) { throw new IllegalArgumentException("Node hash unavailable"); }
    }
    static void hash(String s, boolean prefixed) {
        if (!s.matches((prefixed ? "sha256:" : "") + "[0-9a-f]{64}")) throw new IllegalArgumentException("Node digest required");
    }
    static BigDecimal decimal(String s) {
        if (!s.matches("-?(0|[1-9][0-9]{0,9})(\\.[0-9]{1,6})?") || !new BigDecimal(s).stripTrailingZeros().toPlainString().equals(s))
            throw new IllegalArgumentException("Canonical bounded decimal required");
        return new BigDecimal(s);
    }
    static Map<String, JsonNode> specs(JsonNode array, boolean values) {
        if (!array.isArray() || array.isEmpty() || array.size() > 8) throw new IllegalArgumentException("Finite fields required");
        Map<String, JsonNode> result = new TreeMap<>(); String last = "";
        for (JsonNode f : array) {
            if (values) keys(f, "name", "type", "unit", "period", "basis", "value");
            else keys(f, "name", "type", "unit", "period", "basis");
            String name = text(f, "name");
            if (!Set.of("component_a", "component_b", "reference", "total", "difference").contains(name)
                    || !text(f, "type").equals("decimal") || !text(f, "unit").equals("USD-million")
                    || !text(f, "period").equals("FY2025") || !text(f, "basis").equals("consolidated")
                    || result.put(name, f) != null || values && name.compareTo(last) <= 0)
                throw new IllegalArgumentException("Engineering field semantics differ");
            last = name;
            if (values) decimal(text(f, "value"));
        }
        return result;
    }
    static void fields(JsonNode values, JsonNode expected) {
        Map<String, JsonNode> v = specs(values, true), e = specs(expected, false);
        if (!v.keySet().equals(e.keySet())) throw new IllegalArgumentException("Required fields differ");
        for (String name : e.keySet()) {
            ObjectNode f = v.get(name).deepCopy(); f.remove("value");
            if (!f.equals(e.get(name))) throw new IllegalArgumentException("Field contract differs");
        }
    }
    static void contract(JsonNode c, String service) {
        keys(c, "schema_version", "input_fields", "output_fields");
        if (!text(c, "schema_version").equals("hub-node-contract/v1")) throw new IllegalArgumentException("Unknown node contract");
        boolean sum = service.equals("h03-engineering-sum");
        if (!sum && !service.equals("h03-engineering-reconcile")) throw new IllegalArgumentException("Unknown engineering service");
        if (!specs(c.path("input_fields"), false).keySet().equals(sum ? Set.of("component_a", "component_b", "reference") : Set.of("reference", "total"))
                || !specs(c.path("output_fields"), false).keySet().equals(sum ? Set.of("reference", "total") : Set.of("difference")))
            throw new IllegalArgumentException("Role contract differs");
    }
    static JsonNode input(JsonNode call, JsonNode contract, String run, String step) {
        String raw = text(call, "input"); JsonNode n = parse(raw);
        keys(call.path("binding"), "agent_id", "version", "card_digest", "contract_digest");
        hash(text(call.path("binding"), "card_digest"), true); hash(text(call.path("binding"), "contract_digest"), true);
        keys(n, "schema_version", "execution_binding_sha256", "node_contract_sha256", "plan_hash", "run_id", "node_execution_id", "node_path", "attempt", "fields", "dependencies");
        contract(contract, text(call.path("binding"), "agent_id"));
        if (!canonical(n).equals(raw) || !text(n, "schema_version").equals("hub-bound-input/v1") || !text(n, "run_id").equals(run)
                || !text(n, "node_path").equals(step) || !text(call, "node_path").equals(step)
                || !n.path("attempt").isInt() || n.path("attempt").asInt() != 1 || !call.path("attempt").isInt() || call.path("attempt").asInt() != 1
                || !text(n, "node_execution_id").equals(text(call, "node_execution_id"))
                || !sha(run, text(call, "graph_revision"), step, "iteration:0").equals(text(n, "node_execution_id"))
                || !rawHash(contract).equals(text(n, "node_contract_sha256"))) throw new IllegalArgumentException("Authoritative node input identity differs");
        hash(text(n, "execution_binding_sha256"), true); hash(text(n, "plan_hash"), true);
        hash(text(call, "input_digest"), true); text(call, "context_id"); text(call, "request_id");
        if (!text(call.path("binding"), "version").equals("0.1.0")) throw new IllegalArgumentException("Engineering version differs");
        fields(n.path("fields"), contract.path("input_fields"));
        JsonNode deps = n.path("dependencies");
        if (!deps.isArray() || deps.size() != 1) throw new IllegalArgumentException("Complete dependency required");
        JsonNode d = deps.get(0);
        if (text(call.path("binding"), "agent_id").equals("h03-engineering-sum")) {
            keys(d, "kind", "id", "sha256");
            if (!text(d, "kind").equals("material") || !text(d, "id").equals("engineering-material-a")) throw new IllegalArgumentException("Material identity differs");
            hash(text(d, "sha256"), true);
        } else {
            keys(d, "kind", "id", "node_execution_id", "artifact_digest");
            if (!text(d, "kind").equals("step") || text(d, "id").equals(step)) throw new IllegalArgumentException("Upstream identity differs");
            hash(text(d, "node_execution_id"), true); hash(text(d, "artifact_digest"), true);
        }
        return n;
    }
    static String output(String raw, JsonNode call, JsonNode input, JsonNode contract) {
        JsonNode n = parse(raw);
        keys(n, "schema_version", "execution_binding_sha256", "node_contract_sha256", "input_digest", "fields");
        if (!canonical(n).equals(raw) || !text(n, "schema_version").equals("hub-node-output/v1")
                || !text(n, "execution_binding_sha256").equals(text(input, "execution_binding_sha256"))
                || !text(n, "node_contract_sha256").equals(text(input, "node_contract_sha256"))
                || !text(n, "input_digest").equals(text(call, "input_digest"))) throw new IllegalArgumentException("Output binding differs");
        fields(n.path("fields"), contract.path("output_fields"));
        return raw;
    }
}
