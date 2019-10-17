package acmeclient

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/fhofherr/golf/log"
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
//
// The zero value of HTTP01Solver is fully functional.
type HTTP01Solver struct {
	Logger     log.Logger
	challenges map[string]string
	mu         sync.Mutex
}

// Present registers a solution for a HTTP01 challenge with the HTTP01Solver.
//
// This method is intended to be used by lego and should not be called directly.
func (p *HTTP01Solver) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.challenges == nil {
		p.challenges = make(map[string]string)
	}

	key := challengeKey(domain, token)
	p.challenges[key] = keyAuth
	return nil
}

// CleanUp removes the solution for a HTTP01 challenge from the HTTP01Solver.
//
// This method is intended to be used by lego and should not be called directly.
func (p *HTTP01Solver) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := challengeKey(domain, token)
	delete(p.challenges, key)
	return nil
}

// Handler creates an http.Handler serving the actual HTTP01 challenge.
//
// The extractParams function is used to extract the required parameters from
// the request. It must return a map containing the keys domain and token, or nil if the map cannot be constructed.
func (p *HTTP01Solver) Handler(extractParams func(*http.Request) map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()

		params := extractParams(req)
		domain := params["domain"]
		token := params["token"]

		key := challengeKey(domain, token)
		keyAuth, ok := p.challenges[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			p.writeBody(w, []byte("Not found"))
			return
		}
		p.writeBody(w, []byte(keyAuth))
	})
}

func (p *HTTP01Solver) writeBody(w http.ResponseWriter, body []byte) {
	if _, err := w.Write(body); err != nil {
		log.Log(p.Logger, "level", "warn", "error", err)
	}
}

func challengeKey(domain, token string) string {
	return fmt.Sprintf("%s__%s", domain, token)
}
