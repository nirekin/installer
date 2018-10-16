package installer

import (
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/util"
	"github.com/stretchr/testify/assert"
)

func TestNoProxy(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(util.HttpProxyEnvVariableKey)
	os.Unsetenv(util.HttpsProxyEnvVariableKey)
	os.Unsetenv(util.NoProxyEnvVariableKey)

	sc := fproxy(c)
	e := sc.Contexts[0].Error
	assert.Nil(t, e)
	assert.Equal(t, "", c.httpProxy)
	assert.Equal(t, "", c.httpsProxy)
	assert.Equal(t, "", c.noProxy)
}

func TestProxy(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(util.HttpProxyEnvVariableKey, "http_value")
	os.Setenv(util.HttpsProxyEnvVariableKey, "https_value")
	os.Setenv(util.NoProxyEnvVariableKey, "no_value")
	sc := fproxy(c)
	e := sc.Contexts[0].Error
	assert.Nil(t, e)
	assert.Equal(t, "http_value", c.httpProxy)
	assert.Equal(t, "https_value", c.httpsProxy)
	assert.Equal(t, "no_value", c.noProxy)
}
