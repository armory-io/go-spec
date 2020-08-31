package go_spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Foo string
	Bar bool
}

func TestApplicationContext_GetConfig(t *testing.T) {
	cases := map[string]struct {
		input    map[string]interface{}
		expected testConfig
	}{
		"happy path": {
			input: map[string]interface{}{
				"foo": "hello",
				"bar": true,
			},
			expected: testConfig{
				Foo: "hello",
				Bar: true,
			},
		},
	}
	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			ac := &applicationContext{config: c.input}
			var target testConfig
			if err := ac.GetConfig(&target); err != nil {
				t.Fatalf("failed to convert config: %s", err.Error())
			}
			assert.EqualValues(t, c.expected, target)

		})
	}
}
