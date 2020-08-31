package main

import (
	"fmt"
	"net/http"
	"os"

	spec "github.com/armory-io/go-spec"
)

func main() {
	// Apply any application specific configuration
	appConfig := spec.ApplicationContextConfig{
		ConfigNames: []string{"spinnaker", "basicapp"},
	}

	// By creating an ApplicationContext, we can get access
	// to all of the pre-configured items we need for operations
	appContext, err := spec.NewApplicationContext(appConfig)
	if err != nil {
		fmt.Printf("failed to create a new application context: %s", err.Error())
		os.Exit(1)
	}

	// logger is a standardized logger
	logger := appContext.Logger()

	go func() {
		if err := appContext.CollectMetrics(); err != nil {
			logger.Fatalf("metrics collection failed: %s", err.Error())
		}
	}()

	// using the ApplicationContext's router ensures that our routes are logged
	// and instrumented properly
	router, _ := appContext.GetRouter()
	router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world"))
	})

	logger.Infof("starting basic application")
	if err := appContext.Start(nil); err != http.ErrServerClosed {
		logger.Fatalf("application failed unexpectedly: %s", err.Error())
	}
	logger.Infof("basic application exiting")
}
