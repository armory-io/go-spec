package logging

import (
	"fmt"
	"os"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/sirupsen/logrus"
)

func NewLeveledLogger(c Config) (LeveledLogger, error) {
	return makeAndConfigure(c)
}

func makeAndConfigure(c Config) (*LogrusAdapter, error) {
	l := logrus.New()
	if err := configureLogrus(l, c); err != nil {
		return nil, err
	}
	return &LogrusAdapter{logrus.NewEntry(l)}, nil
}

func configureLogrus(l *logrus.Logger, config Config) error {
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
		return fmt.Errorf("failed to instantiate remote log formatter: %w", err)
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
