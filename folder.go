package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/util"
)

func fexchangeFoldef(c *InstallerContext) stepContexts {
	sc := InitStepContext("Creating the root of the exchange folder", nil, noCleanUpRequired)
	var err error
	c.ef, err = util.CreateExchangeFolder(util.InstallerVolume, "")
	if err != nil {
		InstallerFail(&sc, fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "ekara_installer", err.Error()), "")
	}
	return sc.Array()
}
