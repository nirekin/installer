package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/util"
)

func saveComponentMap(c *InstallerContext, dest *util.FolderPath, sc *stepContext) bool {
	e := c.ekara.ComponentManager().SaveComponentsPaths(c.log, *dest)
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured saving the components file into :%v", dest.Path()))
		return true
	}
	return false
}
