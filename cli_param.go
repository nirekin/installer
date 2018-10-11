package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine/ansible"
	"github.com/lagoon-platform/engine/util"
)

func fcliparam(c *InstallerContext) stepContexts {
	sc := InitStepContext("Reading substitution parameters", nil, noCleanUpRequired)
	ok := c.ef.Location.Contains(util.CliParametersFileName)
	if ok {
		p, e := ansible.ParseParams(util.JoinPaths(c.ef.Location.Path(), util.CliParametersFileName))
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
