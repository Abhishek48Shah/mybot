package middleware

import (
	"fmt"
	"net/http"

	"github.com/Abhishek48Shah/bot/internal/util"
)

func (m *Middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				util.WriteError(w, r, util.InternalServerErr("internal server error", fmt.Errorf("panic recovery: %w", err)), m.logger)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
