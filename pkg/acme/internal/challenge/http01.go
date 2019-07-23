package challenge

import (
	"fmt"
	"sync"
)

// HTTP01Solver is a custom challenge provider for the HTTP01 challenge.
//
// In contrast to the default HTTP01 challenge provider shipped with lego it
// does not start a web-server but instead provides a handle challenge method
// which can be used in an http.Handler or http.HandlerFunc.
//
// HTTP01Solver is safe for concurrent access by multiple Go routines.
//
// The methods Present and CleanUp are intended for use by lego and should not
// be called directly.
type HTTP01Solver struct {
	challenges map[string]string
	mu         *sync.Mutex
}

// NewHTTP01Solver creates and initializes an HTTP01Solver.
func NewHTTP01Solver() *HTTP01Solver {
	return &HTTP01Solver{
		challenges: make(map[string]string),
		mu:         &sync.Mutex{},
	}
}

// Present registers a solution for a HTTP01 challenge with the HTTP01Solver.
//
// This method is intended to be used by lego and should not be called directly.
func (p *HTTP01Solver) Present(domain, token, keyAuth string) error {
	k := challengeKey(domain, token)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.challenges[k] = keyAuth
	return nil
}

// CleanUp removes the solution for a HTTP01 challenge from the HTTP01Solver.
//
// This method is intended to be used by lego and should not be called directly.
func (p *HTTP01Solver) CleanUp(domain, token, keyAuth string) error {
	k := challengeKey(domain, token)

	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.challenges, k)
	return nil
}

// HandleChallenge tries to find the key authorization for the passed domain
// and token.
//
// If the key authorization could not be found an instance of ErrChallengeFailed
// is returned as error.
func (p *HTTP01Solver) HandleChallenge(domain, token string) (string, error) {
	k := challengeKey(domain, token)

	p.mu.Lock()
	defer p.mu.Unlock()

	keyAuth, ok := p.challenges[k]
	if !ok {
		return "", ErrChallengeFailed{Domain: domain, Token: token}
	}
	return keyAuth, nil
}

// ErrChallengeFailed signals that the HTTP01Solver could not solve the challenge.
type ErrChallengeFailed struct {
	Domain string
	Token  string
}

func (e ErrChallengeFailed) Error() string {
	return fmt.Sprintf("challenge failed for: %s", e.Domain)
}

func challengeKey(domain, token string) string {
	return fmt.Sprintf("%s__%s", domain, token)
}
