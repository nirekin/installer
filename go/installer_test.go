package main

import (
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
