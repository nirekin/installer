package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/util"
)

func fexchangeFoldef(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Creating the root of the exchange folder", nil, noCleanUpRequired)
	var err error
	c.ef, err = util.CreateExchangeFolder(util.InstallerVolume, "")
	if err != nil {
		FailsOnCode(&sc, err, fmt.Sprintf(ERROR_CREATING_EXCHANGE_FOLDER, c.qualifiedName, err.Error()), nil)
	}
	return sc.Array()
}
