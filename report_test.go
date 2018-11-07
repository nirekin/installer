package installer

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine/util"
	"github.com/stretchr/testify/assert"
)

type MockDescriber struct {
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

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepInstaller(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnCode(&sc, fmt.Errorf("DUMMY_ERROR"), "", nil)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepInstallerNilError(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnCode(&sc, nil, "", nil)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepDescriptor(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnDescriptor(&sc, fmt.Errorf("DUMMY_ERROR"), "", nil)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentSingleStepDescriptorNilError(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sc := InitCodeStepResult("DUMMY_STEP", nil, noCleanUpRequired)
	FailsOnDescriptor(&sc, nil, "", nil)
	r.Steps = sc.Array()
	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReportContentMultipleSteps(t *testing.T) {

	r := ExecutionReport{}
	_, err := r.Content()
	assert.Nil(t, err)

	sCs := InitStepResults()
	sc1 := InitCodeStepResult("DUMMY_STEP1", nil, noCleanUpRequired)
	sCs.Add(sc1)
	sc2 := InitCodeStepResult("DUMMY_STEP2", nil, noCleanUpRequired)
	sCs.Add(sc2)
	sc3 := InitCodeStepResult("DUMMY_STEP2", nil, noCleanUpRequired)
	sCs.Add(sc3)

	_, err = r.Content()
	assert.Nil(t, err)

}

func TestReadReport(t *testing.T) {
	var err error
	c := CreateContext(log.New(os.Stdout, util.InstallerLogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds))
	c.ef, err = util.CreateExchangeFolder("./testdata/report", "")
	assert.Nil(t, err)

	ok := c.ef.Output.Contains(REPORT_OUTPUT_FILE)
	assert.True(t, ok)

	stepC := freport(c)
	assert.NotNil(t, stepC)
	assert.NotNil(t, c.report)
	assert.Equal(t, 3, len(c.report.Results))
	has, cpt := c.report.hasFailure()
	assert.True(t, has)
	assert.Equal(t, 0, len(cpt.otherFailures))
	assert.Equal(t, 1, len(cpt.playBookFailures))
}
