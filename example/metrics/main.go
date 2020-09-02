package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/armon/go-metrics"

	spec "github.com/armory-io/go-spec"
)

func main() {
	// creating a cancelable, top-level context allows us to shutdown
	// any components using this shared context
	ctx, cancelFunc := context.WithCancel(context.Background())

	// instantiate new metrics server
	metricsConfig := spec.MetricsServerConfig{
		ServiceName: "",
		Ctx:         ctx,
	}
	ms, err := spec.NewDefaultMetricsServer(metricsConfig)
	if err != nil {
		log.Fatalf("failed creating default metrics server: %s", err.Error())
	}

	// grab a handle to the common metrics registry. adding any extra metrics you like
	// to this registry
	registry := ms.MetricsRegistry()

	// setup routes
	router := mux.NewRouter()
	router.HandleFunc("/canals", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		start := time.Now()
		// measure request latency
		defer registry.MeasureSinceWithLabels([]string{"canalsEndpoint.latency"}, start, []metrics.Label{{Name: "endpoint", Value: "canals"}})
		w.Write([]byte(`["panama", "suez"]`))
	})

	router.HandleFunc("/canals/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// instrument request handlers with common http request metrics
	router.Use(ms.InstrumentMuxRouter)

	// create a new http server, you could also use the server implementation from
	// go-yaml-tools/server as well
	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	// watchForInterrupt watches for kill signals and cancels the
	// shared context, causing any consumers of that context
	// to close gracefully. in this example, the metrics server
	// and main application server are both using this context
	go watchForInterrupt(cancelFunc)

	// start the metrics server. by default, this will serve
	// metrics on port 3001 at `/armory-observability/metrics`
	go func() {
		if err := ms.Start(); err != http.ErrServerClosed {
			log.Fatalf("metrics server quit unexpectedly: %s", err.Error())
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("shutting down main application server")
				server.Shutdown(context.Background())
				return
			}
		}
	}()

	log.Println("starting application server on port 3000")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("application quit unexpectedly: %s", err.Error())
	}
	log.Println("application quit gracefully")
}

func watchForInterrupt(cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-c:
			cancelFunc()
			return
		}
	}
}
