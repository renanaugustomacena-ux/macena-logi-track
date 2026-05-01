// Package obs implements a tiny hand-rolled Prometheus exposition
// endpoint shared between the HTTP middleware (counter increments) and
// the /metrics handler (serialisation).
//
// Keeping it in its own package breaks the handlers ↔ middleware import
// cycle that a single-package solution would require.
package obs

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Render writes the exposition text to w. Called by the /metrics handler.
func Render(w http.ResponseWriter, service, version string, started time.Time) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "# HELP logitrack_build_info Static build metadata.\n")
	_, _ = fmt.Fprintf(w, "# TYPE logitrack_build_info gauge\n")
	_, _ = fmt.Fprintf(w, "logitrack_build_info{service=%q,version=%q,go_version=%q} 1\n",
		service, version, runtime.Version())

	_, _ = fmt.Fprintf(w, "# HELP process_start_time_seconds Start time of the process since UNIX epoch.\n")
	_, _ = fmt.Fprintf(w, "# TYPE process_start_time_seconds gauge\n")
	_, _ = fmt.Fprintf(w, "process_start_time_seconds %d\n", started.Unix())

	_, _ = fmt.Fprintf(w, "# HELP logitrack_http_requests_total HTTP requests, labelled by method/path/status.\n")
	_, _ = fmt.Fprintf(w, "# TYPE logitrack_http_requests_total counter\n")
	httpCounters.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(*uint64)
		_, _ = fmt.Fprintf(w, "logitrack_http_requests_total{%s} %d\n", k, atomic.LoadUint64(v))
		return true
	})

	_, _ = fmt.Fprintf(w, "# HELP logitrack_ws_connections_active Active WebSocket subscribers.\n")
	_, _ = fmt.Fprintf(w, "# TYPE logitrack_ws_connections_active gauge\n")
	_, _ = fmt.Fprintf(w, "logitrack_ws_connections_active %d\n", atomic.LoadInt64(&wsConnections))

	_, _ = fmt.Fprintf(w, "# HELP logitrack_ws_broadcasts_total Total events broadcast.\n")
	_, _ = fmt.Fprintf(w, "# TYPE logitrack_ws_broadcasts_total counter\n")
	_, _ = fmt.Fprintf(w, "logitrack_ws_broadcasts_total %d\n", atomic.LoadUint64(&wsBroadcasts))

	_, _ = fmt.Fprintf(w, "# HELP go_goroutines Current goroutines.\n")
	_, _ = fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
	_, _ = fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
}

// Global counters — package-private so callers go through the helpers.
var (
	httpCounters  sync.Map
	wsConnections int64
	wsBroadcasts  uint64
)

// IncHTTP increments the HTTP counter for a request. The full path is
// reduced to Gin's route template upstream (see middleware.PromMiddleware)
// so cardinality stays bounded.
func IncHTTP(method, path string, status int) {
	key := fmt.Sprintf("method=%q,path=%q,status=%q", method, path, fmt.Sprintf("%d", status))
	raw, ok := httpCounters.Load(key)
	if !ok {
		var zero uint64
		raw, _ = httpCounters.LoadOrStore(key, &zero)
	}
	atomic.AddUint64(raw.(*uint64), 1)
}

// IncWSConnect / IncWSDisconnect / IncWSBroadcast are called by the hub.
func IncWSConnect()    { atomic.AddInt64(&wsConnections, 1) }
func IncWSDisconnect() { atomic.AddInt64(&wsConnections, -1) }
func IncWSBroadcast()  { atomic.AddUint64(&wsBroadcasts, 1) }
