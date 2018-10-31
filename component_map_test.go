package installer

import (
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine/util"
	"github.com/ekara-platform/model"
	"github.com/stretchr/testify/assert"
)

func TestSaveComponentMapOk(t *testing.T) {

	ef, e := util.CreateExchangeFolder("./", "testFolder")
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
		ekara: EkaraMock{
			Env: model.Environment{
				Name:      "NameContent",
				Qualifier: "QualifierContent",
			},
		},
	}

	ko := saveComponentMap(&c, ef.Input, &sc)
	assert.False(t, ko)
	assert.Equal(t, sc.ErrorDetail, "")
	assert.Nil(t, sc.Error)
	assert.Nil(t, sc.ErrorOrigin)

	ok := ef.Input.Contains(util.ComponentPathsFileName)
	assert.True(t, ok)

}