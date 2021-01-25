package main

import (
	"github.com/armory-io/go-spec/logging"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var jsonFormatting = `
logging:
  json:
    enabled: true
waterways:
  - tennessee river
  - mississippi river`

// example configuration for enabling json logging
// with field overrides
var jsonFormattingWithFieldOverrides = `
logging:
  json:
    enabled: true
	fields:
      msg: "message"
      time: "@timestamp"
waterways:
  - tennessee river
  - mississippi river`

// appConfig defines a struct that we can use to configure out application
type appConfig struct {
	Logging   logging.Config `yaml:"logging"`
	Canals    []string       `yaml:"canals"`
	Waterways []string       `yaml:"waterways"`
}

func main() {

	var config appConfig
	// we're reading some yaml from a variable here but, in practice, you will
	// most likely be reading the content from a file. the variable is used to
	// to keep this example simple.
	if err := yaml.Unmarshal([]byte(jsonFormatting), &config); err != nil {
		panic(err)
	}

	// here we create a logger and pass it into our logging package
	// which will configure it with the appropriate formatting
	// and hooks for remote logging
	logger := logrus.New()
	if err := logging.ConfigureLogrus(logger, config.Logging); err != nil {
		panic(err)
	}

	logger.Infof("configured waterways: %v", config.Waterways)
	logger.Infof("configured canals: %v", config.Canals)
}
