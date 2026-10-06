package middlewares

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/response"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

type Middleware struct {
	logger    *slog.Logger
	responder *response.Responder
	limiter   *rate.Limiter
}

func New(logger *slog.Logger, responder *response.Responder, perSecond int, burst int) *Middleware {
	return &Middleware{
		logger:    logger,
		responder: responder,
		limiter:   rate.NewLimiter(rate.Limit(perSecond), burst),
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return m.requestID(m.recover(m.rateLimter(next)))
}

const HeaderRequestIDKey = "X-Request-ID"

func (m *Middleware) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		if spanContext := span.SpanContext(); spanContext.IsValid() {
			traceID := spanContext.TraceID().String()
			// set on response headers so clients can use the same traceId for their requests
			w.Header().Set(HeaderRequestIDKey, traceID)
			// also set on request headers so downstream handler can read it
			r.Header.Set(HeaderRequestIDKey, traceID)
		}
		next.ServeHTTP(w, r)
	})
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

func (m *Middleware) rateLimter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			m.responder.WriteErrorWithCode(w, http.StatusTooManyRequests, errors.New("request rate exceeded"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
