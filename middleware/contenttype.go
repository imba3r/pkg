package middleware

import (
	"net/http"
)

type ContentType struct {
	value string
}

func NewContentType(ct string) *ContentType {
	return &ContentType{ct}
}

func (ct *ContentType) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ct.value)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
