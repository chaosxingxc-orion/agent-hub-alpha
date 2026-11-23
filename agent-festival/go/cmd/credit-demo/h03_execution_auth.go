package main

import (
	"bytes"
	"crypto/subtle"
	"io"
	"net/http"
	"strings"
)

const h03RuntimeHeader = "X-Hub-Runtime-Credential"

// The actual main HTTP surface uses this wrapper. A role capability token is
// never a runtime credential. Disabled legacy deployments keep their behavior.
func (s *Service) h03ControlHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.H03Profile != nil {
			path := r.URL.Path
			if strings.HasPrefix(path, "/exec/") {
				credential := r.Header.Get(h03RuntimeHeader)
				if r.Method != http.MethodPost || len(s.cfg.H03RuntimeCredential) < 32 || subtle.ConstantTimeCompare([]byte(credential), []byte(s.cfg.H03RuntimeCredential)) != 1 {
					http.Error(w, "runtime identity required", http.StatusUnauthorized)
					return
				}
				allowed := map[string]bool{"/exec/lease": true, "/exec/heartbeat": true, "/exec/call/begin": true, "/exec/call/update": true, "/exec/call/commit": true, "/exec/done": true}
				if !allowed[path] {
					http.NotFound(w, r)
					return
				}
			}
			if strings.HasPrefix(path, "/debug/") || strings.HasPrefix(path, "/member-mcp/") || path == "/agent/model" {
				http.NotFound(w, r)
				return
			}
			if strings.HasPrefix(path, "/exec/") || path == "/agent/context" {
				raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 262144))
				if err != nil {
					http.Error(w, "bounded execution JSON required", http.StatusBadRequest)
					return
				}
				if _, err = builderJSON(raw); err != nil {
					http.Error(w, "strict execution JSON required", http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(raw))
			}
		}
		next.ServeHTTP(w, r)
	})
}
