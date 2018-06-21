package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
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
func enrichExchangeFolder(f engine.ExchangeFolder, e model.Environment, c *InstallerContext) (r map[string]engine.ExchangeFolder, err error) {
	r = make(map[string]engine.ExchangeFolder)
	for _, n := range c.lagoon.Environment().NodeSets {
		p := n.Provider.ProviderName()
		if !f.Input.Contains(p) {
			var pEf engine.ExchangeFolder
			pEf, err = engine.CreateExchangeFolder(f.Input.Path(), p)
			if err != nil {
				return
			}
			err = pEf.Create()
			if err != nil {
				return
			}
			r[p] = pEf
		}
	}
	return
}
