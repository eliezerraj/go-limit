package api

import (
	"fmt"
	"encoding/json"
	"reflect"
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
	childLogger.Info().Str("func","Health").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About show all context values
func (h *HttpRouters) Context(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Context").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	contextValues := reflect.ValueOf(req.Context()).Elem()
	json.NewEncoder(rw).Encode(fmt.Sprintf("%v",contextValues))
}

// About show pgx stats
func (h *HttpRouters) Stat(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	res := h.workerService.Stat(req.Context())

	json.NewEncoder(rw).Encode(res)
}

// About check and transaction
func (h *HttpRouters) CheckLimitTransaction(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","CheckLimitTransaction").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	span := tracerProvider.Span(req.Context(), "adapter.api.CheckLimitTransaction")
	defer span.End()

	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	limit := model.Limit{}
	err := json.NewDecoder(req.Body).Decode(&limit)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.CheckLimitTransaction(req.Context(), limit)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}