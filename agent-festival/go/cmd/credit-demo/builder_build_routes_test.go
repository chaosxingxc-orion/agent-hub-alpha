package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func builderBuildBrowser(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "cd_session", Value: sign("alice")})
	r.Header.Set("Origin", "http://127.0.0.1:8123")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}

func TestBuilderBuildRoutesLostReplyAnd404NeverRedispatch(t *testing.T) {
	var posts, gets atomic.Int64
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			posts.Add(1)
			_, _ = w.Write([]byte("lost receipt"))
		} else {
			gets.Add(1)
			w.WriteHeader(404)
		}
	}))
	defer up.Close()
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderBuildRoutes(mux, c)
	for i := 0; i < 3; i++ {
		w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"local_status":"unknown"`) {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	w := builderBuildBrowser(mux, "GET", "/api/builder-actions/builds", "")
	var listed struct {
		Intents []struct {
			CommandID string `json:"command_id"`
		}
	}
	_ = json.Unmarshal(w.Body.Bytes(), &listed)
	if w.Code != 200 || len(listed.Intents) != 1 {
		t.Fatal("new browser cannot recover durable association")
	}
	w = builderBuildBrowser(mux, "GET", "/api/builder-actions/commands/"+listed.Intents[0].CommandID, "")
	if w.Code != 200 || posts.Load() != 1 || gets.Load() != 3 {
		t.Fatal("unknown/404 caused another send", posts.Load(), gets.Load(), w.Code)
	}
}

func TestBuilderBuildRoutesRejectOverridesAndQueryProof(t *testing.T) {
	var calls atomic.Int64
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer up.Close()
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderBuildRoutes(mux, c)
	for _, body := range []string{`{"alias":"request","request_id":"another-browser-key"}`, `{"alias":"request","command_id":"forged"}`, `{"alias":"request","alias":"request"}`, `{"alias":"missing"}`} {
		if w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", body); w.Code < 400 {
			t.Fatal("override accepted", body)
		}
	}
	r := httptest.NewRequest("POST", "/api/builder-actions/build", strings.NewReader(`{"alias":"request"}`))
	r.AddCookie(&http.Cookie{Name: "cd_session", Value: sign("alice")})
	r.Header.Set("Origin", "http://evil.invalid")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross origin")
	}
	r.Header.Set("Origin", "http://127.0.0.1:8123")
	r.Header.Set(builderContextHeader, "query-proof")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 400 || calls.Load() != 0 || len(s.st.BuilderActions) != 0 {
		t.Fatal("untrusted input caused a dispatch or save")
	}
}

func TestBuilderBuildRoutesRevocationAfterSaveAndAfterSend(t *testing.T) {
	var calls atomic.Int64
	var allowed atomic.Bool
	allowed.Store(true)
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		allowed.Store(false)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer up.Close()
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	c.visible = func(string, string, string, string) bool { return allowed.Load() }
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderBuildRoutes(mux, c)
	original := s.writeState
	s.writeState = func(path string, v any) error { e := original(path, v); allowed.Store(false); return e }
	w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`)
	if w.Code != 403 || calls.Load() != 0 || len(s.st.BuilderActions) != 1 {
		t.Fatal("revocation before send")
	}
	allowed.Store(true)
	s.writeState = original
	w = builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`)
	if w.Code != 403 || calls.Load() != 1 || !strings.Contains(w.Body.String(), "unavailable") {
		t.Fatal("post-send revocation disclosed source")
	}
	if strings.Contains(w.Body.String(), "h04.") {
		t.Fatal("revoked command disclosed")
	}
	allowed.Store(true)
	c.builds["new-owner-request"] = c.builds["request"]
	w = builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"new-owner-request"}`)
	if w.Code != 403 || calls.Load() != 2 || len(s.st.BuilderActions) != 2 || strings.Contains(w.Body.String(), "h04.") {
		t.Fatal("fresh post response leaked after revocation")
	}
}

func TestBuilderBuildRoutesOriginalReceiptOperationEvidenceBinding(t *testing.T) {
	var command, digest string
	var tamper atomic.Bool
	ref := func(id string) map[string]any { return map[string]any{"id": id, "sha256": builderBuildSHA([]byte(id))} }
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var m map[string]any
			d := json.NewDecoder(r.Body)
			d.UseNumber()
			_ = d.Decode(&m)
			command = m["command_id"].(string)
			raw, _ := builderBuildCanonical(m)
			_, _, digest, _ = builderBuildMaterial(raw, true)
		}
		var value map[string]any
		switch {
		case strings.Contains(r.URL.Path, "/commands"):
			value = map[string]any{"schema_version": "builder-hub/v1", "command_id": command, "content_digest": digest, "status": "accepted", "operation_ref": ref("original-operation"), "task_ref": ref("original-task"), "candidate_ref": ref("original-candidate")}
		case strings.Contains(r.URL.Path, "/operations/"):
			operationRef := ref("original-operation")
			if tamper.Load() {
				operationRef["sha256"] = strings.Repeat("e", 64)
			}
			value = map[string]any{"schema_version": "builder-hub/v1", "operation_ref": operationRef, "scenario": map[string]any{"id": "intent", "version": "1.0", "profile": "gate"}, "generation": 0, "status": "completed", "quality": "not_run", "cleanup": "unknown", "task_ref": ref("original-task"), "candidate_ref": ref("original-candidate"), "evidence_ref": ref("original-evidence")}
		default:
			value = map[string]any{"schema_version": "builder-hub/v1", "evidence_ref": ref("original-evidence"), "scenario": map[string]any{"id": "intent", "version": "1.0", "profile": "gate"}, "kind": "engineering", "source_revision": strings.Repeat("a", 40), "status": "verified", "candidate_ref": ref("original-candidate"), "run_ref": nil}
		}
		_ = json.NewEncoder(w).Encode(value)
	}))
	defer up.Close()
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderBuildRoutes(mux, c)
	if w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	before := s.st.Rev
	w := builderBuildBrowser(mux, "GET", "/api/builder-actions/commands/"+command, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"kind":"engineering"`) || !strings.Contains(w.Body.String(), `"quality":"not_run"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	tamper.Store(true)
	w = builderBuildBrowser(mux, "GET", "/api/builder-actions/commands/"+command, "")
	if !strings.Contains(w.Body.String(), `"operation":null`) || !strings.Contains(w.Body.String(), `"evidence":null`) || s.st.Rev != before {
		t.Fatal("wrong operation source promoted or GET wrote")
	}
}

func TestBuilderBuildRoutesDisabledAndStorageUnavailable(t *testing.T) {
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderBuildRoutes(mux, nil)
	if w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`); w.Code != 404 {
		t.Fatal("disabled route mounted")
	}
	c, _ := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": builderBuildTemplate()})
	mux = http.NewServeMux()
	s.builderBuildRoutes(mux, c)
	s.storageErr = errStorage
	if w := builderBuildBrowser(mux, "GET", "/api/builder-actions/builds", ""); w.Code != 503 {
		t.Fatal("storage error presented as empty history")
	}
}

func TestBuilderBuildOriginalPageMountAndHostDenial(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }))
	defer up.Close()
	q, _ := newBuilderHubClient(builderTestConfig(up.URL))
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	mux := http.NewServeMux()
	s.builderHubRoutes(mux, q)
	if w := builderBuildBrowser(mux, "GET", "/ui/builder", ""); strings.Contains(w.Body.String(), `id="builder-build"`) {
		t.Fatal("default-off UI exposed writes")
	}
	s.cfg.BuilderBuild = c
	if w := builderBuildBrowser(mux, "GET", "/ui/builder", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `id="build-history"`) || strings.Contains(w.Body.String(), "localStorage") {
		t.Fatal("original page has no server recovery surface", w.Code)
	}
	if w := builderBuildBrowser(mux, "GET", "/ui/builder", ""); strings.Count(w.Body.String(), `id="builder-build"`) != 1 {
		t.Fatal("build panel must be unique")
	}
	s.builderBuildRoutes(mux, c)
	w := builderBuildBrowser(mux, "POST", "/api/builder-actions/build", `{"alias":"request"}`)
	if w.Code != 403 || strings.Contains(w.Body.String(), "h04.") || len(s.st.BuilderActions) != 1 {
		t.Fatal("original Host rejection disclosed command or erased uncertainty")
	}
}

func TestBuilderBuildPrivateHistoryNeverDebugExport(t *testing.T) {
	for _, mode := range []string{"enabled", "disabled-history", "disabled-typed-history", "legacy"} {
		t.Run(mode, func(t *testing.T) {
			s := builderBuildTestService(t, "")
			if mode == "enabled" {
				s.cfg.BuilderBuild, _ = newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": builderBuildTemplate()})
			}
			if mode == "disabled-history" {
				s.st.BuilderActions = map[string]builderActionIntent{"private-command": {User: "another-user", Subject: "private-subject", Material: json.RawMessage(`{"private":"material"}`)}}
			}
			if mode == "disabled-typed-history" {
				s.st.Plans["private-plan"] = []*Plan{{ID: "private-plan", ExecutionBinding: &H03ExecutionBinding{Profile: "two-role-identity/v1"}}}
			}
			mux := http.NewServeMux()
			s.routes(mux)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("GET", "/debug/state", nil))
			if mode == "legacy" {
				if w.Code != 200 {
					t.Fatal("unchanged legacy debug unexpectedly removed")
				}
				return
			}
			if w.Code < 400 || strings.Contains(w.Body.String(), "private-") || strings.Contains(w.Body.String(), "material") {
				t.Fatal("private integration state exposed through debug", w.Code, w.Body.String())
			}
		})
	}
}
