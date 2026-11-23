package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func builderBuildTestService(t *testing.T, dir string) *Service {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	s := &Service{cfg: Config{DataDir: dir, Origins: []string{"http://127.0.0.1:8123"}}, st: &State{Plans: map[string][]*Plan{}, Runs: map[string]*Run{}}, writeState: writeJSON}
	s.lastSaved, _ = json.Marshal(s.st)
	return s
}

func TestBuilderBuildPrepareConcurrentAndActualRestart(t *testing.T) {
	c, _ := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	var fresh atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, first, e := s.builderPrepare(c, "alice", "alice", "request")
			if e != nil {
				t.Error(e)
			}
			if first {
				fresh.Add(1)
			}
		}()
	}
	wg.Wait()
	if fresh.Load() != 1 || len(s.st.BuilderActions) != 1 {
		t.Fatal("more than one first-send opportunity")
	}
	rev := s.st.Rev
	// Exercise the original Service loader, rather than reconstructing associations.
	reopened, e := OpenService(Config{DataDir: s.cfg.DataDir, SpecPath: filepath.Join("..", "..", "..", "schema", "examples", "judgment", "valid", "judgment-spec-credit.json"), Model: "fixture-model"})
	if e != nil {
		t.Fatal(e)
	}
	row, first, e := reopened.builderPrepare(c, "alice", "alice", "request")
	if e != nil || first || reopened.st.Rev != rev {
		t.Fatal("restart recreated send opportunity", e)
	}
	if _, e = reopened.builderIntent(c, "other", "alice", row.CommandID); e == nil {
		t.Fatal("cross user read")
	}
	before, _ := os.ReadFile(filepath.Join(s.cfg.DataDir, "state.json"))
	_, _, _ = reopened.builderPrepare(c, "alice", "alice", "request")
	after, _ := os.ReadFile(filepath.Join(s.cfg.DataDir, "state.json"))
	if string(before) != string(after) {
		t.Fatal("read/reuse wrote state")
	}
}

func TestBuilderBuildSaveFailureFreezesAndListsUnavailable(t *testing.T) {
	c, _ := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	s.writeState = func(string, any) error { return errors.New("disk failure") }
	if _, first, e := s.builderPrepare(c, "alice", "alice", "request"); e == nil || first {
		t.Fatal("failed save allowed send")
	}
	if _, e := s.builderIntentList(c, "alice", "alice"); e == nil {
		t.Fatal("storage failure became an empty history")
	}
	if _, first, e := s.builderPrepare(c, "alice", "alice", "request"); e == nil || first {
		t.Fatal("failed memory state unfroze")
	}
}

func TestBuilderBuildDriftConflictAndNamedNewRequest(t *testing.T) {
	c, _ := newBuilderBuildClient(builderTestConfig("http://127.0.0.1:1"), map[string]json.RawMessage{"request": builderBuildTemplate()})
	s := builderBuildTestService(t, "")
	old, _, _ := s.builderPrepare(c, "alice", "alice", "request")
	if _, first, e := s.builderPrepare(c, "alice", "changed-subject", "request"); e == nil || first {
		t.Fatal("identity remapping regained a send opportunity")
	}
	m, _, _, _ := builderBuildMaterial(c.builds["request"], false)
	m["input_ref"].(map[string]any)["id"] = "different-source"
	changed, _ := builderBuildCanonical(m)
	c.builds["request"] = changed
	if _, first, e := s.builderPrepare(c, "alice", "alice", "request"); e == nil || first {
		t.Fatal("same alias drift accepted")
	}
	if _, e := s.builderIntent(c, "alice", "alice", old.CommandID); e == nil {
		t.Fatal("drifted source disclosed")
	}
	// Only trusted server registration supplies a separately named source request.
	c.builds["request-version-2"] = changed
	next, first, e := s.builderPrepare(c, "alice", "alice", "request-version-2")
	if e != nil || !first || next.CommandID == old.CommandID {
		t.Fatal("legitimate new request swallowed")
	}
	if _, ok := s.st.BuilderActions[old.CommandID]; !ok {
		t.Fatal("old uncertainty deleted")
	}
}
