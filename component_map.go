package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func saveComponentMap(c *InstallerContext, dest *engine.FolderPath, sc *stepContext) bool {
	e := c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *dest)
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured saving the components file into :%v", dest.Path()))
		return true
	}
	return false
}
