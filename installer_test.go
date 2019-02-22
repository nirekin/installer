package installer

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/util"
	"github.com/stretchr/testify/assert"
)

/*
ActionFailId ActionId 1
ActionReportId 2
ActionCreateId 3
ActionInstallId 4
ActionDeployId 5
ActionCheckId 6
ActionDumpId 7
ActionUpdateId 8
ActionDeleteId 9
ActionNilId 10

case engine.ActionCreateId.String(): 3
case engine.ActionInstallId.String(): 4
case engine.ActionDeployId.String(): 5
case engine.ActionCheckId.String(): 6
case engine.ActionDumpId.String():  7
*/
func TestNoAction(t *testing.T) {
	c := InstallerContext{}
	c.logger = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Unsetenv(util.ActionEnvVariableKey)
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), "the action \"No action specified\" is not supported by the installer")
}

func checkUnsupportedAction(t *testing.T, a engine.ActionId) {
	c := InstallerContext{}
	c.logger = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, a.String())
	e := Run(c)
	assert.NotNil(t, e)
	assert.Equal(t, e.Error(), fmt.Sprintf("the action \"%s\" is not supported by the installer", a))
}

func TestWrongActions(t *testing.T) {
	checkUnsupportedAction(t, engine.ActionFailId)
	checkUnsupportedAction(t, engine.ActionReportId)
	checkUnsupportedAction(t, engine.ActionUpdateId)
	checkUnsupportedAction(t, engine.ActionDeleteId)
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
