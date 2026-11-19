package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func builderBuildTemplate() json.RawMessage {
	return json.RawMessage(`{"schema_version":"builder-hub/v1","scenario":{"id":"intent","version":"1.0","profile":"gate"},"operation":"build","input_ref":{"id":"fixed-gate-v1","sha256":"` + strings.Repeat("b", 64) + `"},"expected_revision":"` + strings.Repeat("a", 40) + `","expected_generation":0}`)
}

func TestBuilderBuildOriginalPythonVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/build-command-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		KeyHex  string `json:"key_hex"`
		Vectors []struct {
			Name     string
			Input    json.RawMessage `json:"input_command"`
			Command  string          `json:"command_utf8"`
			Material string          `json:"canonical_material_utf8"`
			Digest   string          `json:"content_digest"`
			BodySHA  string          `json:"body_sha256"`
			Claims   map[string]any
			Proof    string
		}
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	key, e := hex.DecodeString(fixture.KeyHex)
	if e != nil {
		t.Fatal(e)
	}
	if len(fixture.Vectors) != 5 {
		t.Fatal("missing original denominator")
	}
	for _, v := range fixture.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			m, canonical, digest, e := builderBuildMaterial(v.Input, true)
			if e != nil {
				t.Fatal(e)
			}
			command, _ := builderBuildCanonical(m)
			if string(canonical) != v.Material || string(command) != v.Command || digest != v.Digest || builderBuildSHA(command) != v.BodySHA {
				t.Fatal("original normalized bytes or no-NUL digest differ")
			}
			proof, e := builderBuildProof(key, v.Claims)
			if e != nil || proof != v.Proof {
				t.Fatal("action proof differs")
			}
		})
	}
}

func TestBuilderBuildTemplatesAreFiniteAndCopied(t *testing.T) {
	for _, raw := range []string{`{}`, strings.Replace(string(builderBuildTemplate()), `"operation":"build"`, `"operation":"evaluate"`, 1), strings.Replace(string(builderBuildTemplate()), `"expected_generation":0`, `"expected_generation":1`, 1), strings.TrimSuffix(string(builderBuildTemplate()), "}") + `,"command_id":"browser-key"}`} {
		if _, e := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": json.RawMessage(raw)}); e == nil {
			t.Fatal("untrusted template accepted")
		}
	}
	raw := builderBuildTemplate()
	c, e := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": raw})
	if e != nil {
		t.Fatal(e)
	}
	before := bytes.Clone(c.builds["request"])
	raw[0] = 'x'
	if !bytes.Equal(before, c.builds["request"]) {
		t.Fatal("caller can mutate source")
	}
}

func TestBuilderBuildTransportOneSendNoRedirectOrBadMethod(t *testing.T) {
	var first, second atomic.Int64
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { second.Add(1) }))
	defer other.Close()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { first.Add(1); http.Redirect(w, r, other.URL, 307) }))
	defer up.Close()
	c, e := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	if e != nil {
		t.Fatal(e)
	}
	m, _, _, _ := builderBuildMaterial(c.builds["request"], false)
	m["command_id"] = "fixed-command"
	body, _ := builderBuildCanonical(m)
	if _, e = c.request(context.Background(), "alice", "POST", builderPrefix+"commands", body); e == nil {
		t.Fatal("redirect accepted")
	}
	if first.Load() != 1 || second.Load() != 0 {
		t.Fatal("request retried or redirected")
	}
	for _, p := range []string{builderPrefix + "commands/x?override=1", builderPrefix + "supervision/x", builderPrefix + "operations/a/b"} {
		if _, e = c.request(context.Background(), "alice", "GET", p, nil); e == nil {
			t.Fatal("arbitrary path")
		}
	}
	if first.Load() != 1 {
		t.Fatal("unsafe read sent")
	}
}

func TestBuilderBuildResponseLimitsAndExpiry(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"same":1,"same":2}`)) }))
	defer up.Close()
	c, _ := newBuilderBuildClient(builderTestConfig(up.URL), map[string]json.RawMessage{"request": builderBuildTemplate()})
	if _, e := c.request(context.Background(), "alice", "GET", builderPrefix+"commands/id", nil); e == nil {
		t.Fatal("duplicate JSON accepted")
	}
	now := time.Unix(1700000000, 0)
	calls := 0
	c.now = func() time.Time {
		calls++
		if calls > 1 {
			return now.Add(time.Minute)
		}
		return now
	}
	if _, e := c.request(context.Background(), "alice", "GET", builderPrefix+"commands/id", nil); e == nil {
		t.Fatal("expired response accepted")
	}
}

func TestBuilderBuildEnvironmentRequiresIndependentActionKey(t *testing.T) {
	t.Setenv("BUILDER_BUILD_ENABLED", "")
	if c, e := builderBuildFromEnv(nil); e != nil || c != nil {
		t.Fatal("default mounted a write client")
	}
	query, _ := newBuilderHubClient(builderTestConfig("http://127.0.0.1:8124"))
	values := map[string]string{"BUILDER_BUILD_ENABLED": "1", "CREDIT_DEMO_PASSWORD": "explicit-local-password", "BUILDER_BUILD_URL": query.cfg.Target, "BUILDER_BUILD_CHANNEL_TOKEN": strings.Repeat("t", 32), "BUILDER_BUILD_ISSUER": query.cfg.Issuer, "BUILDER_BUILD_KEY_ID": "independent-build-key", "BUILDER_BUILD_AUDIENCE": query.cfg.Audience, "BUILDER_BUILD_SCOPE": query.cfg.Scope, "BUILDER_BUILD_KEY_HEX": hex.EncodeToString(bytes.Repeat([]byte{42}, 32)), "BUILDER_BUILD_USERS_JSON": `{"alice":"controlled-operator"}`}
	encoded, _ := json.Marshal(map[string]json.RawMessage{"request": builderBuildTemplate()})
	values["BUILDER_BUILD_TEMPLATES_JSON"] = string(encoded)
	for k, v := range values {
		t.Setenv(k, v)
	}
	if c, e := builderBuildFromEnv(query); e != nil || c == nil {
		t.Fatal("explicit finite configuration rejected", e)
	}
	t.Setenv("BUILDER_BUILD_KEY_HEX", hex.EncodeToString(query.cfg.Key))
	if _, e := builderBuildFromEnv(query); e == nil {
		t.Fatal("query key obtained build capability")
	}
	t.Setenv("BUILDER_BUILD_KEY_HEX", values["BUILDER_BUILD_KEY_HEX"])
	t.Setenv("BUILDER_BUILD_SCOPE", "another-scope")
	if _, e := builderBuildFromEnv(query); e == nil {
		t.Fatal("query/write identity namespace silently diverged")
	}
}
