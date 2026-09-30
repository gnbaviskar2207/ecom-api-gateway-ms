package httpapi

import (
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
	// write json response
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(`{"message": "Product handler"}`))
	resp, err := ph.client.FindOneByPid(r.Context(), &productsV1.FindOneByPidRequest{Pid: "6A325F38-D56D-4626-95A0-5FD54F151C8E"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.GetProduct().String()))
}
