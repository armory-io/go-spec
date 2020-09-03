package go_spec

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
)

var (
	DefaultObservabilityPath = "/armory-observability/metrics"
	DefaultObservabilityAddr = ":3001"
)

type MetricsServerConfig struct {
	ServiceName   string
	Addr          string
	Path          string
	Ctx           context.Context
	DefaultLabels []string
}

type MetricsServer struct {
	metrics       *metrics.Metrics
	server        *http.Server
	ctx           context.Context
	defaultLabels []metrics.Label
}

func NewDefaultMetricsServer(cfg MetricsServerConfig) (*MetricsServer, error) {
	if cfg.ServiceName == "" {
		return nil, errors.New("metrics server requires an application name be provided by configuration")
	}

	sink, err := prometheus.NewPrometheusSink()
	if err != nil {
		return nil, err
	}
	m, err := metrics.New(metrics.DefaultConfig(""), sink)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	pth := cfg.Path
	if pth == "" {
		pth = DefaultObservabilityPath
	}
	mux.Handle(pth, promhttp.Handler())

	addr := cfg.Addr
	if addr == "" {
		addr = DefaultObservabilityAddr
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	ctx := cfg.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	defaultLabels := labelsFromPairs(append(cfg.DefaultLabels, "appName", cfg.ServiceName))
	ms := &MetricsServer{
		metrics:       m,
		server:        server,
		ctx:           ctx,
		defaultLabels: defaultLabels,
	}
	return ms, nil
}

func labelsFromPairs(a []string) []metrics.Label {
	labels := []metrics.Label{}
	for i := 0; i < len(a); i = i + 2 {
		labels = append(labels, metrics.Label{Name: a[i], Value: a[i+1]})
	}
	return labels
}

// MetricsRegistry returns the metrics collector used by the metrics server
func (ms *MetricsServer) MetricsRegistry() *metrics.Metrics {
	return ms.metrics
}

func (ms *MetricsServer) WatchForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-signalChan:
			ms.server.Shutdown(ms.ctx)
			return
		case <-ms.ctx.Done():
			ms.server.Shutdown(ms.ctx)
			return
		}
	}
}

func (ms *MetricsServer) Start() error {
	go func() {
		for {
			select {
			case <-ms.ctx.Done():
				cancelContext, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
				defer func() {
					cancelFunc()
				}()
				ms.server.Shutdown(cancelContext)
			}
		}
	}()

	return ms.server.ListenAndServe()
}

// wrappedResponseWriter is used to capture the status code
// response so we can use it in metrics
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (wrw wrappedResponseWriter) WriteHeader(code int) {
	wrw.statusCode = code
	wrw.ResponseWriter.WriteHeader(code)
}

// URIMapperFunc is used by RequestMetricsMiddleware
// to map requests URIs to their parametrized counterpart
type URIMapperFunc func(r *http.Request) string

var MuxURIMapperFunc = func(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return DefaultURIMapperFunc(r)
	}
	// TODO: should we handle the error or not?
	t, _ := route.GetPathTemplate()
	return t
}

// DefaultURIMapperFunc returns the RequestURI from the URL
var DefaultURIMapperFunc = func(r *http.Request) string {
	return r.URL.RequestURI()
}

// RequestMetricsMiddleware instruments incoming http requests using Go's
// default ServerMux
func (ms *MetricsServer) RequestMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := wrappedResponseWriter{w, http.StatusOK}
		startTime := time.Now()
		next.ServeHTTP(wrappedWriter, r)
		requestLabels := requestMetricLabels(wrappedWriter, r, DefaultURIMapperFunc)
		ms.metrics.MeasureSinceWithLabels([]string{"http.server.requests"}, startTime, requestLabels)
	})
}

// InstrumentMuxRouter should be paired with a router from gorilla/mux
func (ms *MetricsServer) InstrumentMuxRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := wrappedResponseWriter{w, http.StatusOK}
		startTime := time.Now()
		next.ServeHTTP(wrappedWriter, r)
		requestLabels := requestMetricLabels(wrappedWriter, r, MuxURIMapperFunc)
		labels := append(requestLabels, ms.defaultLabels...)
		ms.metrics.MeasureSinceWithLabels([]string{"http.server.requests"}, startTime, labels)
	})
}

func requestMetricLabels(w wrappedResponseWriter, r *http.Request, uriMapper func(r *http.Request) string) []metrics.Label {
	// TODO: add the rest of the required labels
	method := r.Method
	uri := uriMapper(r)
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
