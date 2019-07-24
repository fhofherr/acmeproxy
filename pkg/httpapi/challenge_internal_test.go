package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/httpapi/httpapitest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestServeHTTP01Challenge(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		token          string
		keyAuth        string
		solverErr      error
		expectedStatus int
	}{
		{
			name:           "solve challenge successfully",
			host:           "www.example.com",
			token:          "some-token",
			keyAuth:        "key-auth",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "solving the challenge failed",
			host:           "www.example.org",
			token:          "some-missing-token",
			solverErr:      httpapitest.ErrChallengeFailed{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "other error while solving the challenge",
			host:           "www.example.net",
			token:          "some-token",
			solverErr:      errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challengeSolver := &httpapitest.MockChallengeSolver{}
			challengeSolver.
				On("SolveChallenge", tt.host, tt.token).
				Return(tt.keyAuth, tt.solverErr)
			handler := challengeHandler{
				Params: func(*http.Request) map[string]string {
					return map[string]string{
						"token": tt.token,
					}
				},
				Solver: challengeSolver,
			}
			target := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s",
				tt.host, tt.token)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/octet-stream", rr.Header().Get("content-type"))
			assert.Equal(t, tt.keyAuth, rr.Body.String())

			challengeSolver.AssertCalled(t, "SolveChallenge", tt.host, tt.token)
		})
	}
}
