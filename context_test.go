package installer

import (
	"testing"

	"github.com/lagoon-platform/model"
	"github.com/stretchr/testify/assert"
)

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
