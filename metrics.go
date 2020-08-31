package go_spec

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
)

var (
	DefaultObservabilityPath = "/armory-observability/metrics"
	DefaultObservabilityAddr = ":3001"
)

type MetricsServerConfig struct {
	ServiceName string
	Addr        string
	Path        string
	Ctx         context.Context
}

type MetricsServer struct {
	metrics *metrics.Metrics
	server  *http.Server
	ctx     context.Context
}

func NewDefaultMetricsServer(cfg MetricsServerConfig) (*MetricsServer, error) {
	sink, err := prometheus.NewPrometheusSink()
	if err != nil {
		return nil, err
	}
	m, err := metrics.New(metrics.DefaultConfig(cfg.ServiceName), sink)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}
	ctx := cfg.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	ms := &MetricsServer{metrics: m, server: server, ctx: ctx}
	return ms, nil
}

// MetricsRegistry returns the metrics collector used by the metrics server
func (mc *MetricsServer) MetricsRegistry() *metrics.Metrics {
	return mc.metrics
}

func (mc *MetricsServer) WatchForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-signalChan:
			mc.server.Shutdown(mc.ctx)
			return
		case <-mc.ctx.Done():
			mc.server.Shutdown(mc.ctx)
			return
		}
	}
}

func (ms *MetricsServer) Start() error {
	return ms.server.ListenAndServe()
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (wrw wrappedResponseWriter) WriteHeader(code int) {
	wrw.statusCode = code
	wrw.ResponseWriter.WriteHeader(code)
}

// RequestMetricsMiddleware instruments incoming http requests
func (ms *MetricsServer) RequestMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := wrappedResponseWriter{w, http.StatusOK}
		next.ServeHTTP(wrappedWriter, r)
		requestLabels := requestMetricLabels(wrappedWriter, r)
		ms.metrics.IncrCounterWithLabels([]string{"http.server.requests"}, 1, requestLabels)
	})
}

func requestMetricLabels(w wrappedResponseWriter, r *http.Request) []metrics.Label {
	// TODO: add the rest of the required labels
	method := r.Method
	uri := r.URL.RequestURI()
	code := w.statusCode
	outcome := codeToOutcome(code)
	return []metrics.Label{
		{Name: "method", Value: method},
		{Name: "uri", Value: uri},
		{Name: "outcome", Value: outcome},
		{Name: "status", Value: strconv.Itoa(code)},
	}
}

func codeToOutcome(code int) string {
	switch {
	case code >= 100 && code <= 199:
		return "INFORMATIONAL"
	case code >= 200 && code <= 299:
		return "SUCCESS"
	case code >= 300 && code <= 399:
		return "REDIRECTION"
	case code >= 400 && code <= 499:
		return "CLIENT_ERROR"
	case code >= 500 && code <= 599:
		return "SERVER_ERROR"
	}
	return "UNKNOWN_STATUS"
}
