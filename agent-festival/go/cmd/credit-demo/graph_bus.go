package main

type GraphExecution struct {
	Kind     string      `json:"kind"`
	PathMode string      `json:"path_mode,omitempty"`
	Binding  *BusBinding `json:"binding,omitempty"`
}

type GraphBusSummary struct {
	PlannedNodes int `json:"planned_nodes"`
	Services     int `json:"services"`
	LinkedTasks  int `json:"linked_tasks"`
	Completed    int `json:"completed"`
	Unresolved   int `json:"unresolved"`
	Failed       int `json:"failed"`
	Reused       int `json:"reused"`
}

func busSummary(nodes []GraphNode) *GraphBusSummary {
	s := &GraphBusSummary{}
	services := map[string]bool{}
	for _, node := range nodes {
		if node.Execution.Kind != "agent_bus" {
			continue
		}
		s.PlannedNodes++
		if node.Execution.Binding != nil {
			services[node.Execution.Binding.AgentID] = true
		}
		if task, _ := node.Remote["task_id"].(string); task != "" {
			s.LinkedTasks++
		}
		switch node.Status {
		case "done", "reused":
			s.Completed++
			if node.Status == "reused" {
				s.Reused++
			}
		case "unknown":
			s.Unresolved++
		case "failed":
			s.Failed++
		}
	}
	if s.PlannedNodes == 0 {
		return nil
	}
	s.Services = len(services)
	return s
}

func remoteView(step *StepRec) map[string]any {
	c := step.Call
	view := map[string]any{
		"agent_id": c.Binding.AgentID, "node_execution_id": c.NodeExecutionID, "attempt": c.Attempt,
		"request_id": c.RequestID, "task_id": c.RemoteTaskID, "context_id": c.ContextID, "status": c.Status,
		"input_digest": c.InputDigest, "artifact_id": c.ArtifactID, "artifact_digest": c.ArtifactDigest,
		"model_calls": len(c.ModelEvents), "agent_started_at": c.AgentStartedAt, "updated_at": c.UpdatedAt,
		"intent_saved_at": c.CreatedAt, "committed_at": step.FinishedAt, "reason": c.Reason, "reused_count": step.Reused,
	}
	if c.Discovery != nil {
		view["discovery"] = c.Discovery
	}
	if gateway := gatewayView(c); gateway != nil {
		view["gateway"] = gateway
	}
	return view
}
