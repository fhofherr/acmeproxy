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

func TestHandleHTTP01Challenge(t *testing.T) {
	tests := []struct {
		name           string
		domain         string
		port           string
		token          string
		keyAuth        string
		solverErr      error
		expectedStatus int
	}{
		{
			name:           "solve challenge successfully",
			domain:         "www.example.com",
			token:          "some-token",
			keyAuth:        "key-auth",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ignore port number when solving challenge",
			domain:         "www.example.com",
			port:           "8080",
			token:          "some-token",
			keyAuth:        "key-auth",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "solving the challenge failed",
			domain:         "www.example.org",
			token:          "some-missing-token",
			solverErr:      httpapitest.ErrChallengeFailed{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "other error while solving the challenge",
			domain:         "www.example.net",
			token:          "some-token",
			solverErr:      errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			challengeSolver := &httpapitest.MockChallengeSolver{}
			challengeSolver.Test(t)
			challengeSolver.
				On("SolveChallenge", tt.domain, tt.token).
				Return(tt.keyAuth, tt.solverErr)
			handler := challengeHandler{
				Params: func(*http.Request) map[string]string {
					return map[string]string{
						"token": tt.token,
					}
				},
				Solver: challengeSolver,
			}
			host := tt.domain
			if tt.port != "" {
				host = fmt.Sprintf("%s:%s", tt.domain, tt.port)
			}
			target := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s",
				host, tt.token)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/octet-stream", rr.Header().Get("content-type"))
			assert.Equal(t, tt.keyAuth, rr.Body.String())

			challengeSolver.AssertCalled(t, "SolveChallenge", tt.domain, tt.token)
		})
	}
}
