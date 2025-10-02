package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
)

// MockServer wraps a httptest.Server with a simple handler map.
type MockServer struct {
	Server *httptest.Server
}

// JSONResponse creates a handler that returns JSON with status.
func JSONResponse(status int, payload any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// StartMockServer starts a mock server with per-path handlers.
func StartMockServer(routes map[string]http.Handler) (MockServer, string) {
	mux := http.NewServeMux()
	for p, h := range routes {
		mux.Handle(p, h)
	}
	srv := httptest.NewServer(mux)
	return MockServer{Server: srv}, srv.URL
}

// Close shuts down the mock server.
func (m MockServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// WithEnv temporarily sets environment variables for the duration of f.
func WithEnv(vars map[string]string, f func()) {
	// Save old values
	old := map[string]*string{}
	for k, v := range vars {
		if ov, ok := os.LookupEnv(k); ok {
			ovCopy := ov
			old[k] = &ovCopy
		} else {
			old[k] = nil
		}
		_ = os.Setenv(k, v)
	}
	defer func() {
		for k, v := range old {
			if v == nil {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, *v)
			}
		}
	}()
	f()
}
