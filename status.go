package portico

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/MrBlindMouse/Portico/runtime"
)

// StatusFunc supplies readiness and metrics for the status HTTP mux.
type StatusFunc func() statusSnapshot

type statusSnapshot struct {
	ready            bool
	sessionsActive   int
	reloadFailures   uint64
	hasDataDir       bool
}

func statusHandler(sshAddr string, actions *atomic.Uint64, status StatusFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		snap := status()
		if !snap.ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready\n"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		snap := status()
		payload := map[string]any{
			"actions_total":     actions.Load(),
			"sessions_active": snap.sessionsActive,
			"reload_failures":   snap.reloadFailures,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		snap := status()
		_, _ = fmt.Fprintf(w, "Portico SSH: %s\nActions run: %d\nSessions: %d\nReload failures: %d\nReady: %v\nConnect: ssh -p %s ... localhost\n",
			sshAddr, actions.Load(), snap.sessionsActive, snap.reloadFailures, snap.ready, portFromAddr(sshAddr))
	})
	return mux
}

func portFromAddr(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[i+1:]
		}
	}
	return "2222"
}

func defaultStatus(srv *runtime.Server, eng *runtime.Engine, dataDir string) StatusFunc {
	return func() statusSnapshot {
		ready := eng != nil && eng.Experience() != nil && !srv.Draining()
		_ = dataDir
		return statusSnapshot{
			ready:          ready,
			sessionsActive: srv.ActiveSessions(),
			reloadFailures: eng.ReloadFailures(),
			hasDataDir:     dataDir != "",
		}
	}
}
