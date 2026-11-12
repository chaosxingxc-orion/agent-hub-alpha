package e2e

// MR1Review probes — vendored VERBATIM from the MR !1 review's reproduction
// attachment (comment 199470012; author: xingchao_ml). These assert the
// ENTRY CONTRACT as the reviewer specified it; per the reply commitment
// (comment 199763878) they are permanent suite members — the fix batch is
// done when all six pass. No assertion is modified.

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
)

func reviewFirst(op string) *firstExecInput {
	return &firstExecInput{req: requestForT2(op, "er-1", "ev-1", t2Digest('a'), "br-"+op), evidence: creditEvidence{assetMillions: 35, assetReportOK: true}}
}
func reviewModel() *countingModel {
	return newCountingModel(map[string]creditEvidence{"*": reviewFirst("unused").evidence})
}

func TestMR1Review_PositiveCThenA(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	cm := newCountingModel(map[string]creditEvidence{"*": reviewFirst("op-good").evidence})
	out, err := h.nodeEntryExec("op-good", chainGrants(), "unused", nil, reviewFirst("op-good"), Ports{Model: cm})
	if err != nil || out.Branch != "C-model" {
		t.Fatalf("initial: %+v %v", out, err)
	}
	if o := h.completeNode("op-good"); o.Status != "completed" {
		t.Fatalf("completion: %+v", o)
	}
	h = h.reopen()
	retry, err := h.nodeEntryExec("op-good", chainGrants(), "unused", nil, reviewFirst("op-good"), Ports{Model: cm})
	if err != nil || retry.Branch != "A-retry" || retry.Result.Receipt != out.Result.Receipt || cm.calls != 1 {
		t.Fatalf("retry: %+v %v calls=%d", retry, err, cm.calls)
	}
}

func TestMR1Review_ReadOnlyCannotDispatch(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	cm := reviewModel()
	out, err := h.nodeEntryExec("op-read", fullGrantsN3(), "unused", nil, reviewFirst("op-read"), Ports{Model: cm})
	if err == nil || h.machine.State("op-read") == adopt.StateDispatched {
		t.Fatalf("read-only caller dispatched: branch=%s state=%s calls=%d err=%v", out.Branch, h.machine.State("op-read"), cm.calls, err)
	}
}

func TestMR1Review_MismatchedIdentityHasNoEffects(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	cm := reviewModel()
	_, err := h.nodeEntryExec("op-outer", chainGrants(), "unused", nil, reviewFirst("op-inner"), Ports{Model: cm})
	if err == nil || cm.calls != 0 || h.machine.State("op-inner") != "" {
		t.Fatalf("identity mismatch polluted state: outer=%q inner=%q calls=%d err=%v", h.machine.State("op-outer"), h.machine.State("op-inner"), cm.calls, err)
	}
}

func TestMR1Review_ChangedRetryMustConflict(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	cm := reviewModel()
	if _, err := h.nodeEntryExec("op-conflict", chainGrants(), "unused", nil, reviewFirst("op-conflict"), Ports{Model: cm}); err != nil {
		t.Fatal(err)
	}
	if o := h.completeNode("op-conflict"); o.Status != "completed" {
		t.Fatalf("completion: %+v", o)
	}
	changed := reviewFirst("op-conflict")
	changed.req.PayloadDigest = t2Digest('b')
	changed.req.Identity.BusinessRequestID = "br-other"
	changed.evidence.assetMillions = 0
	out, err := h.nodeEntryExec("op-conflict", chainGrants(), "unused", nil, changed, Ports{Model: cm})
	if err == nil {
		t.Fatalf("changed command silently accepted as retry: branch=%s stored_br=%s requested_br=%s calls=%d", out.Branch, out.Result.BusinessReqID, changed.req.Identity.BusinessRequestID, cm.calls)
	}
}

func TestMR1Review_KnowledgeResultSurvivesRetry(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	cm := reviewModel()
	h.registry.mustPublish(t, knowledgeMethod{ID: "review-method", Version: "1.0.0", ServesNode: "history-verification", InputFields: []string{"overdue_months", "business_request_id"}, Valid: true})
	in := map[string]string{"overdue_months": "0", "business_request_id": "br-op-method"}
	first := reviewFirst("op-method")
	original, err := h.nodeEntryExec("op-method", chainGrants(), "history-verification", in, first, Ports{Model: cm})
	if err != nil || original.Branch != "B-method" {
		t.Fatalf("initial: %+v %v", original, err)
	}
	_, stored := h.store.get("op-method")
	state := h.machine.State("op-method")
	repeat, err := h.nodeEntryExec("op-method", chainGrants(), "history-verification", in, first, Ports{Model: cm})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.registry.invalidate("review-method"); err != nil {
		t.Fatal(err)
	}
	h = h.reopen()
	out, err := h.nodeEntryExec("op-method", chainGrants(), "history-verification", in, first, Ports{Model: cm})
	if err != nil || out.Branch != "A-retry" || out.Result.Receipt != original.Result.Receipt || cm.calls != 0 {
		t.Fatalf("successful knowledge result not reusable: original=%s stored=%v state=%q immediate_retry=%s retry_after_invalidation=%s model_calls=%d err=%v", original.Result.Receipt, stored, state, repeat.Branch, out.Branch, cm.calls, err)
	}
}

type reviewBlockingModel struct {
	calls   atomic.Int32
	entered chan struct{}
	release chan struct{}
}

func (m *reviewBlockingModel) Invoke(_ string, e creditEvidence) s1Verdict {
	m.calls.Add(1)
	m.entered <- struct{}{}
	<-m.release
	return s1Judge(e)
}

func TestMR1Review_ConcurrentSameOperationSingleInvocation(t *testing.T) {
	h := newT2Host(t, t.TempDir())
	model := &reviewBlockingModel{entered: make(chan struct{}, 2), release: make(chan struct{})}
	done := make(chan error, 2)
	invoke := func() {
		_, err := h.nodeEntryExec("op-race", chainGrants(), "unused", nil, reviewFirst("op-race"), Ports{Model: model})
		done <- err
	}
	go invoke()
	select {
	case <-model.entered:
	case <-time.After(5 * time.Second):
		close(model.release)
		t.Fatal("first invocation did not enter")
	}
	go invoke()
	finished := 0
	select {
	case <-model.entered:
	case <-done:
		finished = 1
	case <-time.After(time.Second):
	}
	close(model.release)
	for i := finished; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("entry did not finish")
		}
	}
	if model.calls.Load() != 1 {
		t.Fatalf("same operation called model %d times before admission", model.calls.Load())
	}
}
