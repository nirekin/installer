package main

import (
	"github.com/lagoon-platform/engine"
)

func fproxy() (error, cleanup) {
	// We check if the proxy is well defined, the proxy is required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	httpProxy, httpsProxy, noProxy = engine.CheckProxy()
	return nil, noCleanUpRequired
}
