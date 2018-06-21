package installer

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

func flocation(c *InstallerContext) (error, cleanup) {
	c.location = os.Getenv(engine.StarterEnvVariableKey)
	if c.location == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey), nil
	}
	return nil, noCleanUpRequired
}
