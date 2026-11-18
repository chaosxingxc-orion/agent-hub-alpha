/*
 * AgentChatClient — openjiuwen agent chat 端点的 Java 客户端。
 *
 * 框架合规：不硬接 LLM API，所有 LLM 逻辑在 openjiuwen DeepAgent 内。
 * 本类只负责 HTTP 调用 skill-app-agent 的 /v1/chat (SSE) + 解析响应。
 *
 * 降级策略：连接失败/超时 → 返回 null，调用方回退模板正文。
 */

package com.agenthub.w06.checker.server;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;

final class AgentChatClient {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private static final String AGENT_URL = System.getenv()
            .getOrDefault("AGENT_CHAT_URL", "http://skill-app-agent:8310/v1/chat");

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(10))
            .build();

    /**
     * 调用 skill-app-agent chat 端点（独立一次性会话）。
     */
    static String chat(String message) {
        return chat(message, "w06-pipeline-" + System.currentTimeMillis());
    }

    /**
     * 调用 skill-app-agent chat 端点，返回 agent 生成的文本。
     * 框架记忆支撑：openjiuwen 按 sessionId 维护会话上下文，同一生成流程内的
     * 任务书分派 / 缺节续写 / 违规反思修订复用同一 sessionId，agent 即可看到
     * 此前的任务书与自己已产出的成稿。
     *
     * @param message   发送给 agent 的完整消息（任务书 / 续写指令 / 违规反馈）
     * @param sessionId 会话 id（同一生成流程内复用）
     * @return agent 生成的文本；连接失败/超时 → null（调用方降级）
     */
    static String chat(String message, String sessionId) {
        try {
            ObjectNode body = MAPPER.createObjectNode();
            body.put("message", message);
            body.put("session_id", sessionId);

            HttpRequest req = HttpRequest.newBuilder()
                    .uri(URI.create(AGENT_URL))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(600))
                    .POST(HttpRequest.BodyPublishers.ofString(
                            MAPPER.writeValueAsString(body)))
                    .build();

            HttpResponse<InputStream> resp = HTTP.send(req,
                    HttpResponse.BodyHandlers.ofInputStream());

            // 解析 SSE 流：event: <type>\ndata: <json>\n\n
            StringBuilder answer = new StringBuilder();
            String currentEvent = "";
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(resp.body(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (line.startsWith("event: ")) {
                        currentEvent = line.substring(7).trim();
                    } else if (line.startsWith("data: ")) {
                        String json = line.substring(6);
                        JsonNode data = MAPPER.readTree(json);
                        switch (currentEvent) {
                            case "token" -> answer.append(
                                    data.path("content").asText(""));
                            case "result" -> {
                                String ans = data.path("answer").asText("");
                                if (!ans.isEmpty()) {
                                    answer.setLength(0);
                                    answer.append(ans);
                                }
                            }
                            case "error" -> {
                                System.err.println("AgentChatClient: agent error: "
                                        + data.path("error").asText());
                                return null;
                            }
                            default -> { /* phase, hb, done, skill_load, tool — 忽略 */ }
                        }
                    }
                }
            }
            String result = answer.toString().trim();
            return result.isEmpty() ? null : result;
        } catch (Exception e) {
            System.err.println("AgentChatClient: " + e.getMessage());
            return null;
        }
    }
}
