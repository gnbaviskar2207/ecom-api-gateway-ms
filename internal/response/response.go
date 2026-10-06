package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Responder struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

type Response struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`
}

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *Responder) addHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func (r *Responder) writeJSON(w http.ResponseWriter, statusCode int, response Response) {
	r.addHeaders(w)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (r *Responder) WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	r.writeJSON(w, statusCode, Response{
		Success: true,
		Data:    data,
	})
}

func (r *Responder) WriteGRPCError(w http.ResponseWriter, err error) {
	// httpStatus, apiErr := grpcErrToHttpErr(err)
	st, ok := status.FromError(err)
	if !ok {
		// internal error
		r.writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error: &ApiError{
				Code:    "Internal Server Error",
				Message: err.Error(),
			},
		})
		return
	}
	httpStatus := grpcCodeToHttpStatus(st.Code())
	r.writeJSON(w, httpStatus, Response{
		Success: false,
		Error: &ApiError{
			Code:    st.Code().String(),
			Message: err.Error(),
		},
	})
}

func (r *Responder) WriteErrorWithCode(w http.ResponseWriter, errCode int, err error) {
	r.writeJSON(w, errCode, Response{
		Success: false,
		Error: &ApiError{
			Code:    http.StatusText(errCode),
			Message: err.Error(),
		},
	})
}

func grpcCodeToHttpStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
