package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

// serveResume exposes the existing human resume action to the WorkSwarm plugin.
// Demo credentials are checked here; a Gateway/client identity is not authority.
// This queues the original run, never creates a plan or sends an Agent request.
func (s *Service) serveResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	reply := func(code int, body any) {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		reply(http.StatusMethodNotAllowed, map[string]string{"error": "只接受 POST"})
		return
	}
	var in struct {
		Run      string `json:"run"`
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
		reply(http.StatusBadRequest, map[string]string{"error": "请求格式不对"})
		return
	}
	if !validLogin(in.User, in.Password) {
		reply(http.StatusUnauthorized, map[string]string{"error": "账号或密码不对"})
		return
	}
	if err := s.ResumeRun(in.Run); err != nil {
		if errors.Is(err, errStorage) {
			reply(http.StatusServiceUnavailable, map[string]string{"error": "状态保存失败，未接受恢复；请联系维护人员"})
		} else {
			reply(http.StatusConflict, map[string]string{"error": "任务不存在或已不在等待状态，请刷新查看原任务"})
		}
		return
	}
	reply(http.StatusOK, map[string]string{"run": in.Run, "status": "queued"})
}
