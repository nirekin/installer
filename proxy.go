package installer

import (
	"github.com/lagoon-platform/engine"
)

// fproxy loads the proxy settings form the environmant variables into the
// context
func fproxy(c *InstallerContext) (error, cleanup) {
	// We check if the proxy is well defined, the proxy is required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	c.httpProxy, c.httpsProxy, c.noProxy = engine.CheckProxy()
	return nil, noCleanUpRequired
}
