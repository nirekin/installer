package installer

import (
	"fmt"
	"os"

	"github.com/ekara-platform/engine/util"
)

// flocation extracts the descriptor location and descriptor  file name from the
// environment variables "engine.StarterEnvVariableKey" and
// engine.StarterEnvNameVariableKey
func flocation(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Reading the descriptor location", nil, noCleanUpRequired)
	c.location = os.Getenv(util.StarterEnvVariableKey)
	if c.location == "" {
		FailsOnCode(&sc, fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvVariableKey), "", nil)
		goto MoveOut
	}
	c.name = os.Getenv(util.StarterEnvNameVariableKey)
	if c.name == "" {
		FailsOnCode(&sc, fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvNameVariableKey), "", nil)
		goto MoveOut
	}
MoveOut:
	return sc.Array()
}
