package middleware

import (
	"net/http"
	"runtime/debug"
)

// Recoverer implements a middleware which recovers from panics.
type Recoverer struct {
	logger Logger
}

// Logger interface describes the kind of logger we'd like to have.
type Logger interface {
	Errorf(msgFormat string, args ...interface{})
}

// NewRecoverer constructs a new Recoverer middleware with an optional logger.
func NewRecoverer(logger Logger) *Recoverer {
	return &Recoverer{logger}
}

// Handler returns the http.Handler of this middleware.
func (rvr *Recoverer) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover := recover(); recover != nil {
				rvr.error("Panic: %+v\n", recover)
				rvr.error("%x", debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (rvr *Recoverer) error(msgFormat string, args ...interface{}) {
	if rvr.logger != nil {
		rvr.logger.Errorf(msgFormat, args...)
	}
}
