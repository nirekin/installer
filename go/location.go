package main

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

func flocation(c *installerContext) (error, cleanup) {
	c.location = os.Getenv(engine.StarterEnvVariableKey)
	if c.location == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey), nil
	}
	return nil, noCleanUpRequired
}
