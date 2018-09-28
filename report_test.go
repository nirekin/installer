package installer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDescriber struct {
}

func (m MockDescriber) HumanDescribe() string {
	return "MockDescriber_Content"
}

func TestReportContent(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStep(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepInstaller(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	InstallerFail(&sc, fmt.Errorf("DUMMY_ERROR"), "")
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepInstallerNilError(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	InstallerFail(&sc, nil, "")
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepDescriptor(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	DescriptorFail(&sc, fmt.Errorf("DUMMY_ERROR"), "")
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepDescriptorNilError(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitStepContext("DUMMY_STEP", nil, noCleanUpRequired)
	DescriptorFail(&sc, nil, "")
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentMultipleSteps(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sCs := InitStepContexts()
	sc1 := InitStepContext("DUMMY_STEP1", nil, noCleanUpRequired)
	sCs.Add(sc1)
	sc2 := InitStepContext("DUMMY_STEP2", nil, noCleanUpRequired)
	sCs.Add(sc2)
	sc3 := InitStepContext("DUMMY_STEP2", nil, noCleanUpRequired)
	sCs.Add(sc3)

	_, err = r.Content()
	assert.Nil(t, err)

}
