package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
)

func fcliparam(c *InstallerContext) stepResults {
	sc := InitParameterStepResult("Reading substitution parameters", nil, noCleanUpRequired)
	ok := c.ef.Location.Contains(util.CliParametersFileName)
	if ok {
		p, e := ansible.ParseParams(util.JoinPaths(c.ef.Location.Path(), util.CliParametersFileName))
		if e != nil {
			FailsOnCode(&sc, e, fmt.Sprintf(ERROR_LOADING_CLI_PARAMETERS, e), nil)
			goto MoveOut
		}
		c.cliparams = p
		c.log.Printf(LOG_CLI_PARAMS, c.cliparams)
	}
MoveOut:
	return sc.Array()
}
