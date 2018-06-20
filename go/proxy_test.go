package main

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoProxy(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.HttpProxyEnvVariableKey)
	os.Unsetenv(engine.HttpsProxyEnvVariableKey)
	os.Unsetenv(engine.NoProxyEnvVariableKey)

	e, _ := fproxy()
	assert.Nil(t, e)
	assert.Equal(t, "", httpProxy)
	assert.Equal(t, "", httpsProxy)
	assert.Equal(t, "", noProxy)
}

func TestProxy(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(engine.HttpProxyEnvVariableKey, "http_value")
	os.Setenv(engine.HttpsProxyEnvVariableKey, "https_value")
	os.Setenv(engine.NoProxyEnvVariableKey, "no_value")
	e, _ := fproxy()
	assert.Nil(t, e)
	assert.Equal(t, "http_value", httpProxy)
	assert.Equal(t, "https_value", httpsProxy)
	assert.Equal(t, "no_value", noProxy)
}
