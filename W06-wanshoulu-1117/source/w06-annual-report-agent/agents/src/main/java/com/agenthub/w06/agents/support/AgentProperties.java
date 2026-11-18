/*
 * AgentProperties — W06 角色 agent 的通用配置（库层，无 Spring 绑定）。
 * 照上游 VerifyAgentProperties 的模式：requireConfigured 校验必填项。
 */

package com.agenthub.w06.agents.support;

import java.time.Duration;

/**
 * 五个专业角色共用的最小配置集。wrapper 层从 application.yml 绑定后传入库层工厂。
 * modelClient 各参数对应上游 ReActAgentConfig.configureModelClient(...) 的入参。
 */
public class AgentProperties {

    private String agentId;
    private String agentName;
    private String provider;
    private String apiKey;
    private String apiBase;
    private String modelName;
    private boolean sslVerify = true;
    private double temperature = 0.2;
    private double topP = 0.8;
    private int maxIterations = 8;
    private String sysOperationId;
    private Duration completionTimeout = Duration.ofMinutes(10);

    public void requireConfigured() {
        if (agentId == null || agentId.isBlank()) {
            throw new IllegalStateException("agentId is required");
        }
        if (provider == null || provider.isBlank()) {
            throw new IllegalStateException("provider is required");
        }
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalStateException("apiKey is required");
        }
        if (modelName == null || modelName.isBlank()) {
            throw new IllegalStateException("modelName is required");
        }
    }

    public String getAgentId() { return agentId; }
    public void setAgentId(String v) { this.agentId = v; }
    public String getAgentName() { return agentName; }
    public void setAgentName(String v) { this.agentName = v; }
    public String getProvider() { return provider; }
    public void setProvider(String v) { this.provider = v; }
    public String getApiKey() { return apiKey; }
    public void setApiKey(String v) { this.apiKey = v; }
    public String getApiBase() { return apiBase; }
    public void setApiBase(String v) { this.apiBase = v; }
    public String getModelName() { return modelName; }
    public void setModelName(String v) { this.modelName = v; }
    public boolean isSslVerify() { return sslVerify; }
    public void setSslVerify(boolean v) { this.sslVerify = v; }
    public double getTemperature() { return temperature; }
    public void setTemperature(double v) { this.temperature = v; }
    public double getTopP() { return topP; }
    public void setTopP(double v) { this.topP = v; }
    public int getMaxIterations() { return maxIterations; }
    public void setMaxIterations(int v) { this.maxIterations = v; }
    public String getSysOperationId() { return sysOperationId; }
    public void setSysOperationId(String v) { this.sysOperationId = v; }
    public Duration getCompletionTimeout() { return completionTimeout; }
    public void setCompletionTimeout(Duration v) { this.completionTimeout = v; }
}
