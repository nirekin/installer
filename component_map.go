package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/util"
)

func saveComponentMap(c *InstallerContext, dest *util.FolderPath, sr *stepResult) bool {
	e := c.ekara.ComponentManager().SaveComponentsPaths(c.log, *dest)
	if e != nil {
		FailsOnCode(sr, e, fmt.Sprintf("An error occured saving the components file into :%v", dest.Path()), nil)
		return true
	}
	return false
}
