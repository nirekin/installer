package main

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoClient(t *testing.T) {
	c := &installerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.ClientEnvVariableKey)
	e, _ := fclient(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_CLIENT\" should be defined")
}

func TestClient(t *testing.T) {
	c := &installerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(engine.ClientEnvVariableKey, "test_client")
	e, _ := fclient(c)
	assert.Nil(t, e)
	assert.Equal(t, c.client, "test_client")
}
