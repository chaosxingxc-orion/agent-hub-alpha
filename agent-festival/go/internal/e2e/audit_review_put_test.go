package e2e

// Failure-path acceptance for the T2 result store (external audit GC#28 /
// GH#51; the same defect is the reviewer's B0-RV-03).
//
// Contract pinned here: put returning nil means the new value is visible in
// memory AND readable from the durable file after a restart; put returning an
// error means NOTHING changed — no phantom committed result in memory, the
// previous durable bytes intact, no temp residue. The failure is injected
// deterministically by making the store's DIRECTORY read-only: the durable
// replace is a same-directory rename, which needs directory write permission
// (a read-only FILE is no longer a write barrier — see the put doc).
//
// Temp-file family in this store is ".t2-results.json<random>" (dot + base).
// Mutation anchors: restoring the old "assign to rs.data before persisting"
// order leaves a phantom result (get sees op-new) and fails the memory arms
// here and in the residue arm; replacing the temp+rename with a direct
// os.WriteFile fails the "must return an error" arm (the 0500 directory is
// writable-by-nobody, so a direct create/truncate errors) and the mode arm
// (a 0444 file rejects the direct write).

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func storeFileSum(t *testing.T, path string) [32]byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read store file: %v", err)
	}
	return sha256.Sum256(raw)
}

// TestT2Store_PutFailureChangesNothing pins the failure path end to end.
func TestT2Store_PutFailureChangesNothing(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions — the injection cannot fire (skips are not a reproduction)")
	}
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-old", storedResult{Receipt: "r-old", CmdDigest: "input-1", Facts: []string{"profit=100"}}); err != nil {
		t.Fatal(err)
	}
	before := storeFileSum(t, rs.path)

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	if err := rs.put("op-new", storedResult{Receipt: "r-new", CmdDigest: "input-2"}); err == nil {
		t.Fatal("put into a read-only directory must return an error")
	}

	// 1) no phantom: the failed op is not committed in memory
	if _, ok := rs.get("op-new"); ok {
		t.Error("failed put left a phantom committed result in memory")
	}
	// 2) the previously committed result is untouched in memory
	if r, ok := rs.get("op-old"); !ok || r.Receipt != "r-old" || r.CmdDigest != "input-1" {
		t.Errorf("old result changed by a failed put: %+v (ok=%v)", r, ok)
	}
	// 3) the durable file is byte-identical
	if after := storeFileSum(t, rs.path); after != before {
		t.Error("failed put changed the durable file")
	}
	// (Residue is asserted in TestT2Store_FailureAfterTempLeavesNoResidue: in
	// THIS injection CreateTemp fails before a temp file can exist, so a
	// residue check here would be vacuously green and is deliberately absent.)
	// 5) a restart sees the old value only
	reopened, err := newResultStore(dir)
	if err != nil {
		t.Fatalf("reopen after failed put: %v", err)
	}
	if _, ok := reopened.get("op-new"); ok {
		t.Error("restart exposes a result whose put failed")
	}
	if r, ok := reopened.get("op-old"); !ok || r.Receipt != "r-old" {
		t.Errorf("restart lost the previous result: %+v (ok=%v)", r, ok)
	}

	// 6) recovery: once the directory is writable again the put succeeds,
	// survives a restart, and does not drop the old entry.
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-new", storedResult{Receipt: "r-new", CmdDigest: "input-2"}); err != nil {
		t.Fatalf("put after recovery: %v", err)
	}
	reopened2, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reopened2.get("op-new"); !ok {
		t.Error("recovered put did not survive a restart")
	}
	if _, ok := reopened2.get("op-old"); !ok {
		t.Error("recovered put dropped the previously committed result")
	}
}

// TestT2Store_ConflictLeavesNothing pins the dedup guard's failure path too:
// a rejected re-put (different receipt) must not touch the durable file.
func TestT2Store_ConflictLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-c", storedResult{Receipt: "r-1", CmdDigest: "input-1"}); err != nil {
		t.Fatal(err)
	}
	before := storeFileSum(t, rs.path)
	err = rs.put("op-c", storedResult{Receipt: "r-2", CmdDigest: "input-2"})
	if err == nil {
		t.Fatal("a different receipt must be rejected")
	}
	// Diagnostic wording: the guard keys on the receipt, so the message must
	// name the receipt (not "digest" — that sends readers to CmdDigest).
	if !strings.Contains(err.Error(), "receipt") || strings.Contains(err.Error(), "digest") {
		t.Errorf("conflict message must name the receipt, got %q", err.Error())
	}
	if after := storeFileSum(t, rs.path); after != before {
		t.Error("a rejected put changed the durable file")
	}
}

// TestT2Store_DurableModePreserved pins that a whole-file replace keeps the
// operator's permission bits instead of widening them to the component
// default (the evidlog/adopt twin doctrine).
func TestT2Store_DurableModePreserved(t *testing.T) {
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-m", storedResult{Receipt: "r-m"}); err != nil {
		t.Fatal(err)
	}
	// First durable write must be owner-only 0600 (the twin doctrine), so a
	// permissive umask cannot widen the store behind the operator's back.
	if fi, err := os.Stat(rs.path); err != nil {
		t.Fatal(err)
	} else if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("first write mode = %04o, want 0600", got)
	}
	if err := os.Chmod(rs.path, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-m2", storedResult{Receipt: "r-m2"}); err != nil {
		t.Fatalf("put over a read-only file: %v", err)
	}
	info, err := os.Stat(rs.path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o444 {
		t.Errorf("replace widened the file mode to %04o, want 0444 preserved", got)
	}
}

// TestT2Store_RejectsUnrepresentableStore pins the open-time guard on BOTH
// faces: a top-level JSON that is not an object must fail loudly rather than
// silently becoming an empty store (a `null` file left a nil map that panicked
// on the next put), and an entry that decodes to a receipt-less record must be
// rejected too (it would otherwise lock its opID forever).
func TestT2Store_RejectsNonObjectTopLevel(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/t2-results.json"
	for _, body := range []string{"null", "[]", `"x"`, "42", `{"op": null}`, `{"op": {}}`} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := newResultStore(dir); err == nil {
			t.Errorf("top-level %s must be rejected at open, not silently accepted", body)
		}
	}
}

// TestT2Store_FailureAfterTempLeavesNoResidue pins the cleanup arm: the store
// path is turned into a non-empty directory, so the temp file IS created and
// written and only the final rename fails. The failed put must remove its temp
// file (no hidden residue) and still change nothing visible.
// Mutation anchor: dropping the deferred cleanup leaves a ".t2-results-*"
// entry behind and fails here (the read-only-directory injection cannot reach
// this arm — CreateTemp itself fails there).
func TestT2Store_FailureAfterTempLeavesNoResidue(t *testing.T) {
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-a", storedResult{Receipt: "r-a"}); err != nil {
		t.Fatal(err)
	}
	// Replace the store file with a NON-EMPTY directory of the same name: the
	// rename onto it fails (EISDIR/ENOTEMPTY), everything before it succeeds.
	if err := os.Remove(rs.path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(rs.path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rs.path, "occupied"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := rs.put("op-b", storedResult{Receipt: "r-b"}); err == nil {
		t.Fatal("rename onto a non-empty directory must fail")
	}
	if _, ok := rs.get("op-b"); ok {
		t.Error("failed put left a phantom committed result in memory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".t2-results") {
			t.Errorf("temp residue after a failed rename: %s", e.Name())
		}
	}
}

// TestT2Store_BarePathKeepsTempBesideStore pins the same-directory rule for a
// path without a directory component (filepath.Split returns "" there; without
// the guard the temp would land in os.TempDir and a cross-device rename would
// fail). Mutation anchor: dropping the dir=="" guard writes into TMPDIR — the
// probe points TMPDIR at a read-only directory, so the put fails instead.
func TestT2Store_BarePathKeepsTempBesideStore(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses the read-only TMPDIR injection — skips are not a reproduction")
	}
	work := t.TempDir()
	readonly := filepath.Join(work, "ro-tmp")
	if err := os.Mkdir(readonly, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", readonly)
	t.Chdir(work)

	rs := &resultStore{path: "t2-results.json", data: map[string]storedResult{}}
	if err := rs.put("op-bare", storedResult{Receipt: "r-bare"}); err != nil {
		t.Fatalf("put with a bare store path must keep its temp beside the store: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "t2-results.json")); err != nil {
		t.Fatalf("store file not written beside the bare path: %v", err)
	}
	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".t2-results") {
			t.Errorf("temp residue in the working directory: %s", e.Name())
		}
	}
}

// TestT2Store_WriteRejectsWhatOpenRejects pins write/read symmetry for the
// record gate: the open path rejects a receipt-less record, so put must refuse
// to create one — otherwise a single zero-valued put would make the WHOLE
// store unopenable after a restart.
// Mutation anchor: dropping the put-side validate call lets the empty receipt
// through and this test fails.
func TestT2Store_WriteRejectsWhatOpenRejects(t *testing.T) {
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-empty", storedResult{CmdDigest: "input-1"}); err == nil {
		t.Fatal("put with an empty receipt must be rejected (the open path would reject it)")
	}
	if _, ok := rs.get("op-empty"); ok {
		t.Error("rejected put must not leave a record in memory")
	}
	if _, err := os.Stat(rs.path); err == nil {
		if _, err := newResultStore(dir); err != nil {
			t.Errorf("store must stay openable after a rejected put: %v", err)
		}
	}
}
