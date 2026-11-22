package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// A human review is a distinct immutable fact. It never rewrites the checker
// artifact or says that the program proved a conclusion outside its coverage.
type CheckReview struct {
	CheckArtifact string `json:"check_artifact"`
	ObjectDigest  string `json:"object_digest"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	Principal     string `json:"principal"`
	Receipt       string `json:"receipt"`
	TaskID        string `json:"task_id,omitempty"`
	GraphRevision string `json:"graph_revision,omitempty"`
	GraphDigest   string `json:"graph_digest,omitempty"`
}

func reviewRequired(t *Task, outcome string) bool {
	return outcome == "HUMAN" || (outcome == "PASS" && t.Graph.HumanReviewRequired)
}

func reviewMatches(t *Task, ns *NodeState, digest string) bool {
	r := ns.Review
	if r == nil || r.Receipt == "" || r.Principal == "" || r.CheckArtifact != ns.Result.ArtifactID || r.ObjectDigest != digest {
		return false
	}
	if !t.Graph.HumanReviewRequired && r.TaskID == "" && r.GraphRevision == "" && r.GraphDigest == "" {
		return true // Legacy reviews were already stored within their task and checker node.
	}
	return r.TaskID == t.ID && r.GraphRevision == t.Revision && r.GraphDigest == t.GraphDigest
}

func effectiveCheck(t *Task, ns *NodeState) string {
	if ns == nil || ns.Result == nil || !ns.Result.Valid {
		return "UNKNOWN"
	}
	var check struct {
		Outcome      string `json:"outcome"`
		ObjectDigest string `json:"object_digest"`
	}
	if json.Unmarshal(ns.Result.Output, &check) != nil || check.ObjectDigest == "" {
		return "UNKNOWN"
	}
	if !reviewRequired(t, check.Outcome) {
		return check.Outcome
	}
	if reviewMatches(t, ns, check.ObjectDigest) {
		switch ns.Review.Decision {
		case "ACCEPT":
			return "PASS"
		case "REVISE":
			return "REVISE"
		case "REJECT":
			return "FAIL"
		}
	}
	return "HUMAN"
}

func (s *Store) ReviewCheck(ctx context.Context, id, node string, in CheckReview, principal string) (any, error) {
	if (in.Decision != "ACCEPT" && in.Decision != "REVISE" && in.Decision != "REJECT") || len(in.Reason) < 12 || len(in.Reason) > 4000 {
		return nil, fault("INVALID_REQUEST", "decision and substantive review reason required", 400)
	}
	return s.mutate(ctx, id, "HUMAN_REVIEW_RECORDED", func(tx pgx.Tx, t *Task) (any, error) {
		ns := t.Nodes[node]
		if ns == nil || ns.Result == nil || !ns.Result.Valid || ns.Result.ArtifactID != in.CheckArtifact {
			return nil, fault("INPUT_CONFLICT", "current valid checker artifact required", 409)
		}
		var kind, outcome, object string
		if err := tx.QueryRow(ctx, "SELECT checker_version,outcome,object_digest FROM checks WHERE scope=$1 AND task_id=$2 AND id=$3", s.Scope, id, in.CheckArtifact).Scan(&kind, &outcome, &object); err != nil {
			return nil, fault("NOT_FOUND", "qualified checker fact missing", 404)
		}
		if !reviewRequired(t, outcome) || object != in.ObjectDigest {
			return nil, fault("INPUT_CONFLICT", "only the exact report requiring human review can receive this decision", 409)
		}
		var checked struct {
			Failures []string `json:"failures"`
		}
		if json.Unmarshal(ns.Result.Output, &checked) != nil || len(checked.Failures) > 0 {
			return nil, fault("CHECK_REQUIRED", "program failures cannot be waived by this review", 409)
		}
		validReport := false
		for _, n := range t.Nodes {
			if n.Result != nil && n.Result.Valid && n.Result.ArtifactDigest == object {
				validReport = true
			}
		}
		if !validReport {
			return nil, fault("INVALIDATED", "reviewed report no longer valid", 409)
		}
		if ns.Review != nil {
			if reviewMatches(t, ns, object) && ns.Review.Decision == in.Decision && ns.Review.Reason == in.Reason && ns.Review.Principal == principal {
				return ns.Review, nil
			}
			return nil, fault("INPUT_CONFLICT", "review fact is immutable; new changed content requires a new check", 409)
		}
		in.Principal = principal
		in.Receipt = newID("human-review-")
		in.TaskID, in.GraphRevision, in.GraphDigest = t.ID, t.Revision, t.GraphDigest
		ns.Review = &in
		if in.Decision == "REJECT" {
			t.Status = "NEEDS_REVIEW"
			if _, err := tx.Exec(ctx, "UPDATE run_leases SET status='RELEASED' WHERE scope=$1 AND task_id=$2 AND domain='graph' AND status='ACTIVE'", s.Scope, t.ID); err != nil {
				return nil, err
			}
		} else {
			t.Status = "WAITING_INPUT"
		}
		return in, nil
	})
}

func (s *Store) qualifiedReportCheck(ctx context.Context, tx pgx.Tx, t *Task, digest string) (bool, error) {
	rows, err := tx.Query(ctx, "SELECT c.id FROM checks c JOIN committed_results r ON r.scope=c.scope AND r.task_id=c.task_id AND r.artifact_id=c.id WHERE c.scope=$1 AND c.task_id=$2 AND c.object_digest=$3 AND r.valid", s.Scope, t.ID, digest)
	if err != nil {
		return false, err
	}
	ids := map[string]bool{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return false, err
		}
		ids[id] = true
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return false, err
	}
	for _, ns := range t.Nodes {
		if ns.Result != nil && ids[ns.Result.ArtifactID] && effectiveCheck(t, ns) == "PASS" {
			return true, nil
		}
	}
	return false, nil
}

func (s *Server) humanRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/tasks/{id}/checks/{node}/review", s.wrap([]string{"user"}, func(w http.ResponseWriter, r *http.Request) (any, error) {
		principal, err := s.userWrite(r)
		if err != nil {
			return nil, err
		}
		var in CheckReview
		if err = readBody(r, &in); err != nil {
			return nil, err
		}
		if in.Principal != "" || in.Receipt != "" || in.TaskID != "" || in.GraphRevision != "" || in.GraphDigest != "" {
			return nil, fault("INVALID_REQUEST", "review identity and receipt are assigned by server", 400)
		}
		if _, err = s.Files.Read(in.CheckArtifact); err != nil {
			return nil, err
		}
		return s.Store.ReviewCheck(r.Context(), r.PathValue("id"), r.PathValue("node"), in, principal)
	}))
}
