package httpapi

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/httpapi/dto"
)

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"email taken", domain.ErrEmailTaken, 409, CodeBadRequest},
		{"invalid credentials", domain.ErrInvalidCredentials, 401, CodeUnauthorized},
		{"unauthorized", domain.ErrUnauthorized, 401, CodeUnauthorized},
		{"slot not found", domain.ErrSlotNotFound, 404, CodeNotFound},
		{"unknown error", errors.New("boom"), 500, CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mapDomainError(w, tt.err)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			var body dto.Error
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", body.Code, tt.wantCode)
			}
		})
	}
}
