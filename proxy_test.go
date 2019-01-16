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
	c.logger = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, engine.ActionCreateId.String())
	os.Unsetenv(util.HttpProxyEnvVariableKey)
	os.Unsetenv(util.HttpsProxyEnvVariableKey)
	os.Unsetenv(util.NoProxyEnvVariableKey)
	fillProxy(c)
	assert.Equal(t, "", c.httpProxyContent)
	assert.Equal(t, "", c.httpsProxyContent)
	assert.Equal(t, "", c.noProxyContent)
}

func TestProxy(t *testing.T) {
	c := &InstallerContext{}
	c.logger = log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	os.Setenv(util.ActionEnvVariableKey, engine.ActionCreateId.String())
	os.Setenv(util.HttpProxyEnvVariableKey, "http_value")
	os.Setenv(util.HttpsProxyEnvVariableKey, "https_value")
	os.Setenv(util.NoProxyEnvVariableKey, "no_value")
	fillProxy(c)
	assert.Equal(t, "http_value", c.httpProxyContent)
	assert.Equal(t, "https_value", c.httpsProxyContent)
	assert.Equal(t, "no_value", c.noProxyContent)
}
