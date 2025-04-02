package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-limit/internal/core/service"
	"github.com/go-limit/internal/core/model"
	"github.com/go-limit/internal/core/erro"

	"github.com/eliezerraj/go-core/coreJson"
	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component", "go-limit").Str("package", "internal.adapter.api").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

// Above create routers
func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Msg("Health")

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About add card
func (h *HttpRouters) GetTransactionLimit(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetTransactionLimit").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	span := tracerProvider.Span(req.Context(), "adapter.api.GetTransactionLimit")
	defer span.End()

	transactionLimit := model.TransactionLimit{}
	err := json.NewDecoder(req.Body).Decode(&transactionLimit)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.GetTransactionLimit(req.Context(), transactionLimit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}