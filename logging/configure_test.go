package logging

import (
	"os"
	"testing"

	"github.com/armory-io/monitoring/log/formatters"

	"github.com/armory-io/monitoring/log/hooks"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"
)

func createHttpDebugHook(t *testing.T) *hooks.HttpDebugHook {
	// TODO - this method is basically copy pasta from the implementation, improve
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err.Error())
	}

	formatter, err := formatters.NewHttpLogFormatter(hostname, "12345", "12345")
	if err != nil {
		t.Fatal(err.Error())
	}

	return &hooks.HttpDebugHook{
		Endpoint:  "http://sometest.com",
		LogLevels: logrus.AllLevels,
		Formatter: formatter,
	}
}

func TestConfigureLogrus(t *testing.T) {

	// create a hook to use for this test
	httpHook := createHttpDebugHook(t)

	cases := map[string]struct {
		cfg                Config
		formatterType      logrus.Formatter
		expectedHooksTypes logrus.LevelHooks
	}{
		"default config leaves logger unmodified": {
			cfg:                Config{},
			formatterType:      &logrus.TextFormatter{},
			expectedHooksTypes: logrus.LevelHooks{},
		},
		"with json enabled": {
			cfg: Config{JSON: FormatJson{
				Enabled: true,
			}},
			formatterType:      &logrus.JSONFormatter{FieldMap: logrus.FieldMap{}},
			expectedHooksTypes: logrus.LevelHooks{},
		},
		"injects remote hook when enabled": {
			cfg: Config{Remote: RemoteLoggingConfig{
				Enabled:    true,
				Endpoint:   "http://sometest.com",
				Version:    "12345",
				CustomerID: "12345",
			}},
			formatterType:      &logrus.TextFormatter{},
			expectedHooksTypes: newLevelHooks(httpHook),
		},
		"injects field modifiers if set": {
			cfg: Config{JSON: FormatJson{
				Enabled: true,
				Fields: map[string]string{
					"time": "@timestamp",
				},
			}},
			formatterType: &logrus.JSONFormatter{
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime: "@timestamp",
				},
			},
			expectedHooksTypes: logrus.LevelHooks{},
		},
	}
	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			l := logrus.New()
			err := configureLogrus(l, c.cfg)
			if err != nil {
				t.Fatal(err.Error())
			}
			assert.IsType(t, c.formatterType, l.Formatter)
			assert.EqualValues(t, c.formatterType, l.Formatter)
			assert.EqualValues(t, c.expectedHooksTypes, l.Hooks)
		})
	}
}

func newLevelHooks(hooks ...logrus.Hook) logrus.LevelHooks {
	lh := logrus.LevelHooks{}
	for _, h := range hooks {
		lh.Add(h)
	}
	return lh
}
