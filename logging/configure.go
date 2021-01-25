package logging

import (
	"fmt"
	"os"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Remote RemoteLoggingConfig `json:"remote" yaml:"remote"`
	JSON   FormatJson          `json:"json" yaml:"json"`
}

type RemoteLoggingConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Version    string `json:"version" yaml:"version"`
	CustomerID string `json:"customerId" yaml:"customerId"`
}

type FormatJson struct {
	Enabled bool              `json:"enabled" yaml:"enabled"`
	Fields  map[string]string `json:"fields" yaml:"fields"`
}

func (fj *FormatJson) Configure(l *logrus.Logger) {
	// TODO - support configuration of field names
	fm := logrus.FieldMap{}
	for f, v := range fj.Fields {
		// we have to switch here because setting fm[f] is of type
		// fieldKey, not a string...
		switch f {
		case logrus.FieldKeyTime:
			fm[logrus.FieldKeyTime] = v
		case logrus.FieldKeyMsg:
			fm[logrus.FieldKeyMsg] = v
		case logrus.FieldKeyLevel:
			fm[logrus.FieldKeyLevel] = v
		}
	}
	formatter := logrus.JSONFormatter{FieldMap: fm}
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
	if !config.JSON.Enabled {
		return
	}
	config.JSON.Configure(l)
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
		// TODO - return a specific error type here so users can ignore failures making this optional?
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
