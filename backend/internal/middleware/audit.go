package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/logitrack/backend/internal/audit"
)

// AuditMiddleware wraps the handler chain so every 2xx response on a
// state-changing method emits a Record. Reads (GET/HEAD/OPTIONS) do
// not generate audit rows — they live in the access log.
//
// The middleware is deliberately coarse: it captures who, when, from,
// to at the HTTP envelope level. Fine-grained before/after diffs must
// be emitted by the service layer with direct Enqueue calls on the
// writer passed here (see audit.Writer).
func AuditMiddleware(w *audit.Writer) gin.HandlerFunc {
	mutating := map[string]struct{}{
		"POST":   {},
		"PUT":    {},
		"PATCH":  {},
		"DELETE": {},
	}
	return func(c *gin.Context) {
		if _, ok := mutating[c.Request.Method]; !ok {
			c.Next()
			return
		}
		c.Next()
		if w == nil {
			return
		}
		// Only audit successful or client-failed calls; 5xx go through
		// the error log, not the audit trail.
		if c.Writer.Status() >= 500 {
			return
		}
		rec := audit.Record{
			Method:        c.Request.Method,
			Path:          c.FullPath(),
			Status:        c.Writer.Status(),
			Action:        deriveAction(c),
			TargetKind:    deriveTargetKind(c.FullPath()),
			TargetID:      c.Param("id"),
			SourceIP:      c.ClientIP(),
			CorrelationID: c.GetHeader("X-Request-ID"),
		}
		if v, exists := c.Get(ClaimsContextKey); exists {
			if cl, ok := v.(*Claims); ok && cl != nil {
				rec.TenantID = cl.TenantID
				rec.ActorID = cl.UserID
				rec.ActorRoles = cl.Roles
			}
		}
		w.Enqueue(rec)
	}
}

// deriveAction produces a dotted action name from method and path, for
// example: POST /api/v1/shipments -> shipments.create.
func deriveAction(c *gin.Context) string {
	kind := deriveTargetKind(c.FullPath())
	verb := "modify"
	switch c.Request.Method {
	case "POST":
		verb = "create"
	case "PUT", "PATCH":
		verb = "update"
	case "DELETE":
		verb = "delete"
	}
	return kind + "." + verb
}

func deriveTargetKind(fullPath string) string {
	// Strip the /api/v1/ prefix and take the first segment.
	p := strings.TrimPrefix(fullPath, "/api/v1/")
	if idx := strings.IndexByte(p, '/'); idx > 0 {
		p = p[:idx]
	}
	if p == "" {
		p = "unknown"
	}
	return p
}
