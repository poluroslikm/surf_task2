package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chef-stol/backend/internal/domain"
)

func doForceCancel(router http.Handler, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/internal/slots/slot-1/force-cancel", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Internal-Token", token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestInternalHandler_ForceCancelSuccess(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "test-internal-token", `{"reason":"кухня затоплена"}`)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204, body: %s", w.Code, w.Body.String())
	}
	if !forceCancelRepo.called {
		t.Fatal("ForceCancel was not called")
	}
	if forceCancelRepo.calledSlotID != "slot-1" || forceCancelRepo.calledReason != "кухня затоплена" {
		t.Fatalf("ForceCancel called with slotID=%q reason=%q, want slotID=slot-1 reason=%q",
			forceCancelRepo.calledSlotID, forceCancelRepo.calledReason, "кухня затоплена")
	}
}

func TestInternalHandler_ForceCancelMissingTokenIs401(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "", `{"reason":"кухня затоплена"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
	if forceCancelRepo.called {
		t.Fatal("ForceCancel should not have been called without a valid token")
	}
}

func TestInternalHandler_ForceCancelWrongTokenIs401(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "not-the-right-token", `{"reason":"кухня затоплена"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
	if forceCancelRepo.called {
		t.Fatal("ForceCancel should not have been called with a wrong token")
	}
}

func TestInternalHandler_ForceCancelSlotNotFoundIs404(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{err: domain.ErrSlotNotFound}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "test-internal-token", `{"reason":"кухня затоплена"}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body: %s", w.Code, w.Body.String())
	}
}

func TestInternalHandler_ForceCancelAlreadyCancelledIs409(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{err: domain.ErrSlotAlreadyCancelled}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "test-internal-token", `{"reason":"кухня затоплена"}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body: %s", w.Code, w.Body.String())
	}
}

func TestInternalHandler_ForceCancelMissingReasonIs400(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "test-internal-token", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
	if forceCancelRepo.called {
		t.Fatal("ForceCancel should not have been called with a missing reason")
	}
}

func TestInternalHandler_ForceCancelEmptyReasonIs400(t *testing.T) {
	forceCancelRepo := &fakeForceCancelRepo{}
	router := newPushInternalTestRouter(newFakeAuthRepo(), &fakePushRepo{}, forceCancelRepo, "test-internal-token")

	w := doForceCancel(router, "test-internal-token", `{"reason":""}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
	if forceCancelRepo.called {
		t.Fatal("ForceCancel should not have been called with an empty reason")
	}
}
