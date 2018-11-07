package installer

import (
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine/util"
	"github.com/ekara-platform/model"
	"github.com/stretchr/testify/assert"
)

func TestSaveBaseParamOk(t *testing.T) {

	ef, e := util.CreateExchangeFolder("./", "testFolder")
	assert.Nil(t, e)
	assert.NotNil(t, ef)
	defer ef.Delete()

	assert.Nil(t, e)

	sc := InitCodeStepResult("DummyStep", nil, noCleanUpRequired)
	c := InstallerContext{
		ef:            ef,
		log:           log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds),
		sshPublicKey:  "sshPublicKey_content",
		sshPrivateKey: "sshPrivateKey_content",
		ekara: EkaraMock{
			Env: model.Environment{
				Name:      "NameContent",
				Qualifier: "QualifierContent",
			},
		},
	}
	bp := c.BuildBaseParam("nodeId", "providerName")
	ko := saveBaseParams(bp, &c, ef.Input, &sc)
	assert.False(t, ko)

	assert.Nil(t, sc.error)
	assert.Equal(t, sc.ErrorMessage, "")
	assert.Equal(t, string(sc.FailureCause), "")
	assert.Equal(t, string(sc.Status), string(STEP_STATUS_SUCCESS))

	ok, _, err := ef.Input.ContainsParamYaml()
	assert.True(t, ok)
	assert.Nil(t, err)
}
