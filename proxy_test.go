package installer

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestNoProxy(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Unsetenv(engine.HttpProxyEnvVariableKey)
	os.Unsetenv(engine.HttpsProxyEnvVariableKey)
	os.Unsetenv(engine.NoProxyEnvVariableKey)

	sc := fproxy(c)
	e := sc.contexts[0].Err
	assert.Nil(t, e)
	assert.Equal(t, "", c.httpProxy)
	assert.Equal(t, "", c.httpsProxy)
	assert.Equal(t, "", c.noProxy)
}

func TestProxy(t *testing.T) {
	c := &InstallerContext{}
	c.log = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(engine.ActionEnvVariableKey, engine.ActionCreate.String())
	os.Setenv(engine.HttpProxyEnvVariableKey, "http_value")
	os.Setenv(engine.HttpsProxyEnvVariableKey, "https_value")
	os.Setenv(engine.NoProxyEnvVariableKey, "no_value")
	sc := fproxy(c)
	e := sc.contexts[0].Err
	assert.Nil(t, e)
	assert.Equal(t, "http_value", c.httpProxy)
	assert.Equal(t, "https_value", c.httpsProxy)
	assert.Equal(t, "no_value", c.noProxy)
}
