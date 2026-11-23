package main

import "encoding/json"

// Private finite scene adapters. These do not extend CapabilityCard,
// SelectionQuery, the SDK or the A2A message envelope.
type H03FieldSpec struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Unit   string `json:"unit"`
	Period string `json:"period"`
	Basis  string `json:"basis"`
}

type H03Field struct {
	H03FieldSpec
	Value string `json:"value"`
}

type H03NodeContract struct {
	SchemaVersion string         `json:"schema_version"`
	InputFields   []H03FieldSpec `json:"input_fields"`
	OutputFields  []H03FieldSpec `json:"output_fields"`
}

// Paths come only from deployment configuration, never from a planner query.
type H03ProfileMaterial struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
	Path   string `json:"path"`
	raw    []byte
}

type H03Profile struct {
	SchemaVersion      string                     `json:"schema_version"`
	Version            string                     `json:"version"`
	Package            PackageRef                 `json:"package"`
	FieldSemantics     []H03FieldSpec             `json:"field_semantics"`
	CandidateContracts map[string]H03NodeContract `json:"candidate_contracts"`
	Materials          []H03ProfileMaterial       `json:"materials"`
}

type H03MaterialDocument struct {
	SchemaVersion string     `json:"schema_version"`
	Fields        []H03Field `json:"fields"`
}

type H03InputBinding struct {
	ToNode      string `json:"to_node"`
	ToField     string `json:"to_field"`
	SourceKind  string `json:"source_kind"`
	SourceID    string `json:"source_id"`
	SourceField string `json:"source_field"`
}

type H03ExecutionBinding struct {
	// Local adopted execution config; never a node/A2A wire field.
	BusTenant        string                     `json:"bus_tenant"`
	SchemaVersion    string                     `json:"schema_version"`
	Profile          string                     `json:"profile"`
	ScenarioSnapshot json.RawMessage            `json:"scenario_snapshot"`
	Materials        []MaterialRef              `json:"materials"`
	NodeContracts    map[string]H03NodeContract `json:"node_contracts"`
	InputBindings    []H03InputBinding          `json:"input_bindings"`
}

type H03Dependency struct {
	Kind            string `json:"kind"`
	ID              string `json:"id"`
	SHA256          string `json:"sha256,omitempty"`
	NodeExecutionID string `json:"node_execution_id,omitempty"`
	ArtifactDigest  string `json:"artifact_digest,omitempty"`
}

type H03BoundInput struct {
	SchemaVersion          string          `json:"schema_version"`
	ExecutionBindingSHA256 string          `json:"execution_binding_sha256"`
	NodeContractSHA256     string          `json:"node_contract_sha256"`
	PlanHash               string          `json:"plan_hash"`
	RunID                  string          `json:"run_id"`
	NodeExecutionID        string          `json:"node_execution_id"`
	NodePath               string          `json:"node_path"`
	Attempt                int             `json:"attempt"`
	Fields                 []H03Field      `json:"fields"`
	Dependencies           []H03Dependency `json:"dependencies"`
}

type H03NodeOutput struct {
	SchemaVersion          string     `json:"schema_version"`
	ExecutionBindingSHA256 string     `json:"execution_binding_sha256"`
	NodeContractSHA256     string     `json:"node_contract_sha256"`
	InputDigest            string     `json:"input_digest"`
	Fields                 []H03Field `json:"fields"`
}
