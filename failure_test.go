package installer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallerFail(t *testing.T) {

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	InstallerFail(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS")

	assert.NotNil(t, sc.Error)
	assert.NotNil(t, sc.ErrorOrigin)
	assert.Equal(t, sc.Error.Error(), "DUMMY_ERROR")
	assert.Equal(t, sc.ErrorDetail, "DUMMY_DETAILS")
	assert.Equal(t, sc.ErrorOrigin.Localize(), originEkaraInstaller.Localize())

}

func TestDescriptorFail(t *testing.T) {

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	DescriptorFail(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS")

	assert.NotNil(t, sc.Error)
	assert.NotNil(t, sc.ErrorOrigin)
	assert.Equal(t, sc.Error.Error(), "DUMMY_ERROR")
	assert.Equal(t, sc.ErrorDetail, "DUMMY_DETAILS")
	assert.Equal(t, sc.ErrorOrigin.Localize(), originEnvironmentDescriptor.Localize())

}
