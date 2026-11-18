/*
 * Critic Agent 配置 — 语义判定角色（temperature 强制 0）。
 */

package com.agenthub.w06.checker;

public final class CriticAgentProperties extends com.agenthub.w06.agents.support.AgentProperties {
    private int maxReplan = 1;
    private java.util.List<String> criteria = java.util.List.of();

    public int getMaxReplan() { return maxReplan; }
    public void setMaxReplan(int v) { this.maxReplan = v; }
    public java.util.List<String> getCriteria() { return criteria; }
    public void setCriteria(java.util.List<String> v) { this.criteria = v; }
}
