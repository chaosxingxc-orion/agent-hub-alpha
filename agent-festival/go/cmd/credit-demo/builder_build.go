package main

import (
	"bytes"
	"encoding/json"
	"sort"
)

// Reuses the original bounded network-intent association. It never records
// Task execution state, grants, budget, approvals, or an inferred outcome.
type builderActionIntent struct {
	CommandID string          `json:"command_id"`
	User      string          `json:"user"`
	Subject   string          `json:"subject"`
	Scope     string          `json:"scope"`
	Alias     string          `json:"alias"`
	RequestID string          `json:"request_id"`
	Digest    string          `json:"digest"`
	Material  json.RawMessage `json:"material"`
}

func builderBuildIdentity(user, subject, scope, alias, digest string) (string, string) {
	raw, _ := builderBuildCanonical([]string{"build", alias, digest})
	key := builderBuildSHA(raw)
	identity, _ := builderBuildCanonical([]string{user, subject, scope, key})
	return "h04." + builderBuildSHA(identity)[:48], key
}

func (s *Service) builderPrepare(c *builderBuildClient, user, subject, alias string) (builderActionIntent, bool, error) {
	raw, ok := c.builds[alias]
	if !ok {
		return builderActionIntent{}, false, errBuilderHub
	}
	m, _, digest, e := builderBuildMaterial(raw, false)
	if e != nil {
		return builderActionIntent{}, false, e
	}
	id, key := builderBuildIdentity(user, subject, c.cfg.Scope, alias, digest)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.storageErr != nil {
		return builderActionIntent{}, false, s.storageErr
	}
	for _, old := range s.st.BuilderActions {
		if old.User == user && old.Alias == alias {
			if old.Subject != subject || old.Scope != c.cfg.Scope || old.CommandID != id || old.RequestID != key || old.Digest != digest {
				return builderActionIntent{}, false, errBuilderHub
			}
			m["command_id"] = id
			body, _ := builderBuildCanonical(m)
			if !builderBuildStoredMaterial(old.Material, body) {
				return builderActionIntent{}, false, errBuilderHub
			}
			old.Material = bytes.Clone(old.Material)
			return old, false, nil
		}
	}
	if _, exists := s.st.BuilderActions[id]; exists || len(s.st.BuilderActions) >= 256 {
		return builderActionIntent{}, false, errBuilderHub
	}
	if s.st.BuilderActions == nil {
		s.st.BuilderActions = map[string]builderActionIntent{}
	}
	m["command_id"] = id
	body, e := builderBuildCanonical(m)
	if e != nil {
		return builderActionIntent{}, false, e
	}
	row := builderActionIntent{id, user, subject, c.cfg.Scope, alias, key, digest, body}
	s.st.BuilderActions[id] = row
	if e = s.save(); e != nil {
		return builderActionIntent{}, false, e
	}
	row.Material = bytes.Clone(row.Material)
	return row, true, nil
}

func (s *Service) builderIntent(c *builderBuildClient, user, subject, id string) (builderActionIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.storageErr != nil {
		return builderActionIntent{}, s.storageErr
	}
	row, ok := s.st.BuilderActions[id]
	if !ok || row.CommandID != id || row.User != user || row.Subject != subject || row.Scope != c.cfg.Scope {
		return builderActionIntent{}, errBuilderHub
	}
	raw, ok := c.builds[row.Alias]
	if !ok {
		return builderActionIntent{}, errBuilderHub
	}
	m, _, digest, e := builderBuildMaterial(raw, false)
	if e != nil {
		return builderActionIntent{}, e
	}
	expected, key := builderBuildIdentity(user, subject, c.cfg.Scope, row.Alias, digest)
	m["command_id"] = id
	body, _ := builderBuildCanonical(m)
	if expected != id || row.RequestID != key || row.Digest != digest || !builderBuildStoredMaterial(row.Material, body) {
		return builderActionIntent{}, errBuilderHub
	}
	row.Material = bytes.Clone(row.Material)
	return row, nil
}

// The original writeJSON indents json.RawMessage values. Reconcile their exact
// canonical command content on reads; never mutate state or regain a first send.
func builderBuildStoredMaterial(stored, expected []byte) bool {
	m, e := builderJSON(stored)
	if e != nil {
		return false
	}
	canonical, e := builderBuildCanonical(m)
	return e == nil && bytes.Equal(canonical, expected)
}

// Called with the original state lock held. Turning a feature off must not
// turn its persisted subject/source/capability records into anonymous debug data.
func (s *Service) privateIntegrationStateLocked() bool {
	if s.cfg.BuilderBuild != nil || s.cfg.H03Profile != nil {
		return true
	}
	if s.st == nil {
		return false
	}
	if len(s.st.BuilderActions) > 0 {
		return true
	}
	for _, versions := range s.st.Plans {
		for _, p := range versions {
			if p != nil && p.ExecutionBinding != nil {
				return true
			}
		}
	}
	return false
}

func (s *Service) builderIntentList(c *builderBuildClient, user, subject string) ([]builderActionIntent, error) {
	s.mu.Lock()
	if s.storageErr != nil {
		s.mu.Unlock()
		return nil, errStorage
	}
	ids := []string{}
	for id, row := range s.st.BuilderActions {
		if row.User == user && row.Subject == subject && row.Scope == c.cfg.Scope {
			ids = append(ids, id)
		}
	}
	s.mu.Unlock()
	sort.Strings(ids)
	rows := []builderActionIntent{}
	for _, id := range ids {
		row, e := s.builderIntent(c, user, subject, id)
		if e != nil {
			return nil, e
		}
		if c.allowed(user, subject, row.Alias) {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
