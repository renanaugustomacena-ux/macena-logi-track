// Package problem implements RFC 7807 application/problem+json error
// responses. Every error returned from a LogiTrack handler must flow
// through Emit so the wire shape is consistent:
//
//   {
//     "type":     "https://logitrack.it/problems/invalid_body",
//     "title":    "Invalid request body",
//     "status":   400,
//     "detail":   "field 'reference' is required",
//     "instance": "/api/v1/shipments",
//     "code":     "invalid_body",
//     "traceId":  "<w3c traceparent>"
//   }
//
// The `code` and `traceId` extensions are LogiTrack-specific and do
// not violate the RFC 7807 spec (§3.2 allows extension members).
package problem

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TypeBase is the canonical URI prefix for LogiTrack problem types.
// Clients and support agents can resolve the link to a runbook entry.
const TypeBase = "https://logitrack.it/problems/"

// ContentType is the RFC 7807 media type.
const ContentType = "application/problem+json"

// Problem is the marshalled response body.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
	TraceID  string `json:"traceId,omitempty"`
}

// Emit writes a problem+json response with the given status and
// machine-code. Title is derived from the code (underscores to spaces,
// title-cased) unless title is non-empty.
func Emit(c *gin.Context, status int, code, detail string) {
	c.Writer.Header().Set("Content-Type", ContentType)
	p := Problem{
		Type:     TypeBase + code,
		Title:    titleFor(code),
		Status:   status,
		Detail:   detail,
		Instance: c.Request.URL.Path,
		Code:     code,
		TraceID:  c.GetHeader("traceparent"),
	}
	c.AbortWithStatusJSON(status, p)
}

// BadRequest is a shortcut for 400 Bad Request problems.
func BadRequest(c *gin.Context, code, detail string) {
	Emit(c, http.StatusBadRequest, code, detail)
}

// Unauthorized is a shortcut for 401.
func Unauthorized(c *gin.Context, code, detail string) {
	Emit(c, http.StatusUnauthorized, code, detail)
}

// Forbidden is a shortcut for 403.
func Forbidden(c *gin.Context, code, detail string) {
	Emit(c, http.StatusForbidden, code, detail)
}

// NotFound is a shortcut for 404.
func NotFound(c *gin.Context, code, detail string) {
	Emit(c, http.StatusNotFound, code, detail)
}

// Unprocessable is a shortcut for 422.
func Unprocessable(c *gin.Context, code, detail string) {
	Emit(c, http.StatusUnprocessableEntity, code, detail)
}

// TooMany is a shortcut for 429.
func TooMany(c *gin.Context, code, detail string) {
	Emit(c, http.StatusTooManyRequests, code, detail)
}

// Internal is a shortcut for 500.
func Internal(c *gin.Context, code, detail string) {
	Emit(c, http.StatusInternalServerError, code, detail)
}

// titleFor derives a human-readable title from the machine code.
func titleFor(code string) string {
	if code == "" {
		return "Error"
	}
	r := []rune(code)
	// Replace underscores with spaces and upper-case the first rune.
	for i, c := range r {
		if c == '_' {
			r[i] = ' '
		}
	}
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 32
	}
	return string(r)
}
