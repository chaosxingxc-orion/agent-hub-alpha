package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const builderBuildHeader = "X-Builder-Hub-Action-Context"

var errBuilderBuildDenied = errors.New("builder build access denied")

// The action channel is independent of the existing read-only query channel.
// Registered aliases are immutable source requests, never browser input paths.
type builderBuildClient struct {
	cfg     builderHubConfig
	http    *http.Client
	now     func() time.Time
	builds  map[string]json.RawMessage
	visible func(user, subject, scope, alias string) bool
}

func builderBuildCanonical(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		return nil, errBuilderHub
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

func builderBuildSHA(raw []byte) string {
	s := sha256.Sum256(raw)
	return hex.EncodeToString(s[:])
}

func builderBuildRef(v any) bool {
	m, ok := v.(map[string]any)
	if !ok || !builderKeys(m, "id", "sha256") {
		return false
	}
	id, a := m["id"].(string)
	digest, b := m["sha256"].(string)
	return a && b && builderToken.MatchString(id) && builderDigest.MatchString(digest)
}

func builderBuildInteger(v any) bool {
	n, ok := v.(json.Number)
	if !ok {
		return false
	}
	x, e := n.Int64()
	return e == nil && x >= 0 && x <= 9007199254740991
}

// Mirrors the original normalized Python Command material. All accepted string
// values are finite ASCII identifiers. No NUL separator belongs to this hash.
func builderBuildMaterial(raw []byte, withID bool) (map[string]any, []byte, string, error) {
	m, e := builderJSON(raw)
	if e != nil {
		return nil, nil, "", errBuilderHub
	}
	allowed := map[string]bool{"schema_version": true, "scenario": true, "operation": true, "input_ref": true, "expected_revision": true, "expected_generation": true}
	if withID {
		allowed["command_id"] = true
	}
	for k := range m {
		if !allowed[k] {
			return nil, nil, "", errBuilderHub
		}
	}
	if m["schema_version"] != "builder-hub/v1" || m["operation"] != "build" || !builderScenario(m["scenario"]) {
		return nil, nil, "", errBuilderHub
	}
	for _, k := range []string{"input_ref", "expected_revision", "expected_generation"} {
		if _, ok := m[k]; !ok {
			m[k] = nil
		}
	}
	if m["input_ref"] != nil && !builderBuildRef(m["input_ref"]) {
		return nil, nil, "", errBuilderHub
	}
	if m["expected_revision"] != nil {
		s, ok := m["expected_revision"].(string)
		if !ok || !builderToken.MatchString(s) {
			return nil, nil, "", errBuilderHub
		}
	}
	if m["expected_generation"] != nil && !builderBuildInteger(m["expected_generation"]) {
		return nil, nil, "", errBuilderHub
	}
	if withID {
		id, ok := m["command_id"].(string)
		if !ok || !builderToken.MatchString(id) {
			return nil, nil, "", errBuilderHub
		}
	}
	material := map[string]any{}
	for k, v := range m {
		if k != "command_id" {
			material[k] = v
		}
	}
	canonical, e := builderBuildCanonical(material)
	return m, canonical, builderBuildSHA(canonical), e
}

func newBuilderBuildClient(cfg builderHubConfig, templates map[string]json.RawMessage) (*builderBuildClient, error) {
	q, e := newBuilderHubClient(cfg)
	if e != nil || len(templates) == 0 || len(templates) > 64 {
		return nil, errBuilderHub
	}
	c := &builderBuildClient{cfg: q.cfg, http: q.http, now: time.Now, builds: map[string]json.RawMessage{}}
	for alias, raw := range templates {
		if !builderToken.MatchString(alias) {
			return nil, errBuilderHub
		}
		m, canonical, _, e := builderBuildMaterial(raw, false)
		rev, ok := m["expected_revision"].(string)
		generation, gok := m["expected_generation"].(json.Number)
		if e != nil || !ok || !builderRevision.MatchString(rev) || !gok || generation.String() != "0" || !builderBuildRef(m["input_ref"]) {
			return nil, errBuilderHub
		}
		c.builds[alias] = bytes.Clone(canonical)
	}
	return c, nil
}

func builderBuildFromEnv(query *builderHubClient) (*builderBuildClient, error) {
	if os.Getenv("BUILDER_BUILD_ENABLED") == "" {
		return nil, nil
	}
	if os.Getenv("BUILDER_BUILD_ENABLED") != "1" || query == nil {
		return nil, errBuilderHub
	}
	password := os.Getenv("CREDIT_DEMO_PASSWORD")
	key, e := hex.DecodeString(os.Getenv("BUILDER_BUILD_KEY_HEX"))
	if e != nil || strings.TrimSpace(password) == "" || password == "demo" || bytes.Equal(key, query.cfg.Key) || bytes.Equal(key, sessionKey) || bytes.Equal(key, []byte(password)) {
		return nil, errBuilderHub
	}
	m, e := builderJSON([]byte(os.Getenv("BUILDER_BUILD_USERS_JSON")))
	if e != nil || len(m) == 0 || len(m) > 64 {
		return nil, errBuilderHub
	}
	mapping := map[string]string{}
	for user, v := range m {
		sub, ok := v.(string)
		if !ok || !builderIdentifier.MatchString(user) || !builderIdentifier.MatchString(sub) {
			return nil, errBuilderHub
		}
		if _, ok := users[user]; !ok {
			return nil, errBuilderHub
		}
		mapping[user] = sub
	}
	registered, e := builderJSON([]byte(os.Getenv("BUILDER_BUILD_TEMPLATES_JSON")))
	if e != nil {
		return nil, e
	}
	templates := map[string]json.RawMessage{}
	for alias, v := range registered {
		templates[alias], e = builderBuildCanonical(v)
		if e != nil {
			return nil, e
		}
	}
	cfg := builderHubConfig{Target: os.Getenv("BUILDER_BUILD_URL"), ChannelToken: os.Getenv("BUILDER_BUILD_CHANNEL_TOKEN"), Issuer: os.Getenv("BUILDER_BUILD_ISSUER"), KeyID: os.Getenv("BUILDER_BUILD_KEY_ID"), Audience: os.Getenv("BUILDER_BUILD_AUDIENCE"), Scope: os.Getenv("BUILDER_BUILD_SCOPE"), Key: key, LookupSubject: func(user string) (string, bool) { sub, ok := mapping[user]; return sub, ok }}
	if strings.TrimSuffix(cfg.Target, "/") != query.cfg.Target || cfg.Scope != query.cfg.Scope || cfg.Audience != query.cfg.Audience || cfg.Issuer != query.cfg.Issuer {
		return nil, errBuilderHub
	}
	return newBuilderBuildClient(cfg, templates)
}

func builderBuildProof(key []byte, claims map[string]any) (string, error) {
	raw, e := builderBuildCanonical(claims)
	if e != nil {
		return "", e
	}
	m := hmac.New(sha256.New, key)
	m.Write([]byte("builder-hub-action-context/v1\n"))
	m.Write(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + hex.EncodeToString(m.Sum(nil)), nil
}

func (c *builderBuildClient) user(r *http.Request) (string, string, bool) {
	return (&builderHubClient{cfg: c.cfg}).user(r)
}
func (c *builderBuildClient) allowed(user, subject, alias string) bool {
	_, ok := c.builds[alias]
	return ok && (c.visible == nil || c.visible(user, subject, c.cfg.Scope, alias))
}

func (c *builderBuildClient) request(ctx context.Context, subject, method, path string, body []byte) (map[string]any, error) {
	if !builderIdentifier.MatchString(subject) {
		return nil, errBuilderHub
	}
	action := ""
	if method == "POST" && path == builderPrefix+"commands" {
		if _, _, _, e := builderBuildMaterial(body, true); e != nil {
			return nil, e
		}
		action = "build"
	} else if method == "GET" && len(body) == 0 {
		for _, part := range []string{"commands/", "operations/", "evidence/"} {
			if strings.HasPrefix(path, builderPrefix+part) && builderToken.MatchString(strings.TrimPrefix(path, builderPrefix+part)) {
				action = "read-operation"
				if part == "evidence/" {
					action = "read-evidence"
				}
			}
		}
	}
	if action == "" || !builderIdentifier.MatchString(subject) || len(body) > 65536 {
		return nil, errBuilderHub
	}
	now := c.now().Unix()
	nonce := make([]byte, 16)
	if _, e := rand.Read(nonce); e != nil {
		return nil, errBuilderHub
	}
	claims := map[string]any{"version": "builder-hub-action-context/v1", "issuer": c.cfg.Issuer, "key_id": c.cfg.KeyID, "audience": c.cfg.Audience, "subject": subject, "scope": c.cfg.Scope, "action": action, "method": method, "path": path, "body_sha256": builderBuildSHA(body), "issued_at": now, "expires_at": now + 30, "request_id": hex.EncodeToString(nonce)}
	proof, e := builderBuildProof(c.cfg.Key, claims)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, method, c.cfg.Target+path, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.ChannelToken)
	req.Header.Set(builderBuildHeader, proof)
	req.Header.Set("Content-Type", "application/json")
	resp, e := c.http.Do(req)
	if e != nil {
		return nil, errBuilderHub
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errBuilderBuildDenied
	}
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return nil, errBuilderHub
	}
	raw, e := io.ReadAll(io.LimitReader(resp.Body, 65537))
	if e != nil || len(raw) > 65536 || c.now().Unix() >= now+30 {
		return nil, errBuilderHub
	}
	return builderJSON(raw)
}

func builderBuildReceipt(m map[string]any, row builderActionIntent) bool {
	if !builderKeys(m, "schema_version", "command_id", "content_digest", "status", "operation_ref", "task_ref", "candidate_ref") || m["schema_version"] != "builder-hub/v1" || m["command_id"] != row.CommandID || m["content_digest"] != row.Digest || !builderOne(m["status"], "accepted", "pending", "duplicate", "conflict", "rejected", "unknown") {
		return false
	}
	for _, k := range []string{"operation_ref", "task_ref", "candidate_ref"} {
		if m[k] != nil && !builderBuildRef(m[k]) {
			return false
		}
	}
	return true
}

func builderBuildOperation(m map[string]any, id string) bool {
	if !builderKeys(m, "schema_version", "operation_ref", "scenario", "generation", "status", "quality", "cleanup", "task_ref", "candidate_ref", "evidence_ref") || m["schema_version"] != "builder-hub/v1" || !builderScenario(m["scenario"]) || !builderBuildInteger(m["generation"]) || !builderBuildRef(m["operation_ref"]) {
		return false
	}
	if m["operation_ref"].(map[string]any)["id"] != id || !builderOne(m["status"], "running", "waiting", "failed", "cancelled", "stopped", "completed", "unknown") || !builderOne(m["quality"], "not_run", "passed", "failed", "indeterminate") || !builderOne(m["cleanup"], "not_run", "pending", "complete", "failed", "unknown") {
		return false
	}
	for _, k := range []string{"task_ref", "candidate_ref", "evidence_ref"} {
		if m[k] != nil && !builderBuildRef(m[k]) {
			return false
		}
	}
	return true
}

func builderBuildEvidence(m map[string]any, ref, candidate map[string]any) bool {
	if !builderKeys(m, "schema_version", "evidence_ref", "scenario", "kind", "source_revision", "status", "candidate_ref", "run_ref") || m["schema_version"] != "builder-hub/v1" || !builderScenario(m["scenario"]) || m["kind"] != "engineering" || !builderOne(m["status"], "verified", "unavailable", "unverified") {
		return false
	}
	rev, ok := m["source_revision"].(string)
	if !ok || !builderRevision.MatchString(rev) {
		return false
	}
	a, _ := builderBuildCanonical(m["evidence_ref"])
	b, _ := builderBuildCanonical(ref)
	x, _ := builderBuildCanonical(m["candidate_ref"])
	y, _ := builderBuildCanonical(candidate)
	return builderBuildRef(m["evidence_ref"]) && builderBuildRef(m["candidate_ref"]) && bytes.Equal(a, b) && bytes.Equal(x, y) && (m["run_ref"] == nil || builderBuildRef(m["run_ref"]))
}

func builderBuildSameRef(a, b any) bool {
	if !builderBuildRef(a) || !builderBuildRef(b) {
		return false
	}
	x, _ := builderBuildCanonical(a)
	y, _ := builderBuildCanonical(b)
	return bytes.Equal(x, y)
}
