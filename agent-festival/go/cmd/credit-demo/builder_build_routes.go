package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
)

func builderBuildWrite(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func builderBuildError(w http.ResponseWriter, status int) {
	builderBuildWrite(w, status, map[string]any{"error": map[string]any{"code": "builder-build-unavailable"}})
}

func (s *Service) builderBuildRoutes(mux *http.ServeMux, c *builderBuildClient) {
	if c == nil {
		return
	}
	mux.HandleFunc("/api/builder-actions/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" || r.URL.RawPath != "" || len(r.Header.Values(builderBuildHeader)) != 0 || len(r.Header.Values(builderContextHeader)) != 0 {
			builderBuildError(w, 400)
			return
		}
		user, subject, ok := c.user(r)
		if !ok {
			builderBuildError(w, 403)
			return
		}
		same := func(alias string) bool {
			u, sub, ok := c.user(r)
			return ok && u == user && sub == subject && (alias == "" || c.allowed(user, subject, alias))
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/builder-actions/")
		if path == "build" {
			if r.Method != "POST" {
				builderBuildError(w, 405)
				return
			}
			if len(r.Header.Values("Origin")) != 1 || !builderOrigin(r.Header.Get("Origin"), s.cfg.Origins) {
				builderBuildError(w, 403)
				return
			}
			raw, e := io.ReadAll(io.LimitReader(r.Body, 1025))
			if e != nil || len(raw) > 1024 {
				builderBuildError(w, 400)
				return
			}
			m, e := builderJSON(raw)
			if e != nil || !builderKeys(m, "alias") {
				builderBuildError(w, 400)
				return
			}
			alias, ok := m["alias"].(string)
			if !ok || !same(alias) {
				builderBuildError(w, 403)
				return
			}
			row, first, e := s.builderPrepare(c, user, subject, alias)
			if e != nil {
				builderBuildError(w, 409)
				return
			}
			if !same(alias) {
				builderBuildError(w, 403)
				return
			}
			method, upstream, body := "GET", builderPrefix+"commands/"+row.CommandID, []byte(nil)
			if first {
				method, upstream, body = "POST", builderPrefix+"commands", row.Material
			}
			receipt, e := c.request(r.Context(), subject, method, upstream, body)
			if errors.Is(e, errBuilderBuildDenied) {
				builderBuildError(w, 403)
				return
			}
			if !same(alias) {
				builderBuildError(w, 403)
				return
			}
			out := map[string]any{"command_id": row.CommandID, "alias": alias, "local_status": "unknown", "receipt": nil}
			if e == nil && builderBuildReceipt(receipt, row) {
				out["receipt"] = receipt
			}
			builderBuildWrite(w, 200, out)
			return
		}
		if r.Method != "GET" {
			builderBuildError(w, 405)
			return
		}
		body, e := io.ReadAll(io.LimitReader(r.Body, 1))
		if e != nil || len(body) != 0 {
			builderBuildError(w, 400)
			return
		}
		if path == "builds" {
			rows, e := s.builderIntentList(c, user, subject)
			if e != nil {
				builderBuildError(w, 503)
				return
			}
			aliases := []string{}
			for alias := range c.builds {
				if c.allowed(user, subject, alias) {
					aliases = append(aliases, alias)
				}
			}
			sort.Strings(aliases)
			list := []map[string]any{}
			for _, row := range rows {
				if same(row.Alias) {
					list = append(list, map[string]any{"alias": row.Alias, "command_id": row.CommandID, "local_status": "unknown"})
				}
			}
			if !same("") {
				builderBuildError(w, 403)
				return
			}
			// Recheck every alias at disclosure, including registrations not submitted.
			for _, alias := range aliases {
				if !same(alias) {
					builderBuildError(w, 403)
					return
				}
			}
			for _, entry := range list {
				if !same(entry["alias"].(string)) {
					builderBuildError(w, 403)
					return
				}
			}
			builderBuildWrite(w, 200, map[string]any{"aliases": aliases, "intents": list})
			return
		}
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] != "commands" || !builderToken.MatchString(parts[1]) {
			builderBuildError(w, 404)
			return
		}
		row, e := s.builderIntent(c, user, subject, parts[1])
		if e != nil || !same(row.Alias) {
			builderBuildError(w, 403)
			return
		}
		out := map[string]any{"alias": row.Alias, "command_id": row.CommandID, "local_status": "unknown", "receipt": nil, "operation": nil, "evidence": nil}
		receipt, e := c.request(r.Context(), subject, "GET", builderPrefix+"commands/"+row.CommandID, nil)
		if errors.Is(e, errBuilderBuildDenied) {
			builderBuildError(w, 403)
			return
		}
		if !same(row.Alias) {
			builderBuildError(w, 403)
			return
		}
		if e == nil && builderBuildReceipt(receipt, row) {
			out["receipt"] = receipt
			if ref, ok := receipt["operation_ref"].(map[string]any); ok {
				operation, e := c.request(r.Context(), subject, "GET", builderPrefix+"operations/"+ref["id"].(string), nil)
				if errors.Is(e, errBuilderBuildDenied) {
					builderBuildError(w, 403)
					return
				}
				if !same(row.Alias) {
					builderBuildError(w, 403)
					return
				}
				if e == nil && builderBuildOperation(operation, ref["id"].(string)) && builderBuildSameRef(operation["operation_ref"], ref) && (receipt["task_ref"] == nil || builderBuildSameRef(receipt["task_ref"], operation["task_ref"])) && (receipt["candidate_ref"] == nil || builderBuildSameRef(receipt["candidate_ref"], operation["candidate_ref"])) {
					// All reads derive from current original Host receipts, never browser IDs.
					out["operation"] = operation
					candidate, cok := operation["candidate_ref"].(map[string]any)
					if ev, eok := operation["evidence_ref"].(map[string]any); eok && cok {
						evidence, e := c.request(r.Context(), subject, "GET", builderPrefix+"evidence/"+ev["id"].(string), nil)
						if errors.Is(e, errBuilderBuildDenied) {
							builderBuildError(w, 403)
							return
						}
						if !same(row.Alias) {
							builderBuildError(w, 403)
							return
						}
						originalCommand, commandError := builderJSON(row.Material)
						if e == nil && commandError == nil && evidence["source_revision"] == originalCommand["expected_revision"] && builderBuildEvidence(evidence, ev, candidate) {
							out["evidence"] = evidence
						}
					}
				}
			}
		}
		if !same(row.Alias) {
			builderBuildError(w, 403)
			return
		}
		builderBuildWrite(w, 200, out)
	})
}
