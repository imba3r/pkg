package middleware

import (
	"fmt"
	"net/http"
)

// CacheControl implements a middleware which adds the Cache-Control header.
type CacheControl struct {
	value string
}

// NewCacheControl constructs a new CacheControl middleware.
func NewCacheControl(maxAge int, private bool) *CacheControl {
	cc := &CacheControl{
		value: fmt.Sprintf("max-age=%d", maxAge),
	}
	if private {
		cc.value += ", private"
	}
	return cc
}

// Handler returns the http.Handler of this middleware.
func (cc *CacheControl) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", cc.value)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
