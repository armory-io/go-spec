package go_spec

import (
	"context"
	"net/http"

	"github.com/armory/go-yaml-tools/pkg/spring"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	slog "github.com/go-eden/slf4go"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type applicationContext struct {
	logger *slog.Logger
	router *mux.Router
	config map[string]interface{}
	server *server.Server
	ms     *MetricsServer
}

// ApplicationContextConfig is used to supply the ApplicationContext
// with properties about the application. Values supplied here
// influence the way the application is bootstrapped.
type ApplicationContextConfig struct {
	// Name denotes the application's name
	Name string

	// ConfigNames are used by the ApplicationContext to determine
	// which config profiles to load
	ConfigNames []string

	// Ctx allows users to supply a primary context that will
	// be used by all components provided by the ApplicationContext
	Ctx context.Context
}

// ServerConfig is used to extract configuration information
// about how the webserver should be configured
type ServerConfig struct {
	Server server.ServerConfig `yaml:"server"`
}

// NewApplicationContext provides all of the common utilities needed to make
// building an observable application simple. Using the ApplicationContext's
// logger, router & server will ensure that the application is instrumented
// in a common way
func NewApplicationContext(acc ApplicationContextConfig) (*applicationContext, error) {

	// load the configuration
	cfg, err := loadConfig(acc.ConfigNames)
	if err != nil {
		return nil, err
	}

	// TODO: setup logging based on the application configuration
	logger := slog.NewLogger(acc.Name)

	ctx := acc.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	msc := MetricsServerConfig{
		ServiceName: acc.Name,
		Addr:        DefaultObservabilityAddr,
		Path:        DefaultObservabilityPath,
		Ctx:         ctx,
	}
	ms, _ := NewDefaultMetricsServer(msc)

	ac := &applicationContext{
		router: mux.NewRouter(),
		logger: logger,
		ms:     ms,
		config: cfg,
	}

	// use configuration to setup server with TLS if necessary
	var sc ServerConfig
	if err := ac.GetConfig(&sc); err != nil {
		return nil, err
	}
	ac.server = server.NewServer(&sc.Server)

	return ac, nil
}

func loadConfig(propNames []string) (map[string]interface{}, error) {
	return spring.LoadDefault(propNames)
}

// GetConfig deserializes the applications raw configuration into
// a structured type defined by the caller. This is useful if you'd
// like strongly typed configuration vs map[string]interface{}
func (ac *applicationContext) GetConfig(dest interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           dest,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(ac.config); err != nil {
		return err
	}

	return nil
}

// GetRouter returns the ApplicationContext's router
func (ac *applicationContext) GetRouter() (*mux.Router, error) {
	return ac.router, nil
}

// Logger returns the ApplicationContext's logger
func (ac *applicationContext) Logger() *slog.Logger {
	return ac.logger
}

// Start starts the ApplicationContext's web server and starts listening
// on the configured port
func (ac *applicationContext) Start(router *mux.Router) error {
	var r http.Handler
	if router == nil {
		r = ac.router
	}
	// instrument http requests
	r = ac.ms.RequestMetricsMiddleware(r)
	return ac.server.Start(r)
}

// CollectMetrics starts the ApplicationContext's metrics server
func (ac *applicationContext) CollectMetrics() error {
	go ac.ms.WatchForShutdown()
	return ac.ms.Start()
}
