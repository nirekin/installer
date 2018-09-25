package installer

import (
	"log"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
	"github.com/stretchr/testify/assert"
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
	return nil
}

func (m CMMock) Ensure() error {
	return nil
}

func TestBaseParamFromContext(t *testing.T) {
	c := InstallerContext{
		sshPublicKey:  "sshPublicKey_content",
		sshPrivateKey: "sshPrivateKey_content",
		lagoon: LaggonMock{
			Env: model.Environment{
				Name:      "NameContent",
				Qualifier: "QualifierContent",
			},
		},
	}
	bp := c.BuildBaseParam("nodeId", "providerName")
	assert.NotNil(t, bp)

	assert.Equal(t, 2, len(bp.Body))
	log.Printf("BP %v \n", bp)
	val, ok := bp.Body["environment"]
	assert.True(t, ok)
	mSi, ok := val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(mSi))
	assert.Equal(t, "NameContent_QualifierContent", mSi["name"])
	assert.Equal(t, "nodeId", mSi["uid"])

	val, ok = bp.Body["connectionConfig"]
	assert.True(t, ok)
	mSi, ok = val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(mSi))
	assert.Equal(t, "providerName", mSi["provider"])
	assert.Equal(t, "sshPublicKey_content", mSi["machine_public_key"])
	assert.Equal(t, "sshPrivateKey_content", mSi["machine_private_key"])
}

func TestBaseAlmostEmptyParamFromContext(t *testing.T) {
	c := InstallerContext{
		lagoon: LaggonMock{
			Env: model.Environment{},
		},
	}
	bp := c.BuildBaseParam("nodeId", "providerName")
	assert.NotNil(t, bp)

	assert.Equal(t, 2, len(bp.Body))

	val, ok := bp.Body["environment"]
	assert.True(t, ok)
	mSi, ok := val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(mSi))
	assert.Equal(t, "nodeId", mSi["uid"])

	val, ok = bp.Body["connectionConfig"]
	assert.True(t, ok)
	mSi, ok = val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(mSi))
	assert.Equal(t, "providerName", mSi["provider"])
}

func TestBaseEmptyParamFromContext(t *testing.T) {
	c := InstallerContext{
		lagoon: LaggonMock{
			Env: model.Environment{},
		},
	}
	bp := c.BuildBaseParam("", "")
	assert.NotNil(t, bp)

	assert.Equal(t, 2, len(bp.Body))

	val, ok := bp.Body["environment"]
	assert.True(t, ok)
	mSi, ok := val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(mSi))

	val, ok = bp.Body["connectionConfig"]
	assert.True(t, ok)
	mSi, ok = val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(mSi))
}
