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
	ready          bool
	sessionsActive int
	reloadFailures uint64
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
			"actions_total":   actions.Load(),
			"sessions_active": snap.sessionsActive,
			"reload_failures": snap.reloadFailures,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		snap := status()
		_, _ = fmt.Fprintf(w, "Portico SSH: %s\nActions run: %d\nSessions: %d\nReload failures: %d\nReady: %v\nConnect: %s\n",
			sshAddr, actions.Load(), snap.sessionsActive, snap.reloadFailures, snap.ready, sshConnectHint(sshAddr))
	})
	return mux
}

func sshConnectHint(addr string) string {
	host, port := hostPortFromAddr(addr)
	if port == "" {
		port = "2222"
	}
	if host == "" || host == "0.0.0.0" || host == "[::]" {
		return fmt.Sprintf("ssh -p %s ... localhost", port)
	}
	return fmt.Sprintf("ssh -p %s ... %s", port, host)
}

func hostPortFromAddr(addr string) (host, port string) {
	if addr == "" {
		return "", ""
	}
	if addr[0] == ':' {
		return "", addr[1:]
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:]
		}
	}
	return addr, ""
}

func defaultStatus(srv *runtime.Server, eng *runtime.Engine) StatusFunc {
	return func() statusSnapshot {
		ready := eng != nil && eng.Installation() != nil && !srv.Draining()
		return statusSnapshot{
			ready:          ready,
			sessionsActive: srv.ActiveSessions(),
			reloadFailures: eng.ReloadFailures(),
		}
	}
}
