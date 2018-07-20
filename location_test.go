package installer

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoLocation(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.StarterEnvVariableKey)
	os.Setenv(engine.StarterEnvNameVariableKey, "test_name")
	e, _ := flocation(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_ENV_DESCR\" should be defined")
}

func TestNoName(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.StarterEnvNameVariableKey)
	os.Setenv(engine.StarterEnvVariableKey, "test_location")
	e, _ := flocation(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the environment variable \"LAGOON_ENV_DESCR_NAME\" should be defined")
}

func TestLocation(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(engine.StarterEnvVariableKey, "test_location")
	os.Setenv(engine.StarterEnvNameVariableKey, "test_name")
	e, _ := flocation(c)
	assert.Nil(t, e)
	assert.Equal(t, c.location, "test_location")
	assert.Equal(t, c.name, "test_name")
}
