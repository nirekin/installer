package installer

import (
	"fmt"

	"github.com/lagoon-platform/engine"
)

func saveBaseParams(bp engine.BaseParam, c *InstallerContext, dest *engine.FolderPath, sc *stepContext) bool {
	b, e := bp.Content()
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured creating the base parameters"))
		return true
	}
	_, e = engine.SaveFile(c.log, *dest, engine.ParamYamlFileName, b)
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured saving the parameter file into :%v", dest.Path()))
		return true
	}
	return false
}
