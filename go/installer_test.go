package main

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoAction(t *testing.T) {
	c := &installerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Unsetenv(engine.ActionEnvVariableKey)
	e := run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"No action specified\" is not supported by the installer")
}

func TestWrongActionUpdate(t *testing.T) {
	c := &installerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionUpdate.String())
	e := run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"1\" is not supported by the installer")
}

func TestWrongActionDelete(t *testing.T) {
	c := &installerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionDelete.String())
	e := run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"3\" is not supported by the installer")
}
