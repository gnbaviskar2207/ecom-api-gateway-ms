package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/response"
)

type Middleware struct {
	logger    *slog.Logger
	responder *response.Responder
}

func New(logger *slog.Logger, responder *response.Responder) *Middleware {
	return &Middleware{
		logger:    logger,
		responder: responder,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return m.recover(next)
}

func (m *Middleware) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				m.logger.ErrorContext(r.Context(), "Internal server error - panic recovered", "error", recovered, slog.String("stack", string(debug.Stack())))
				m.responder.WriteErrorWithCode(w, http.StatusInternalServerError, errors.New("internal server error - panic recovered"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
