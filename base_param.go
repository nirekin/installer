package installer

import (
	"fmt"

	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
)

func saveBaseParams(bp ansible.BaseParam, c *InstallerContext, dest *util.FolderPath, sr *stepResult) bool {
	b, e := bp.Content()
	if e != nil {
		FailsOnCode(sr, e, fmt.Sprintf("An error occured creating the base parameters"), nil)
		return true
	}
	_, e = util.SaveFile(c.log, *dest, util.ParamYamlFileName, b)
	if e != nil {
		FailsOnCode(sr, e, fmt.Sprintf("An error occured saving the parameter file into :%v", dest.Path()), nil)
		return true
	}
	return false
}
