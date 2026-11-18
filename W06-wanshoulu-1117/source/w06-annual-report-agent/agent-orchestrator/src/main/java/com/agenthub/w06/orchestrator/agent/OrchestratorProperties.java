/*
 * Orchestrator 配置 — root DeepAgent。
 */

package com.agenthub.w06.orchestrator.agent;

import com.agenthub.w06.agents.support.AgentProperties;
import java.time.Duration;

public final class OrchestratorProperties extends AgentProperties {
    private int gatekeepMaxRounds = 1;
    private String workspacePath = "target/w06-workspace";
    private String workspaceLanguage = "zh-CN";
    private boolean enableTaskLoop = true;
    private Duration completionTimeout = Duration.ofMinutes(15);

    public int getGatekeepMaxRounds() { return gatekeepMaxRounds; }
    public void setGatekeepMaxRounds(int v) { this.gatekeepMaxRounds = v; }
    public String getWorkspacePath() { return workspacePath; }
    public void setWorkspacePath(String v) { this.workspacePath = v; }
    public String getWorkspaceLanguage() { return workspaceLanguage; }
    public void setWorkspaceLanguage(String v) { this.workspaceLanguage = v; }
    public boolean isEnableTaskLoop() { return enableTaskLoop; }
    public void setEnableTaskLoop(boolean v) { this.enableTaskLoop = v; }
    @Override
    public Duration getCompletionTimeout() { return completionTimeout; }
    @Override
    public void setCompletionTimeout(Duration v) { this.completionTimeout = v; }

    /** DeepAgentConfig.model(...) 需要的有序 map。 */
    public java.util.Map<String, Object> modelConfig() {
        java.util.Map<String, Object> model = new java.util.LinkedHashMap<>();
        model.put("model", getModelName());
        model.put("temperature", getTemperature());
        model.put("top_p", getTopP());
        return model;
    }

    /** DeepAgentConfig.backend(...) 需要的有序 map。 */
    public java.util.Map<String, Object> backendConfig() {
        java.util.Map<String, Object> backend = new java.util.LinkedHashMap<>();
        backend.put("provider", getProvider());
        backend.put("api_key", getApiKey());
        backend.put("api_base", getApiBase());
        backend.put("verify_ssl", isSslVerify());
        backend.put("timeout", getCompletionTimeout().toSeconds());
        return backend;
    }
}
