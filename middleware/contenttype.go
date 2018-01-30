package middleware

import (
	"net/http"
)

// ContentType implements a middleware which adds the Content-Type header.
type ContentType struct {
	value string
}

// NewContentType constructs a new ContentType middleware.
func NewContentType(ct string) *ContentType {
	return &ContentType{ct}
}

// Handler returns the http.Handler of this middleware.
func (ct *ContentType) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ct.value)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
