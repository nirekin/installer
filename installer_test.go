package installer

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoAction(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Unsetenv(engine.ActionEnvVariableKey)
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"No action specified\" is not supported by the installer")
}

func TestWrongActionUpdate(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionUpdate.String())
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"1\" is not supported by the installer")
}

func TestWrongActionDelete(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	c.log.Println()
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionDelete.String())
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"3\" is not supported by the installer")
}
