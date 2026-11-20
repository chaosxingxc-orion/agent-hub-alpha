package validate

// White-box pins for the per-call root parameter (external audit GC#26).
//
// After the fix, the root document that same-document $defs refs resolve
// against is a parameter of the recursive walk, not shared state. These pins
// make "the parameter is actually consumed" a mechanical fact:
//   - the same schema under two different roots must produce different
//     verdicts (wrong root = RED, so a future edit that ignores the parameter
//     and always uses the schema cannot stay green);
//   - a nil root must stay fail-closed (unresolved $ref is an error);
//   - the public entry keeps its meaning (root = schema).

import (
	"strings"
	"testing"
)

func rootScopeSchema() Schema {
	return Schema{"$ref": "#/$defs/value"}
}

func rootScopeDocRoot(typ string) Schema {
	return Schema{"$defs": map[string]any{
		"value": map[string]any{"type": typ},
	}}
}

// TestRootParameterIsConsumed pins that the root argument decides $defs
// resolution: string-root accepts, integer-root rejects the same document.
// Mutation anchor: wiring the walk to always use `schema` as its own root
// makes the second arm green and this test fails.
func TestRootParameterIsConsumed(t *testing.T) {
	schema := rootScopeSchema()

	var errsA []string
	schemaErrors("hello", schema, rootScopeDocRoot("string"), "<root>", &errsA)
	if len(errsA) != 0 {
		t.Fatalf("string root should accept the document, got %v", errsA)
	}

	var errsB []string
	schemaErrors("hello", schema, rootScopeDocRoot("integer"), "<root>", &errsB)
	if len(errsB) == 0 {
		t.Fatal("integer root must reject a string document — the root parameter is not being consumed")
	}
	if !strings.Contains(strings.Join(errsB, " "), "not of type") {
		t.Fatalf("expected a type error from the integer root, got %v", errsB)
	}
}

// TestNilRootFailsClosed keeps the original fail-closed behaviour: without a
// root, a same-document $ref is an error, never a silent pass.
func TestNilRootFailsClosed(t *testing.T) {
	var errs []string
	schemaErrors("hello", rootScopeSchema(), nil, "<root>", &errs)
	if len(errs) == 0 || !strings.Contains(strings.Join(errs, " "), "unresolved $ref") {
		t.Fatalf("nil root must fail closed with an unresolved $ref error, got %v", errs)
	}
}

// TestPublicEntryRootIsSchema pins the documented meaning of the exported
// entry: $defs resolve against the schema being validated.
func TestPublicEntryRootIsSchema(t *testing.T) {
	root := rootScopeDocRoot("integer")
	var errs []string
	SchemaErrorsDoc("hello", Schema{"$ref": "unresolvable-for-this-pin"}, "<root>", &errs)
	if len(errs) == 0 {
		t.Fatal("unsupported $ref must fail closed")
	}
	var ok []string
	SchemaErrorsDoc("hello", root, "<root>", &ok)
	if len(ok) != 0 {
		t.Fatalf("a document validated against its own root must accept, got %v", ok)
	}

	// Positive $defs arm: the wrapper must resolve against the schema itself.
	// Mutation anchor: defaulting the wrapper's root to nil turns this arm red
	// (unresolved $ref) — which is exactly the contract this entry documents.
	withDefs := Schema{
		"$defs": map[string]any{"value": map[string]any{"type": "string"}},
		"$ref":  "#/$defs/value",
	}
	var selfRooted []string
	SchemaErrorsDoc("hello", withDefs, "<root>", &selfRooted)
	if len(selfRooted) != 0 {
		t.Fatalf("SchemaErrorsDoc must resolve $defs against the schema itself, got %v", selfRooted)
	}

	// Same contract for the other exported entry (same family as the arm
	// above): both wrappers default the root to the schema under validation.
	var viaBare []string
	SchemaErrors("hello", withDefs, "<root>", &viaBare)
	if len(viaBare) != 0 {
		t.Fatalf("SchemaErrors must resolve $defs against the schema itself, got %v", viaBare)
	}
}
