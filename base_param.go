package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
)

func saveBaseParams(bp ansible.BaseParam, c *InstallerContext, dest *util.FolderPath, sc *stepContext) bool {
	b, e := bp.Content()
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured creating the base parameters"))
		return true
	}
	_, e = util.SaveFile(c.log, *dest, util.ParamYamlFileName, b)
	if e != nil {
		InstallerFail(sc, e, fmt.Sprintf("An error occured saving the parameter file into :%v", dest.Path()))
		return true
	}
	return false
}
