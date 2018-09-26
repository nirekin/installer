package installer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitStep(t *testing.T) {

	s := InitStepContext("stepName", nil, noCleanUpRequired)
	assert.NotNil(t, s)
	assert.Nil(t, s.AppliedTo)
	assert.Nil(t, s.Err)
	assert.Equal(t, s.ErrDetail, "")
	assert.Equal(t, s.StepName, "stepName")

	a := s.Array()
	assert.Equal(t, len(a.contexts), 1)

}

func TestInitSteps(t *testing.T) {

	s := InitStepContexts()
	assert.Equal(t, len(s.contexts), 0)
	s.Add(InitStepContext("stepName1", nil, noCleanUpRequired))
	assert.Equal(t, len(s.contexts), 1)
	s.Add(InitStepContext("stepName2", nil, noCleanUpRequired))
	assert.Equal(t, len(s.contexts), 2)

}

func TestLaunchSteps(t *testing.T) {
	calls := []step{
		fStepMock1,
		fStepMock2,
		fStepMock3,
	}
	rep := launch(calls, &InstallerContext{})
	assert.NotNil(t, rep)
	assert.Nil(t, rep.Error)
	scs := rep.Steps.contexts
	assert.Equal(t, len(scs), 3)

	// Check the order of the executed steps
	assert.Equal(t, scs[0].StepName, "Dummy step 1")
	assert.Equal(t, scs[1].StepName, "Dummy step 2")
	assert.Equal(t, scs[2].StepName, "Dummy step 3")
}

func TestLaunchStepsError(t *testing.T) {
	calls := []step{
		fStepMock1,
		fStepMock2,
		fStepMock3,
		fStepMockError,
	}
	rep := launch(calls, &InstallerContext{})
	assert.NotNil(t, rep)
	assert.NotNil(t, rep.Error)
	scs := rep.Steps.contexts
	assert.Equal(t, len(scs), 4)

	// Check the order of the executed steps
	assert.Equal(t, scs[0].StepName, "Dummy step 1")
	assert.Equal(t, scs[1].StepName, "Dummy step 2")
	assert.Equal(t, scs[2].StepName, "Dummy step 3")
	assert.Equal(t, scs[3].StepName, "Dummy step on error")

}

func TestLaunchStepsError2(t *testing.T) {
	calls := []step{
		fStepMock1,
		fStepMock2,
		fStepMockError,
		fStepMock3,
	}
	rep := launch(calls, &InstallerContext{})
	assert.NotNil(t, rep)
	assert.NotNil(t, rep.Error)
	// Because fStepMockError throws an error fStepMock3 is not invoked and
	// then it is never returned into the report
	scs := rep.Steps.contexts
	assert.Equal(t, len(scs), 3)

	// Check the order of the executed steps
	assert.Equal(t, scs[0].StepName, "Dummy step 1")
	assert.Equal(t, scs[1].StepName, "Dummy step 2")
	assert.Equal(t, scs[2].StepName, "Dummy step on error")

}

func TestLaunchStepsMultiples(t *testing.T) {
	calls := []step{
		fStepMock1,
		fStepMock2,
		fStepMock3,
		fStepMockMultipleContext,
	}
	rep := launch(calls, &InstallerContext{})
	assert.NotNil(t, rep)
	assert.Nil(t, rep.Error)
	scs := rep.Steps.contexts
	assert.Equal(t, len(scs), 6)
	// Check the order of the executed steps
	assert.Equal(t, scs[0].StepName, "Dummy step 1")
	assert.Equal(t, scs[1].StepName, "Dummy step 2")
	assert.Equal(t, scs[2].StepName, "Dummy step 3")
	assert.Equal(t, scs[3].StepName, "Dummy step, multiple 1")
	assert.Equal(t, scs[4].StepName, "Dummy step, multiple 2")
	assert.Equal(t, scs[5].StepName, "Dummy step, multiple 3")

}

func fStepMock1(c *InstallerContext) stepContexts {
	sc := InitStepContext("Dummy step 1", nil, noCleanUpRequired)
	return sc.Array()
}

func fStepMock2(c *InstallerContext) stepContexts {
	sc := InitStepContext("Dummy step 2", nil, noCleanUpRequired)
	return sc.Array()
}

func fStepMock3(c *InstallerContext) stepContexts {
	sc := InitStepContext("Dummy step 3", nil, noCleanUpRequired)
	return sc.Array()
}

func fStepMockError(c *InstallerContext) stepContexts {
	sc := InitStepContext("Dummy step on error", nil, noCleanUpRequired)
	sc.Err = fmt.Errorf("Dummy error")
	return sc.Array()
}

func fStepMockMultipleContext(c *InstallerContext) stepContexts {
	scs := InitStepContexts()
	scs.Add(InitStepContext("Dummy step, multiple 1", nil, noCleanUpRequired))
	scs.Add(InitStepContext("Dummy step, multiple 2", nil, noCleanUpRequired))
	scs.Add(InitStepContext("Dummy step, multiple 3", nil, noCleanUpRequired))
	return *scs
}
