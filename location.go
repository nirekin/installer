package installer

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

// flocation extracts the descriptor location and descriptor  file name from the
// environment variables "engine.StarterEnvVariableKey" and
// engine.StarterEnvNameVariableKey
func flocation(c *InstallerContext) stepContexts {
	sc := InitStepContext("Reading the descriptor location", nil, noCleanUpRequired)
	c.location = os.Getenv(engine.StarterEnvVariableKey)
	if c.location == "" {
		InstallerFail(&sc, fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey), "")
		goto MoveOut
	}
	c.name = os.Getenv(engine.StarterEnvNameVariableKey)
	if c.name == "" {
		InstallerFail(&sc, fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvNameVariableKey), "")
		goto MoveOut
	}
MoveOut:
	return sc.Array()
}
