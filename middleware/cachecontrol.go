package middleware

import (
	"fmt"
	"net/http"
)

type CacheControl struct {
	value string
}

func NewCacheControl(maxAge int, private bool) *CacheControl {
	cc := &CacheControl{
		value: fmt.Sprintf("max-age=%d", maxAge),
	}
	if private {
		cc.value += ", private"
	}
	return cc
}

func (cc *CacheControl) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", cc.value)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
