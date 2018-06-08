package main

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoAction(t *testing.T) {
	e := run()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"No action specified\" is not supported by the installer")
}

func TestBadActionUpdate(t *testing.T) {
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionUpdate.String())
	e := run()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"1\" is not supported by the installer")
}

func TestBadActionDelete(t *testing.T) {
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionDelete.String())
	e := run()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"3\" is not supported by the installer")
}

func TestNoClient(t *testing.T) {
	e := fclient()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_CLIENT\" should be defined")
}

func TestNoLocation(t *testing.T) {
	e := flocation()
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_ENV_DESCR\" should be defined")
}

func TestClient(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ClientEnvVariableKey, "test_client")
	e := fclient()
	assert.Nil(t, e)
	assert.Equal(t, client, "test_client")
}

func TestLocation(t *testing.T) {
	loggerLog = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.StarterEnvVariableKey, "test_location")
	e := flocation()
	assert.Nil(t, e)
	assert.Equal(t, location, "test_location")
}
