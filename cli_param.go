package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func fcliparam(c *InstallerContext) stepContexts {
	sc := InitStepContext("Reading substitution parameters", nil, noCleanUpRequired)
	ok := c.ef.Location.Contains(engine.CliParametersFileName)
	if ok {
		p, e := engine.ParseParams(engine.JoinPaths(c.ef.Location.Path(), engine.CliParametersFileName))
		if e != nil {
			InstallerFail(&sc, fmt.Errorf(ERROR_LOADING_CLI_PARAMETERS, e), "")
			goto MoveOut
		}
		c.cliparams = p
		c.log.Printf(LOG_CLI_PARAMS, c.cliparams)
	}
MoveOut:
	return sc.Array()
}
