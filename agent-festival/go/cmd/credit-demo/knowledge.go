package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Cross-task knowledge reuse (A10 / F-02C / ALK-08, #16). A node whose
// capability card is marked reusable and whose plan declares its sources
// (the scenario package's materials, pinned by SHA-256) publishes its result
// as a knowledge item when it commits. A later task's node looks the item up
// before it runs: only an item from another task with the same capability
// version and the same kappa_in (instruction, sub-task, task input, sources
// by digest, upstream outputs, model) is reused, and only while it is not
// revoked. Lookup and the step record happen in one locked transition, so a
// revocation either lands before the reuse (miss, with the reason) or after
// it (the reusing run shows that its source was revoked). Every lookup keeps
// its outcome and reason on the step.
//
// The model in the key is the one the run is bound to (AR-04), not the
// service's current default. A hit copies the item's provenance onto the
// step, so the persisted step alone says which result, capability, package,
// sources, model and output it took (#55 analog).

type KOrigin struct {
	Run  string `json:"run"`
	Plan string `json:"plan"`
	Step string `json:"step"`
}

type Revocation struct {
	By     string `json:"by"`
	Reason string `json:"reason"`
	At     string `json:"at"`
}

type KItem struct {
	ID           string        `json:"id"`
	Capability   string        `json:"capability"` // id@version
	Package      string        `json:"package"`    // id@version the card came from
	KappaIn      string        `json:"kappa_in"`
	Sources      []MaterialRef `json:"sources"`
	Output       string        `json:"output"`
	Model        string        `json:"model"`         // the model the producing run was bound to
	OutputDigest string        `json:"output_digest"` // sha of Output when published
	Origin       KOrigin       `json:"origin"`
	CommittedAt  string        `json:"committed_at"`
	Revoked      *Revocation   `json:"revoked,omitempty"`
}

// Lookup is what a node's knowledge lookup found, kept on the step. On a hit
// it also carries the reused item's provenance, so the step stays traceable
// without the item.
type Lookup struct {
	Outcome      string        `json:"outcome"` // hit miss
	Reason       string        `json:"reason"`
	Item         string        `json:"item,omitempty"` // the item reused, or the one the reason is about
	Source       string        `json:"source,omitempty"`
	KappaIn      string        `json:"kappa_in"`
	At           string        `json:"at"`
	Capability   string        `json:"capability,omitempty"`
	Package      string        `json:"package,omitempty"`
	Sources      []MaterialRef `json:"sources,omitempty"`
	Model        string        `json:"model,omitempty"`
	OutputDigest string        `json:"output_digest,omitempty"`
}

// reuseKey is what a node's result depends on; ok is false when the node's
// capability is not reusable or the plan declares no sources.
// Must be called with s.mu held.
func (s *Service) reuseKey(r *Run, p *Plan, step string) (kappa, model string, card *CapabilityCard, pkg *ScenarioPackage, sources []MaterialRef, ok bool) {
	a := p.agent(step)
	if pkg = p.scenario(); a == nil || pkg == nil || p.Selection == nil {
		return
	}
	if card = pkg.card(a.Expert); card == nil || !card.Reusable {
		return
	}
	rec := s.st.Selections[p.Selection.ID]
	if rec == nil {
		return
	}
	for _, id := range rec.Query.Materials {
		for _, m := range pkg.Materials {
			if m.ID == id {
				sources = append(sources, MaterialRef{ID: m.ID, SHA256: m.SHA256})
			}
		}
	}
	if len(sources) == 0 {
		return
	}
	upstream := map[string]any{}
	for _, st := range p.Steps {
		if st.ID != step {
			continue
		}
		for _, d := range st.DependsOn {
			if dep := r.Steps[d]; dep != nil {
				upstream[d] = sha(dep.Status, stripStepTag(dep.Output))
			}
		}
	}
	model = s.runModel(r)
	kappa, err := canonicalHash(map[string]any{
		"capability": card.ID + "@" + card.Version, "instruction": card.Instruction, "role": a.Role, "task": a.Task,
		"input": s.jobInput(r, p), "sources": sources, "upstream": upstream, "model": model,
	})
	return kappa, model, card, pkg, sources, err == nil
}

func sameSources(a, b []MaterialRef, digests bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || (digests && a[i].SHA256 != b[i].SHA256) {
			return false
		}
	}
	return true
}

// lookupKnowledge finds a reusable item from another run, or says why there
// is none. Must be called with s.mu held.
func (s *Service) lookupKnowledge(runID string, card *CapabilityCard, kappa, model string, sources []MaterialRef) Lookup {
	capRef := card.ID + "@" + card.Version
	lk := Lookup{Outcome: "miss", Reason: "没有其他任务消费过这些来源", KappaIn: kappa, At: now()}
	rank := 0 // how informative the current miss reason is
	miss := func(r int, item *KItem, reason string) {
		if r > rank {
			rank, lk.Reason, lk.Item, lk.Source = r, reason, item.ID, item.Origin.Run+"/"+item.Origin.Step
		}
	}
	for _, item := range s.knowledgeNewestFirst() {
		if item.Origin.Run == runID {
			continue
		}
		sameCap := item.Capability == capRef
		switch {
		case sameCap && item.KappaIn == kappa && item.OutputDigest == "":
			miss(6, item, fmt.Sprintf("已有结果 %s 发布时未记录模型与输出摘要，无法确认与本任务绑定的模型一致，不复用", item.ID))
		case sameCap && item.KappaIn == kappa && (item.Model != model || item.OutputDigest != sha(item.Output)):
			miss(6, item, fmt.Sprintf("已有结果 %s 的模型或输出摘要与记录不符，不复用", item.ID))
		case sameCap && item.KappaIn == kappa && item.Revoked == nil:
			return Lookup{Outcome: "hit", Reason: "能力版本、模型、输入与来源版本完全一致，可复用", Item: item.ID,
				Source: item.Origin.Run + "/" + item.Origin.Step, KappaIn: kappa, At: lk.At,
				Capability: item.Capability, Package: item.Package, Sources: append([]MaterialRef(nil), item.Sources...),
				Model: item.Model, OutputDigest: item.OutputDigest}
		case sameCap && item.KappaIn == kappa:
			miss(5, item, fmt.Sprintf("原结果 %s 已撤回（%s，%s），不复用", item.ID, item.Revoked.Reason, item.Revoked.By))
		case sameCap && sameSources(item.Sources, sources, false) && !sameSources(item.Sources, sources, true):
			miss(4, item, fmt.Sprintf("已有结果 %s 基于同一材料的其他版本（%s），版本不同不复用", item.ID, shortDigest(item.Sources)))
		case sameCap && sameSources(item.Sources, sources, true) && item.OutputDigest != "" && item.Model != model:
			miss(4, item, fmt.Sprintf("已有结果 %s 由模型 %s 产生，本任务绑定模型 %s，不复用", item.ID, item.Model, model))
		case sameCap && sameSources(item.Sources, sources, true):
			miss(3, item, fmt.Sprintf("已有结果 %s 来源相同，但指令、子任务、任务输入或上游结果不同，不复用", item.ID))
		case strings.HasPrefix(item.Capability, card.ID+"@") && sameSources(item.Sources, sources, true):
			miss(2, item, fmt.Sprintf("已有结果 %s 来自能力的其他版本（%s），不复用", item.ID, item.Capability))
		}
	}
	return lk
}

func shortDigest(ms []MaterialRef) string {
	var out []string
	for _, m := range ms {
		out = append(out, m.ID+"@"+clip(m.SHA256, 8))
	}
	return strings.Join(out, ", ")
}

func (s *Service) knowledgeNewestFirst() []*KItem {
	out := make([]*KItem, 0, len(s.st.Knowledge))
	for _, item := range s.st.Knowledge {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.Atoi(strings.TrimPrefix(out[i].ID, "K"))
		b, _ := strconv.Atoi(strings.TrimPrefix(out[j].ID, "K"))
		return a > b
	})
	return out
}

// reuseStep is StepGet's knowledge lookup for a step that has not run in
// this task: on a hit the step is recorded as done with the item's result,
// in the same transition. Must be called with s.mu held.
func (s *Service) reuseStep(r *Run, p *Plan, st *StepRec) (bool, error) {
	kappa, model, card, _, sources, ok := s.reuseKey(r, p, st.ID)
	if !ok {
		return false, nil
	}
	lk := s.lookupKnowledge(r.ID, card, kappa, model, sources)
	st.Lookup = &lk
	if lk.Outcome != "hit" {
		return false, nil
	}
	item := s.st.Knowledge[lk.Item]
	out := item.Output
	if tag := "[step:" + item.Origin.Step + "]"; strings.HasPrefix(out, tag) {
		out = "[step:" + st.ID + "]" + strings.TrimPrefix(out, tag)
	}
	st.Status, st.Output, st.StartedAt, st.FinishedAt, st.Epoch = "done", out, lk.At, lk.At, r.Epoch
	st.ReusedFrom, st.Activity = item.ID, ""
	return true, nil
}

// publishStep makes a committed result of a reusable node available to later
// tasks. Must be called with s.mu held, before the step's save.
func (s *Service) publishStep(r *Run, p *Plan, st *StepRec) {
	kappa, model, card, pkg, sources, ok := s.reuseKey(r, p, st.ID)
	if !ok || st.ReusedFrom != "" {
		return
	}
	if s.st.Knowledge == nil {
		s.st.Knowledge = map[string]*KItem{}
	}
	s.st.NextKnowledge++
	item := &KItem{ID: fmt.Sprintf("K%d", s.st.NextKnowledge), Capability: card.ID + "@" + card.Version, Package: pkg.Ref(),
		KappaIn: kappa, Sources: sources, Output: st.Output, Model: model, OutputDigest: sha(st.Output),
		Origin: KOrigin{Run: r.ID, Plan: p.Key(), Step: st.ID}, CommittedAt: now()}
	s.st.Knowledge[item.ID] = item
	st.Published = item.ID
}

// RevokeKnowledge withdraws an item: later lookups miss with the reason, and
// runs that already reused it show that their source was revoked.
func (s *Service) RevokeKnowledge(id, by, reason string) (*KItem, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("a revocation needs a reason")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.storageErr != nil {
		return nil, s.storageErr
	}
	item := s.st.Knowledge[id]
	if item == nil {
		return nil, fmt.Errorf("unknown knowledge item %q", id)
	}
	if item.Revoked != nil {
		return item, nil
	}
	item.Revoked = &Revocation{By: by, Reason: reason, At: now()}
	if err := s.save(); err != nil {
		return nil, err
	}
	return item, nil
}

// KnowledgeView is one item as the workbench lists it.
type KnowledgeView struct {
	*KItem
	Output   string   `json:"output"` // clipped
	ReusedBy []string `json:"reused_by,omitempty"`
}

func (s *Service) Knowledge() []KnowledgeView {
	s.mu.Lock()
	defer s.mu.Unlock()
	reusedBy := map[string][]string{}
	for _, r := range s.st.Runs {
		for _, st := range r.Steps {
			if st.ReusedFrom != "" {
				reusedBy[st.ReusedFrom] = append(reusedBy[st.ReusedFrom], r.ID+"/"+st.ID)
			}
		}
	}
	out := []KnowledgeView{}
	for _, item := range s.knowledgeNewestFirst() {
		c := *item
		sort.Strings(reusedBy[item.ID])
		out = append(out, KnowledgeView{KItem: &c, Output: clip(stripStepTag(item.Output), 200), ReusedBy: reusedBy[item.ID]})
	}
	return out
}

// GraphKnowledge is a node's knowledge lookup as the graph shows it.
type GraphKnowledge struct {
	Outcome   string `json:"outcome"` // hit miss
	Reason    string `json:"reason"`
	Item      string `json:"item,omitempty"`
	Source    string `json:"source,omitempty"`
	Published string `json:"published,omitempty"` // item this node's result became
	Revoked   string `json:"revoked,omitempty"`   // the reused or published item was revoked since
}

// graphKnowledge must be called with s.mu held.
func (s *Service) graphKnowledge(rec *StepRec) *GraphKnowledge {
	if rec.Lookup == nil && rec.Published == "" {
		return nil
	}
	gk := &GraphKnowledge{Published: rec.Published}
	if rec.Lookup != nil {
		gk.Outcome, gk.Reason, gk.Item, gk.Source = rec.Lookup.Outcome, rec.Lookup.Reason, rec.Lookup.Item, rec.Lookup.Source
	}
	for _, id := range []string{rec.ReusedFrom, rec.Published} {
		if item := s.st.Knowledge[id]; item != nil && item.Revoked != nil {
			gk.Revoked = fmt.Sprintf("%s 已于 %s 被 %s 撤回：%s", id, item.Revoked.At, item.Revoked.By, item.Revoked.Reason)
		}
	}
	return gk
}

// knowledgeRoutes: the list is read-only like /api/plans; revoking needs a
// demo account's password, like approving.
func (s *Service) knowledgeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/knowledge", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": s.Knowledge()})
	})
	mux.HandleFunc("/api/knowledge/revoke", func(w http.ResponseWriter, r *http.Request) {
		jsonOut := func(w http.ResponseWriter, code int, v any) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(v)
		}
		if r.Method != http.MethodPost {
			jsonOut(w, http.StatusMethodNotAllowed, map[string]string{"error": "只接受 POST"})
			return
		}
		var in struct {
			Item     string `json:"item"`
			Reason   string `json:"reason"`
			User     string `json:"user"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
			jsonOut(w, http.StatusBadRequest, map[string]string{"error": "请求格式不对"})
			return
		}
		if !validLogin(in.User, in.Password) {
			jsonOut(w, http.StatusUnauthorized, map[string]string{"error": "密码不对"})
			return
		}
		item, err := s.RevokeKnowledge(in.Item, in.User, in.Reason)
		if err != nil {
			jsonOut(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		jsonOut(w, http.StatusOK, item)
	})
}
