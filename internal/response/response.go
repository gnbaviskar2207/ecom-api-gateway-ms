package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Responder struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Responder {
	return &Responder{logger: logger}
}

type Response struct {
	Success bool      `json:"success"`
	TraceId string    `json:"trace_id"`
	Data    any       `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`

	// rawData holds a pre-marshaled proto payload so the outer encoder
	// embeds it verbatim without re-encoding through encoding/json.
	rawData json.RawMessage
}

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *Responder) addHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// wireResponse is the actual shape written to the wire.
// It uses json.RawMessage for Data so a pre-marshaled proto payload
// is embedded verbatim, preserving zero-value / unpopulated fields.
type wireResponse struct {
	Success bool            `json:"success"`
	TraceId string          `json:"trace_id"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *ApiError       `json:"error,omitempty"`
}

var protoJSONMarshaler = protojson.MarshalOptions{
	// Include fields that are set to their default / zero values.
	EmitUnpopulated: true,
	// Use proto field names (snake_case) to match existing API contract.
	UseProtoNames: true,
}

func (r *Responder) writeJSON(w http.ResponseWriter, statusCode int, response Response) {
	r.addHeaders(w)
	w.WriteHeader(statusCode)
	response.TraceId = w.Header().Get("X-Request-ID")

	wire := wireResponse{
		Success: response.Success,
		TraceId: response.TraceId,
		Error:   response.Error,
	}

	if response.Data != nil {
		if msg, ok := response.Data.(proto.Message); ok {
			// Marshal proto message with protojson so zero-value fields are preserved.
			b, err := protoJSONMarshaler.Marshal(msg)
			if err != nil {
				r.logger.Error("Failed to marshal proto response", "error", err, "trace_id", response.TraceId)
			} else {
				wire.Data = b
			}

		} else {
			// Non-proto data: marshal normally and embed as raw JSON.
			b, err := json.Marshal(response.Data)
			if err != nil {
				r.logger.Error("Failed to marshal response data", "error", err, "trace_id", response.TraceId)
			} else {
				wire.Data = b
			}
		}
	}

	if err := json.NewEncoder(w).Encode(wire); err != nil {
		r.logger.Error("Failed to encode response", "error", err, "trace_id", response.TraceId)
	}
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
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
