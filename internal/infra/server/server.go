package server

import (
	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"

	"github.com/go-limit/internal/adapter/api"	
	"github.com/go-limit/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"  
	go_core_midleware "github.com/eliezerraj/go-core/middleware"
	
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	// trace
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/propagation"
	 sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	// Metrics
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	
	// Logs
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	//"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	logotel "go.opentelemetry.io/otel/sdk/log"
	sdklog "go.opentelemetry.io/otel/log"
)

var (
	childLogger = log.With().
					Str("component","go-limit").
					Str("component","internal.infra.server").
					Logger()
	core_middleware go_core_midleware.ToolsMiddleware
	tracerProvider 	go_core_observ.TracerProvider
	infoTrace 		go_core_observ.InfoTrace
	tracer			trace.Tracer
)

type HttpServer struct {
	httpServer	*model.Server
}

// About create new http server
func NewHttpAppServer(httpServer *model.Server) HttpServer {
	childLogger.Info().
				Str("func","NewHttpAppServer").Send()

				
	return HttpServer{httpServer: httpServer }
}

// About initialize MeterProvider with Prometheus exporter
func initMeterProvider(ctx context.Context, serviceName string) (*sdkmetric.MeterProvider, error) {
	childLogger.Info().
				Str("func","initMeterProvider").
				Send()

	// 1. Configurar o Recurso OTel
	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("env", "DEV"),
		),
	)
	if err != nil {
		return nil, err
	}

	// 2. Criar o Prometheus Exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// 3. Criar o MeterProvider, usando o Prometheus Exporter como Reader.
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	return provider, nil
}

//About initializa logs
func initLogger(ctx context.Context, serviceName string) (*logotel.LoggerProvider, error) {
	childLogger.Info().
				Str("func","initLogger").
				Send()

	// Configure the OTel Resource (Crucial for Loki labels/service identification)
	res, err := resource.New(ctx,
							 resource.WithSchemaURL(semconv.SchemaURL),
							 resource.WithAttributes(
								semconv.ServiceNameKey.String(serviceName),
								semconv.ServiceVersion("1.0.0"),
								attribute.String("env", "DEV"),
							 ),
							)
	if err != nil {
		return nil, err
	}

	// Create an OTLP gRPC exporter that sends logs to the OTEL Collector
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint("localhost:4317"), 
	)

	/*exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint("localhost:4318"),
		// Optional: send to custom path if collector uses a subpath like "/v1/logs"
		otlploghttp.WithURLPath("/v1/logs"),
	)*/

	if err != nil {
		return nil, err
	}

	provider := logotel.NewLoggerProvider(
		logotel.WithResource(res),
		logotel.WithProcessor(
			logotel.NewBatchProcessor(exporter),
		),
	)

	return provider, nil
}

func EmitOtelLog(ctx context.Context, 
				lp *logotel.LoggerProvider, 
				msg string, 
				severity sdklog.Severity, 
				extraAttrs ...attribute.KeyValue) {
	fmt.Println("1 EmitOtelLog")

	if lp == nil {
		return
	}

	// 1. Get the OTel Logger instance
    otelLogger := lp.Logger("go-limit")

    // 2. Create the log record
    rec := sdklog.Record{}

	// Set core record data
	rec.SetTimestamp(time.Now())
	rec.SetSeverity(severity)
	rec.SetBody(sdklog.StringValue(msg))

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fmt.Printf("2 TraceID: %d \n", sc.TraceID().String())
		fmt.Printf("3 SpanID: %d \n", sc.SpanID().String())

		rec.AddAttributes(
			sdklog.String("trace_id", sc.TraceID().String()),
			sdklog.String("span_id", sc.SpanID().String()),
		)
	}

	//var logAttrs []logotelapi.Attribute
	for _, kv := range extraAttrs {
		fmt.Printf("4 Atributew %s %s \n", string(kv.Key), kv.Value.AsInterface() )
		k := string(kv.Key)
		v := fmt.Sprintf("%v", kv.Value)
		rec.AddAttributes(
			sdklog.String( k , v ),
		)
    }
	// Emit the record*/

	fmt.Printf("5 rec: %v \n", rec)
	childLogger.Info().Str("func","EmitOtelLog").Interface("rec", rec).Send()

	otelLogger.Emit(ctx, rec)
}

//About Create Custom Metrics
var httpRequestsCounter metric.Int64Counter
var httpLatencyHistogram metric.Float64Histogram
var err_metric error

func setupCustomMetrics(meter metric.Meter) error {
	childLogger.Info().Str("func","setupCustomMetrics").Send()

	httpRequestsCounter, err_metric = meter.Int64Counter("eliezer-http_requests_total",
				metric.WithDescription("Total number of HTTP requests by path"),
				metric.WithUnit("1"),
	)
	if err_metric != nil {
		childLogger.Error().Err(err_metric).Msg("Erro Create Custom Metrics")
		return err_metric
	}

	httpLatencyHistogram, err_metric = meter.Float64Histogram("eliezer-http_server_latency_seconds",
		metric.WithDescription("Latency of HTTP server requests by path"),
		metric.WithUnit("s"),
	)	
	if err_metric != nil {
		childLogger.Error().Err(err_metric).Msg("Erro Create Custom Metrics")
		return err_metric
	}

	return nil
}

// About start http server
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().
				Str("func","StartHttpAppServer").
				Send()
			
	// --------- OTEL traces ---------------
	var initTracerProvider *sdktrace.TracerProvider
	
	if appServer.InfoPod.OtelTraces {
		infoTrace.PodName = appServer.InfoPod.PodName
		infoTrace.PodVersion = appServer.InfoPod.ApiVersion
		infoTrace.ServiceType = "k8-workload"
		infoTrace.Env = appServer.InfoPod.Env
		infoTrace.AccountID = appServer.InfoPod.AccountID

		initTracerProvider = tracerProvider.NewTracerProvider(	ctx, 
																appServer.ConfigOTEL, 
																&infoTrace)

		otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTracerProvider(initTracerProvider)
		tracer = initTracerProvider.Tracer(appServer.InfoPod.PodName)
	}

	// --------- OTEL metrics ---------------
	var meterProvider *sdkmetric.MeterProvider

	if appServer.InfoPod.OtelMetrics {
		meterProvider, err := initMeterProvider(ctx, infoTrace.PodName)
		if err != nil {
			childLogger.Error().Err(err).Msg("Error start Otel Metrics Provider")
		} else {
			meter := meterProvider.Meter(infoTrace.PodName)

			setupCustomMetrics(meter)
			if err != nil {
				childLogger.Info().Msg("Erro Create Custom Metrics")
			}

			childLogger.Info().Msg("Otel Metrics Provider started SUCCESSFULL")
		}
	}

	// ------------- OTEL logs -------------------
	var logProvider *logotel.LoggerProvider
	var err_log error

	if appServer.InfoPod.OtelLogs {
		logProvider, err_log = initLogger(ctx, infoTrace.PodName)
		if err_log != nil {
			childLogger.Error().
						Err(err_log).
						Msg("Erro initialize Otel Logger Provider")
		} else {
			childLogger.Info().
						Msg("Otel Logger Provider started SUCCESSFULL")
		}
	}

	// handle the final actions
	defer func() {

		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				childLogger.Error().
							Err(err).
							Msg("failed to stop instrumentation")
			}
		}

		if logProvider != nil {
			if err := logProvider.Shutdown(ctx); err != nil {
				childLogger.Error().
							Err(err).
							Msg("failed to shutdown otel log provider")
			}
		}

		if initTracerProvider != nil {
			err := initTracerProvider.Shutdown(ctx)
			if err != nil{
				childLogger.Error().
							Err(err).
							Send()
			}
		}

		childLogger.Info().Msg("stop done !!!")
	}()
	
	// Routers
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)

	// Prometheus metrics	
	myRouter.Handle("/metrics", promhttp.Handler())

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(appServer)
	})

	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpRouters.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpRouters.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpRouters.Header)

	wk_ctx := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    wk_ctx.HandleFunc("/context", httpRouters.Context)
	
	stat := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    stat.HandleFunc("/stat", httpRouters.Stat)

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().Str("HandleFunc","/info").Send()
		start := time.Now()
		targetPath := "/info"

		req_ctx, cancel := context.WithTimeout(req.Context(), 5 * time.Second)
    	defer cancel()

		req_ctx, span := tracerProvider.SpanCtx(req_ctx, "adapter.api.info")
		defer span.End()

		defer func() {
			if httpLatencyHistogram != nil {
				duration := time.Since(start).Seconds()
				httpLatencyHistogram.Record(req.Context(), duration, metric.WithAttributes(attribute.String("http.target", targetPath)))
			}
		}()

		if httpRequestsCounter != nil {
			httpRequestsCounter.Add(req.Context(), 1, metric.WithAttributes(attribute.String("http.target", "/info")))
		}

		EmitOtelLog(req_ctx, 
					logProvider, 
					"handled /info request", 
					sdklog.SeverityInfo,
					attribute.String("http.method", "GET"),
					attribute.String("http.route", "/info"),
		)
	
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(appServer)
	})
	
	addTransactionLimit := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addTransactionLimit.HandleFunc("/checkLimitTransaction", core_middleware.MiddleWareErrorHandler(httpRouters.CheckLimitTransaction))		
	addTransactionLimit.Use(otelmux.Middleware("go-limit"))

	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port", strconv.Itoa(h.httpServer.Port)).Send()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("canceling http mux server !!!")
		}
	}()

	// Get SIGNALS
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGHUP:
			childLogger.Info().Msg("Received SIGHUP: reloading configuration...")
		case syscall.SIGINT, syscall.SIGTERM:
			childLogger.Info().Msg("Received SIGINT/SIGTERM termination signal. Exiting")
			return
		default:
			childLogger.Info().Interface("Received signal:", sig).Send()
		}
	}

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("warning dirty shutdown !!!")
		return
	}
}