package e2e

// Boundary-ownership pins for the result store (external audit GC#27 / GH#52).
//
// The store owns its records: whatever slices a caller hands to put or gets
// back from a read face may be edited freely without changing stored history
// — neither in memory nor in the durable file after a restart. Both directions
// are pinned because each was independently broken: mutating a returned slice
// changed history, and mutating a put argument did too.
//
// Mutation anchors: dropping the clone in put turns PutArgumentIsCopied red;
// dropping it in get / lookupByBusinessRequest turns
// exactly that face's sub-test red (each sub-test builds its own store, so the
// failure name is the culprit's name).

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func aliasStore(t *testing.T) *resultStore {
	t.Helper()
	rs, err := newResultStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

// TestT2Store_ReadFacesReturnCopies pins each read face in its own sub-test
// with a fresh store: a face that returns aliased slices must fail under ITS
// OWN name. (An earlier version exercised all three in one test body; a broken
// query face was then reported by the get assertion first — detection intact,
// attribution wrong. Per-face stores make the culprit unambiguous.)
//
// Mutation anchors: dropping the clone in get /
// lookupByBusinessRequest makes exactly the corresponding sub-test fail.
func TestT2Store_ReadFacesReturnCopies(t *testing.T) {
	faces := []struct {
		name string
		read func(*resultStore) (storedResult, bool)
	}{
		{"get", func(rs *resultStore) (storedResult, bool) { return rs.get("op-1") }},
		{"lookupByBusinessRequest", func(rs *resultStore) (storedResult, bool) {
			r, _, ok := rs.lookupByBusinessRequest("br-1")
			return r, ok
		}},
	}
	for _, f := range faces {
		t.Run(f.name, func(t *testing.T) {
			rs := aliasStore(t)
			if err := rs.put("op-1", storedResult{
				Receipt:       "r-1",
				BusinessReqID: "br-1",
				EvaluationRev: "er-1",
				Facts:         []string{"profit=100"},
				References:    []string{"pg-12"},
			}); err != nil {
				t.Fatal(err)
			}

			r, ok := f.read(rs)
			if !ok {
				t.Fatalf("%s did not return the record", f.name)
			}
			r.Facts[0] = "TAMPERED-FACT"
			r.References[0] = "TAMPERED-REF"

			// the same face, read again: memory must be pristine
			if again, _ := f.read(rs); again.Facts[0] != "profit=100" || again.References[0] != "pg-12" {
				t.Fatalf("%s edits reached stored memory: %v / %v", f.name, again.Facts, again.References)
			}
			// a read through another face must also see pristine (no masking)
			if viaGet, _ := rs.get("op-1"); viaGet.Facts[0] != "profit=100" || viaGet.References[0] != "pg-12" {
				t.Fatalf("%s edits visible through get: %v / %v", f.name, viaGet.Facts, viaGet.References)
			}
			// and nothing may reach the durable file
			if err := rs.put("op-2", storedResult{Receipt: "r-2"}); err != nil {
				t.Fatal(err)
			}
			reopened, err := newResultStore(filepath.Dir(rs.path))
			if err != nil {
				t.Fatal(err)
			}
			if r2, _ := reopened.get("op-1"); r2.Facts[0] != "profit=100" || r2.References[0] != "pg-12" {
				t.Fatalf("%s edits were persisted: %v / %v", f.name, r2.Facts, r2.References)
			}
		})
	}
}

// TestT2Store_PutArgumentIsCopied pins the write direction: the store keeps
// its own copy, so editing the argument after put changes nothing.
func TestT2Store_PutArgumentIsCopied(t *testing.T) {
	rs := aliasStore(t)
	facts := []string{"profit=100"}
	refs := []string{"pg-12"}
	if err := rs.put("op-1", storedResult{Receipt: "r-1", Facts: facts, References: refs}); err != nil {
		t.Fatal(err)
	}
	facts[0] = "TAMPERED-INPUT"
	refs[0] = "TAMPERED-INPUT-REF"

	if r, _ := rs.get("op-1"); r.Facts[0] != "profit=100" || r.References[0] != "pg-12" {
		t.Fatalf("put argument edits reached stored memory: %v / %v", r.Facts, r.References)
	}
	if err := rs.put("op-2", storedResult{Receipt: "r-2"}); err != nil {
		t.Fatal(err)
	}
	reopened, err := newResultStore(filepath.Dir(rs.path))
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := reopened.get("op-1"); r.Facts[0] != "profit=100" {
		t.Fatalf("put argument edit was persisted: %v", r.Facts)
	}
}

// TestT2Store_CloneKeepsNilAndEmptyDistinct pins the durable-form distinction:
// nil marshals as null and an empty slice as [], and both survive a restart
// unchanged — so the clone must not normalise one into the other.
func TestT2Store_CloneKeepsNilAndEmptyDistinct(t *testing.T) {
	dir := t.TempDir()
	rs, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-nil", storedResult{Receipt: "r-nil"}); err != nil {
		t.Fatal(err)
	}
	if err := rs.put("op-empty", storedResult{Receipt: "r-empty", Facts: []string{}, References: []string{}}); err != nil {
		t.Fatal(err)
	}
	rNil, _ := rs.get("op-nil")
	rEmpty, _ := rs.get("op-empty")
	if rNil.Facts != nil || rNil.References != nil {
		t.Errorf("nil slices became non-nil: Facts=%#v References=%#v", rNil.Facts, rNil.References)
	}
	if rEmpty.Facts == nil || rEmpty.References == nil {
		t.Errorf("empty slices became nil: Facts=%#v References=%#v", rEmpty.Facts, rEmpty.References)
	}

	raw, err := os.ReadFile(rs.path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"facts": null`) || !strings.Contains(string(raw), `"facts": []`) ||
		!strings.Contains(string(raw), `"references": null`) || !strings.Contains(string(raw), `"references": []`) {
		t.Errorf("durable form lost the null/[] distinction:\n%s", raw)
	}

	reopened, err := newResultStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	rn, _ := reopened.get("op-nil")
	re, _ := reopened.get("op-empty")
	if rn.Facts != nil || rn.References != nil {
		t.Errorf("restart turned nil into Facts=%#v References=%#v", rn.Facts, rn.References)
	}
	if re.Facts == nil || re.References == nil {
		t.Errorf("restart turned [] into Facts=%#v References=%#v", re.Facts, re.References)
	}
}

// TestStoredResultCloneCoversAllFields is the structural pin: cloneStored only
// handles string and []string fields. A future map/pointer/slice-of-slice
// field would alias silently — this test fails until the clone is extended.
// Mutation anchor: adding `FutureMap map[string]string` to storedResult makes
// this red (verified during the review round); removing it restores green.
func TestStoredResultCloneCoversAllFields(t *testing.T) {
	// The pin is an exact allow-list, not a kind test: ANY new slice field —
	// including a second []string — must fail here until cloneStored is
	// extended, because a kind-only test would let the most likely addition
	// (`Notes []string`) pass while still aliasing.
	cloneHandled := map[string]bool{"Facts": true, "References": true}
	typ := reflect.TypeOf(storedResult{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch f.Type.Kind() {
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
			// value types: no aliasing possible
		case reflect.Slice:
			if f.Type.Elem().Kind() != reflect.String {
				t.Errorf("field %s is []%s — cloneStored copies []string only; a nested/composite element shares its own backing arrays, so define an explicit copy strategy before landing it", f.Name, f.Type.Elem().Kind())
				continue
			}
			if !cloneHandled[f.Name] {
				t.Errorf("field %s is a []string the store must own — copy it in cloneStored and register it here", f.Name)
			}
		case reflect.Array, reflect.Map, reflect.Pointer, reflect.Interface:
			t.Errorf("field %s (kind %s) needs an explicit copy strategy: the clone handles strings and []string only", f.Name, f.Type.Kind())
		default:
			t.Errorf("field %s (kind %s) is not covered by cloneStored — review it (a composite field may hide slices) and extend this pin", f.Name, f.Type.Kind())
		}
	}

	// An allow-list can be defeated by adding a name without adding the copy:
	// prove each listed field is ACTUALLY cloned. The probe is filled by
	// reflection (one-element slice per listed field), so a newly allow-listed
	// field needs no test edit — and if cloneStored forgot it, the backing
	// arrays stay equal and this fails.
	probeV := reflect.New(typ).Elem()
	for name := range cloneHandled {
		f := probeV.FieldByName(name)
		if !f.IsValid() || f.Kind() != reflect.Slice {
			t.Fatalf("allow-listed field %s is missing or not a slice — the pin needs updating", name)
		}
		f.Set(reflect.MakeSlice(f.Type(), 1, 1))
	}
	probe, ok := probeV.Interface().(storedResult)
	if !ok {
		t.Fatal("reflection probe is not a storedResult")
	}
	cv := reflect.ValueOf(cloneStored(probe))
	for name := range cloneHandled {
		if probeV.FieldByName(name).Pointer() == cv.FieldByName(name).Pointer() {
			t.Errorf("field %s is allow-listed but cloneStored does not copy it (same backing array)", name)
		}
	}
}
