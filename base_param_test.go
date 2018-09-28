package installer

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
	"github.com/stretchr/testify/assert"
)

func TestSaveBaseParamOk(t *testing.T) {

	ef, e := engine.CreateExchangeFolder("./", "testFolder")
	assert.Nil(t, e)
	assert.NotNil(t, ef)
	defer ef.Delete()

	assert.Nil(t, e)

	sc := InitStepContext("DummyStep", nil, noCleanUpRequired)
	c := InstallerContext{
		ef:            ef,
		log:           log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds),
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
	ko := saveBaseParams(bp, &c, ef.Input, &sc)
	assert.False(t, ko)
	assert.Equal(t, sc.ErrorDetail, "")
	assert.Nil(t, sc.Error)
	assert.Nil(t, sc.ErrorOrigin)

	ok, _, err := ef.Input.ContainsParamYaml()
	assert.True(t, ok)
	assert.Nil(t, err)

}
