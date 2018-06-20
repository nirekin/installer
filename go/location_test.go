package main

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoLocation(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.StarterEnvVariableKey)
	e, _ := flocation()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_ENV_DESCR\" should be defined")
}

func TestLocation(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(engine.StarterEnvVariableKey, "test_location")
	e, _ := flocation()
	assert.Nil(t, e)
	assert.Equal(t, location, "test_location")
}
