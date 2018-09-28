package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func fexchangeFoldef(c *InstallerContext) stepContexts {
	sc := InitStepContext("Creating the root of the exchange folder", nil, noCleanUpRequired)
	var err error
	c.ef, err = engine.CreateExchangeFolder(engine.InstallerVolume, "")
	if err != nil {
		InstallerFail(&sc, fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "lagoon_installer", err.Error()), "")
	}
	return sc.Array()
}
