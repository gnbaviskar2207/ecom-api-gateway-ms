package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	productsV1 "github.com/gnbaviskar2207/ecom-common/pkg/gen/products"
)

type ProductHandler struct {
	logger *slog.Logger
	client productsV1.ProductServiceClient
}

type ProductClient interface {
	productsV1.ProductServiceClient
}

func NewProductHandler(logger *slog.Logger, client ProductClient) *ProductHandler {
	return &ProductHandler{
		logger: logger,
		client: client,
	}
}

func (ph *ProductHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/product", ph.FindOneByPid)
}

func (ph *ProductHandler) FindOneByPid(w http.ResponseWriter, r *http.Request) {
	resp, err := ph.client.FindOneByPid(r.Context(), &productsV1.FindOneByPidRequest{Pid: "6A325F38-D56D-4626-95A0-5FD54F151C8E"})
	if err != nil {
		writeErrorJSONResponse(w, err)
		return
	}

	writeSuccessJSONResponse(w, resp.GetProduct())
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func writeSuccessJSONResponse(w http.ResponseWriter, data interface{}) {
	writeJSONResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func writeErrorJSONResponse(w http.ResponseWriter, err error) {
	writeJSONResponse(w, http.StatusInternalServerError, Response{
		Success: false,
		Error:   err.Error(),
	})
}
