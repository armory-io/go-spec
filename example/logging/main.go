package main

import (
	"github.com/armory-io/go-spec/logging"
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
    level: debug
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

type appComponent struct {
	l logging.LeveledLogger
}

func (ac *appComponent) doTask(id string) {
	ac.l.Infof("starting task with id %s", id)
}

func main() {

	var config appConfig
	// we're reading some yaml from a variable here but, in practice, you will
	// most likely be reading the content from a file. the variable is used to
	// to keep this example simple.
	if err := yaml.Unmarshal([]byte(jsonFormattingWithFieldOverrides), &config); err != nil {
		panic(err)
	}

	// here we will call our logging package to give us a logger
	// under the hood, logging will create a logger configured
	// exactly the way we want it, eliminating the need for us to
	// do any setup
	logger, err := logging.NewLeveledLogger(config.Logging)

	// used to demonstrate how easy we can swap implementation
	// logger, err := logging.NewZapLeveledLogger()
	if err != nil {
		panic(err)
	}

	logger.Infof("configured waterways: %v", config.Waterways)
	logger.Infof("configured canals: %v", config.Canals)

	// log levels can be customized by setting the level option in the
	// logging format config
	logger.Debugf("this is a debug log")

	// adding contextual fields to your logger is as simple as calling WithField or WithFields
	logger.WithField("context-field", "context-value").Info("i have an extra field")

	// example of passing logger into another application component
	// in this case, appComponent depends on our logging interface
	// rather than a specific implementation like logrus or zap
	ac := &appComponent{l: logger}
	ac.doTask("some-random-task-id")

	// in some cases, like unit testing, you don't actually want
	// to pass a real logger to a component. for that, we provide
	// a NoopLogger that you can use in it's place
	acWithNoop := &appComponent{l: &logging.NoopLeveledLogger{}}
	// notice has calls to doTask don't output any logging in the console
	acWithNoop.doTask("another-random-task-id")
}
