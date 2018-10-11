package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine/util"
)

func saveComponentMap(c *InstallerContext, dest *util.FolderPath, sc *stepContext) bool {
	e := c.lagoon.ComponentManager().SaveComponentsPaths(c.log, *dest)
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured saving the components file into :%v", dest.Path()))
		return true
	}
	return false
}
