package main

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

func fclient(c *installerContext) (error, cleanup) {
	c.client = os.Getenv(engine.ClientEnvVariableKey)
	if c.client == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.ClientEnvVariableKey), noCleanUpRequired
	}
	if c.log != nil {
		c.log.Printf(LOG_CREATION_FOR_CLIENT, c.client)
	}
	return nil, noCleanUpRequired
}
