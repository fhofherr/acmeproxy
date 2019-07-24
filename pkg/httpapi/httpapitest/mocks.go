package httpapitest

import "github.com/stretchr/testify/mock"

// MockChallengeSolver is a mock implementation of the httpapi.ChallengeSolver
// interface.
type MockChallengeSolver struct {
	mock.Mock
}

// SolveChallenge registers the call to the mock and returns the configured
// return values.
func (m *MockChallengeSolver) SolveChallenge(domain, token string) (string, error) {
	args := m.Called(domain, token)
	return args.String(0), args.Error(1)
}

// ErrChallengeFailed represents a failed attempt to solve a challenge during
// testing.
type ErrChallengeFailed struct{}

// ChallengeFailed signals if this challenge failed.
func (ErrChallengeFailed) ChallengeFailed() bool {
	return true
}

func (ErrChallengeFailed) Error() string {
	return "Challenge failed"
}
