package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

type Recoverer struct {
}

func NewRecoverer() *Recoverer {
	return &Recoverer{}
}

func (ct *Recoverer) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
				debug.PrintStack()
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
