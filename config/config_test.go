package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTransform(t *testing.T) {
	cfg := map[string]interface{}{
		"foo": "bar",
		"key": 5,
	}

	d := &struct {
		Foo string
		Key int
	}{}

	err := Transform(cfg, d)

	assert.NoError(t, err)
	assert.Equal(t, "bar", d.Foo)
	assert.Equal(t, 5, d.Key)

	err = Transform(make(chan int), d)
	assert.Error(t, err)
}

func TestGetLogger(t *testing.T) {
	rawJSON := []byte(`{"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}`)
	zCfg := &zap.Config{}
	json.Unmarshal(rawJSON, zCfg)

	assert.NotNil(t, getLogger(zCfg))
	assert.Nil(t, getLogger(nil))
}

func TestGetDefaultLogger(t *testing.T) {
	assert.NotNil(t, getDefaultLogger())
}
