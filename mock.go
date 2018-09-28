package installer

import (
	"log"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
)

type LaggonMock struct {
	// TODO This mock should be deleted once the logic content of the installer has been
	// refactored and moved into the engine
	Env model.Environment
}

func (d LaggonMock) Init(repo string, ref string) error {
	return nil
}
func (d LaggonMock) Environment() model.Environment {
	return d.Env
}
func (d LaggonMock) ComponentManager() engine.ComponentManager {
	return CMMock{}
}

type CMMock struct {
	// TODO This mock should be deleted once the logic content of the installer has been
	// refactored and moved into the engine
}

func (m CMMock) RegisterComponent(c model.Component) {
}

func (m CMMock) ComponentPath(cId string) string {
	return ""
}

func (m CMMock) ComponentsPaths() map[string]string {
	return make(map[string]string)
}

func (m CMMock) SaveComponentsPaths(log *log.Logger, e model.Environment, dest engine.FolderPath) error {
	_, err := engine.SaveFile(log, dest, engine.ComponentPathsFileName, []byte("DUMMY_CONTENT"))
	if err != nil {
		return err
	}
	return nil
}

func (m CMMock) Ensure() error {
	return nil
}
