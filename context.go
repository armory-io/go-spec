package go_spec

import (
	"github.com/armory/go-yaml-tools/pkg/spring"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	slog "github.com/go-eden/slf4go"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// ApplicationContext provides
type ApplicationContext struct {
	Logger *slog.Logger
	router *mux.Router
	Config map[string]interface{}
	server *server.Server
}

type ServerConfig struct {
	Server server.ServerConfig `yaml:"server"`
}

func NewApplicationContext(propNames []string) (*ApplicationContext, error) {
	ac := &ApplicationContext{}
	// load the configuration
	cfg, err := loadConfig(propNames)
	if err != nil {
		return nil, err
	}
	ac.Config = cfg

	// TODO: setup logging based on the application configuration
	logger := slog.NewLogger("application")
	ac.Logger = logger

	ac.router = mux.NewRouter()

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

func (ac *ApplicationContext) GetConfig(dest interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           dest,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(ac.Config); err != nil {
		return err
	}

	return nil
}

func (ac *ApplicationContext) GetRouter() (*mux.Router, error) {
	return ac.router, nil
}

// Start starts the ApplicationContext's web server and starts listening
// on the configured port
func (ac *ApplicationContext) Start(router *mux.Router) error {
	var r *mux.Router
	if router == nil {
		r = ac.router
	}

	return ac.server.Start(r)
}
