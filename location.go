package installer

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

// flocation extracts the descriptor location and descriptor  file name from the
// environment variables "engine.StarterEnvVariableKey" and
// engine.StarterEnvNameVariableKey
func flocation(c *InstallerContext) (error, cleanup) {
	c.location = os.Getenv(engine.StarterEnvVariableKey)
	if c.location == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey), nil
	}
	c.name = os.Getenv(engine.StarterEnvNameVariableKey)
	if c.name == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvNameVariableKey), nil
	}
	return nil, noCleanUpRequired
}
