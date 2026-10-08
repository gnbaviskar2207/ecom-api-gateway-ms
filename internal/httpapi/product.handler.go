package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/response"
	productsV1 "github.com/gnbaviskar2207/ecom-common/pkg/gen/products"
	"github.com/go-playground/validator"
)

type ProductHandler struct {
	logger    *slog.Logger
	client    productsV1.ProductServiceClient
	responder *response.Responder
}

type ProductClient interface {
	productsV1.ProductServiceClient
}

func NewProductHandler(logger *slog.Logger, client ProductClient, responder *response.Responder) *ProductHandler {
	return &ProductHandler{
		logger:    logger,
		client:    client,
		responder: responder,
	}
}

func (ph *ProductHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /product", ph.FindOneByPid)
}

var validate = validator.New()

type FindOneProductParams struct {
	Pid string `validate:"required" json:"pid"`
}

func (ph *ProductHandler) FindOneByPid(w http.ResponseWriter, r *http.Request) {
	params := FindOneProductParams{
		Pid: r.URL.Query().Get("pid"),
	}
	if err := validate.Struct(params); err != nil {
		ph.logger.WarnContext(r.Context(), "invalid request params", "error", err)
		ph.responder.WriteGRPCError(w, err)
		return
	}
	ph.logger.InfoContext(r.Context(), "fetching product from product service", "pid", params.Pid)
	rpcCtx, rpcCancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer rpcCancel()
	resp, err := ph.client.FindOneByPid(rpcCtx, &productsV1.FindOneByPidRequest{Pid: params.Pid})
	if err != nil {
		ph.logger.ErrorContext(r.Context(), "product service rpc failed with error", "error", err)
		ph.responder.WriteGRPCError(w, err)
		return
	}
	ph.logger.InfoContext(r.Context(), "product fetched successfully", "pid", params.Pid, "resp", resp)
	ph.responder.WriteSuccess(w, http.StatusOK, resp.GetProduct())
}
