package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chef-stol/backend/internal/httpapi/dto"
)

func doJSON(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAuthHandler_RegisterSuccess(t *testing.T) {
	router := newTestRouter(newFakeAuthRepo(), &fakeSlotsRepo{})

	w := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "new@example.com",
		Password: "password1",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", w.Code, w.Body.String())
	}
	var resp dto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("response has empty token")
	}
	if resp.Client.Email != "new@example.com" {
		t.Fatalf("client.email = %q, want new@example.com", resp.Client.Email)
	}
}

func TestAuthHandler_RegisterDuplicateEmailIs409WithBadRequestCode(t *testing.T) {
	router := newTestRouter(newFakeAuthRepo(), &fakeSlotsRepo{})

	first := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "dup@example.com",
		Password: "password1",
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want 201", first.Code)
	}

	second := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "dup@example.com",
		Password: "password2",
	})
	if second.Code != http.StatusConflict {
		t.Fatalf("second register status = %d, want 409, body: %s", second.Code, second.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(second.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	// LOGIC-001: 409 uses code "bad_request", not a dedicated "email_taken" code.
	if body.Code != CodeBadRequest {
		t.Fatalf("code = %q, want %q", body.Code, CodeBadRequest)
	}
}

func TestAuthHandler_RegisterShortPasswordIs400(t *testing.T) {
	router := newTestRouter(newFakeAuthRepo(), &fakeSlotsRepo{})

	w := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "short@example.com",
		Password: "short1",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_LoginWrongCredentialsIs401(t *testing.T) {
	authRepo := newFakeAuthRepo()
	router := newTestRouter(authRepo, &fakeSlotsRepo{})

	reg := doJSON(t, router, http.MethodPost, "/auth/register", dto.RegisterRequest{
		Email:    "user@example.com",
		Password: "correctpass",
	})
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", reg.Code)
	}

	w := doJSON(t, router, http.MethodPost, "/auth/login", dto.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpass",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d, want 401, body: %s", w.Code, w.Body.String())
	}
	var body dto.Error
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Code != CodeUnauthorized {
		t.Fatalf("code = %q, want %q", body.Code, CodeUnauthorized)
	}
}
