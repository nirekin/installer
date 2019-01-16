package installer

import (
	"log"
	"testing"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/model"
	"github.com/stretchr/testify/assert"
)

func TestBaseParamFromContext(t *testing.T) {
	c := InstallerContext{
		sshPublicKeyContent:  "sshPublicKey_content",
		sshPrivateKeyContent: "sshPrivateKey_content",
		engine: engine.EkaraMock{
			Env: model.Environment{
				Name:      "NameContent",
				Qualifier: "QualifierContent",
			},
		},
	}
	bp := engine.BuildBaseParam(c, "nodeId", "providerName")
	assert.NotNil(t, bp)
	assert.Equal(t, 2, len(bp.Body))

	val, ok := bp.Body["environment"]
	assert.True(t, ok)
	mSi, ok := val.(map[string]interface{})
	assert.True(t, ok)

	log.Printf("--> MSI params : %v", mSi)

	assert.Equal(t, 4, len(mSi))
	assert.Equal(t, "NameContent_QualifierContent_nodeId", mSi["id"])
	assert.Equal(t, "NameContent", mSi["name"])
	assert.Equal(t, "QualifierContent", mSi["qualifier"])
	assert.Equal(t, "nodeId", mSi["nodeset"])

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
		engine: engine.EkaraMock{
			Env: model.Environment{},
		},
	}
	bp := engine.BuildBaseParam(c, "nodeId", "providerName")
	assert.NotNil(t, bp)

	assert.Equal(t, 2, len(bp.Body))

	val, ok := bp.Body["environment"]
	assert.True(t, ok)
	mSi, ok := val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(mSi))
	assert.Equal(t, "nodeId", mSi["nodeset"])

	val, ok = bp.Body["connectionConfig"]
	assert.True(t, ok)
	mSi, ok = val.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(mSi))
	assert.Equal(t, "providerName", mSi["provider"])
}

func TestBaseEmptyParamFromContext(t *testing.T) {
	c := InstallerContext{
		engine: engine.EkaraMock{
			Env: model.Environment{},
		},
	}
	bp := engine.BuildBaseParam(c, "", "")
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
