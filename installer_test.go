package installer

import (
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/util"
	"github.com/stretchr/testify/assert"
)

func TestNoAction(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Unsetenv(util.ActionEnvVariableKey)
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"No action specified\" is not supported by the installer")
}

func TestWrongActionUpdate(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, engine.ActionUpdate.String())
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"1\" is not supported by the installer")
}

func TestWrongActionDelete(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	c.log.Println()
	os.Setenv(util.ActionEnvVariableKey, engine.ActionDelete.String())
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"3\" is not supported by the installer")
}

func TestRepositoryFlavor(t *testing.T) {

	a, b := repositoryFlavor("aaa")
	assert.Equal(t, a, "aaa")
	assert.Equal(t, b, "")

	a, b = repositoryFlavor("aaa@bbb")
	assert.Equal(t, a, "aaa")
	assert.Equal(t, b, "bbb")

	a, b = repositoryFlavor("aaa@")
	assert.Equal(t, a, "aaa")
	assert.Equal(t, b, "")

	a, b = repositoryFlavor("aaa@bbb@willbeignored")
	assert.Equal(t, a, "aaa")
	assert.Equal(t, b, "bbb")
}
