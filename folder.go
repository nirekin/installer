package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func fexchangeFoldef(c *InstallerContext) (error, cleanup) {
	var err error
	c.ef, err = engine.CreateExchangeFolder(engine.InstallerVolume, "")
	if err != nil {
		return fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, engine.ClientEnvVariableKey, err.Error()), nil
	}
	return nil, noCleanUpRequired
}
