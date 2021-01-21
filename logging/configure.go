package logging

import (
	"fmt"
	"os"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Remote RemoteLoggingConfig `json:"remote"`
	Format struct {
		JSON *FormatJson `json:"json"`
	}
}

type RemoteLoggingConfig struct {
	Enabled    bool   `json:"enabled"`
	Endpoint   string `json:"endpoint"`
	Version    string `json:"version"`
	CustomerID string `json:"customerId"`
}

type FormatJson struct {
	Fields map[string]string
}

func (fj *FormatJson) Configure(l *logrus.Logger) {
	// TODO - support configuration of field names
	formatter := logrus.JSONFormatter{}
	l.SetFormatter(&formatter)
}

func ConfigureLogrus(l *logrus.Logger, config Config) error {
	// configure formatting
	configureLogrusFormat(l, config)

	// configure remote logging
	if config.Remote.Enabled {
		if err := configureRemoteLogging(l, config.Remote); err != nil {
			return err
		}
	}
	return nil
}

func configureLogrusFormat(l *logrus.Logger, config Config) {
	// we only support json as an alternative logging format
	// this will need to change if that changes in the future
	if config.Format.JSON == nil {
		return
	}
	config.Format.JSON.Configure(l)
}

func configureRemoteLogging(l *logrus.Logger, config RemoteLoggingConfig) error {
	hostname, err := resolveHostname()
	if err != nil {
		return err
	}

	formatter, err := formatters.NewHttpLogFormatter(hostname, config.CustomerID, config.Version)
	if err != nil {
		return fmt.Errorf("failed to instantiate remote log formatter: %s", err.Error())
	}

	if config.Endpoint == "" {
		return fmt.Errorf("remote log forwarding enabled but logging.remote.endpoint is unset")
	}

	l.AddHook(&hooks.HttpDebugHook{
		LogLevels: logrus.AllLevels,
		Endpoint:  config.Endpoint,
		Formatter: formatter,
	})
	return nil
}

func resolveHostname() (string, error) {
	var hostname string
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}

	if hostname == "" {
		return "", fmt.Errorf("failed to resolve hostname")
	}

	return hostname, nil
}
