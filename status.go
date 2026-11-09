package portico

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func statusHandler(sshAddr string, actions *atomic.Uint64) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "Portico SSH: %s\nActions run: %d\nConnect: ssh -p %s ... localhost\n",
			sshAddr, actions.Load(), portFromAddr(sshAddr))
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
