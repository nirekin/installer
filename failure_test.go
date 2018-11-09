package installer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeFail(t *testing.T) {

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnCode(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS", nil)

	assert.Equal(t, sc.StepName, "DUMMY_STEP")
	assert.Equal(t, sc.AppliedToType, "")
	assert.Equal(t, sc.AppliedToName, "")
	assert.Equal(t, sc.Status, STEP_STATUS_FAILURE)
	assert.Equal(t, sc.Context, STEP_CONTEXT_CODE)
	assert.Equal(t, sc.FailureCause, CODE_FAILURE)
	assert.NotNil(t, sc.error)
	assert.Equal(t, sc.ErrorMessage, "DUMMY_ERROR")
	assert.Equal(t, sc.ReadableMessage, "DUMMY_DETAILS")
	assert.Nil(t, sc.RawContent)

}

func TestDescriptorFail(t *testing.T) {

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnDescriptor(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS", nil)

	assert.Equal(t, sc.StepName, "DUMMY_STEP")
	assert.Equal(t, sc.AppliedToType, "")
	assert.Equal(t, sc.AppliedToName, "")
	assert.Equal(t, sc.Status, STEP_STATUS_FAILURE)
	assert.Equal(t, sc.Context, STEP_CONTEXT_CODE)
	assert.Equal(t, sc.FailureCause, DESCRIPTOR_FAILURE)
	assert.NotNil(t, sc.error)
	assert.Equal(t, sc.ErrorMessage, "DUMMY_ERROR")
	assert.Equal(t, sc.ReadableMessage, "DUMMY_DETAILS")
	assert.Nil(t, sc.RawContent)

}

func TestPlaybookFail(t *testing.T) {

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnPlaybook(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS", nil)

	assert.Equal(t, sc.StepName, "DUMMY_STEP")
	assert.Equal(t, sc.AppliedToType, "")
	assert.Equal(t, sc.AppliedToName, "")
	assert.Equal(t, sc.Status, STEP_STATUS_FAILURE)
	assert.Equal(t, sc.Context, STEP_CONTEXT_CODE)
	assert.Equal(t, sc.FailureCause, PLAYBOOK_FAILURE)
	assert.NotNil(t, sc.error)
	assert.Equal(t, sc.ErrorMessage, "DUMMY_ERROR")
	assert.Equal(t, sc.ReadableMessage, "DUMMY_DETAILS")
	assert.Nil(t, sc.RawContent)

}

func TestNotImplentedFail(t *testing.T) {

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnNotImplemented(&sc, fmt.Errorf("DUMMY_ERROR"), "DUMMY_DETAILS", nil)

	assert.Equal(t, sc.StepName, "DUMMY_STEP")
	assert.Equal(t, sc.AppliedToType, "")
	assert.Equal(t, sc.AppliedToName, "")
	assert.Equal(t, sc.Status, STEP_STATUS_FAILURE)
	assert.Equal(t, sc.Context, STEP_CONTEXT_CODE)
	assert.Equal(t, sc.FailureCause, NOT_IMPLEMENTED_FAILURE)
	assert.NotNil(t, sc.error)
	assert.Equal(t, sc.ErrorMessage, "DUMMY_ERROR")
	assert.Equal(t, sc.ReadableMessage, "DUMMY_DETAILS")
	assert.Nil(t, sc.RawContent)

}
