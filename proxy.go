package installer

import (
	"github.com/ekara-platform/engine"
)

// fproxy loads the proxy settings form the environmant variables into the
// context
func fproxy(c *InstallerContext) stepContexts {
	sc := InitStepContext("Checking the proxy definition", nil, noCleanUpRequired)
	// We check if the proxy is well defined, the proxy can be required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	c.httpProxy, c.httpsProxy, c.noProxy = engine.CheckProxy()
	return sc.Array()
}
