package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func fexchangeFoldef(c *InstallerContext) (error, cleanup) {
	var err error
	c.ef, err = engine.CreateExchangeFolder(engine.InstallerVolume, "")
	if err != nil {
		return fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, engine.ClientEnvVariableKey), nil
	}
	return nil, noCleanUpRequired
}

//enrichExchangeFolder adds a sub level of ExchangeFolder to the received one.
//This will add a "sub ExchangeFolder" for each provider willing to be in charge
// of a nodeset creation.
func fenrichExchangeFolder(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		pName := n.Provider.ProviderName()
		child, err := c.ef.Input.AddChildExchangeFolder(pName)
		if err != nil {
			return err, nil
		}
		child.Create()
	}
	return nil, nil
}
