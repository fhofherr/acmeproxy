package httpapi

import (
	"net/http"
	"sync"
)

// MockHandlerFactory creates an http.Handler which remembers the params it
// extracted from a request.
//
// MockHandlerFactory may be called concurrently from multiple go routines. In
// this case the order of entries in Params is not guaranteed.
type MockHandlerFactory struct {
	DelegateHandler http.Handler
	Params          []map[string]string
	mu              sync.Mutex
}

// Handler creates a http.Handler which uses extractParams to extract parameters
// from req. The extracted parameters are stored in mh.Params.
func (mh *MockHandlerFactory) Handler(extractParams func(*http.Request) map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mh.mu.Lock()
		defer mh.mu.Unlock()

		mh.Params = append(mh.Params, extractParams(req))
		if mh.DelegateHandler == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mh.DelegateHandler.ServeHTTP(w, req)
	})
}
