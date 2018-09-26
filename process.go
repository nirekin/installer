package installer

import (
	"encoding/json"
	"fmt"

	"github.com/lagoon-platform/model"
)

type ErrorLocalizer interface {
	Localize() string
}

type SimpleErrorOrigin string

func (c SimpleErrorOrigin) Localize() string {
	return string(c)
}

const (
	OriginLagoonInstaller       SimpleErrorOrigin = "Lagoon Installer"
	OriginEnvironmentDescriptor SimpleErrorOrigin = "Environment descriptor"
)

type PlayBookErrorOrigin struct {
	Playbook  string
	Compoment string
}

func (p PlayBookErrorOrigin) Localize() string {
	return fmt.Sprintf("Playbook: %s in component %s", p.Playbook, p.Compoment)
}

// stepContexts represents a chain of steps execution results
type stepContexts struct {
	Contexts []stepContext
}

// stepContext represents the execution result of a single step with its context
type stepContext struct {
	StepName    string
	AppliedTo   model.HumanDescriber
	Error       error
	ErrorDetail string
	ErrorOrigin ErrorLocalizer
	CleanUp     cleanup
}

func (r stepContext) MarshalJSON() ([]byte, error) {

	s := struct {
		StepName  string `json:",omitempty"`
		AppliedTo string `json:",omitempty"`
		Status    string
		Error     *struct {
			Message         string `json:",omitempty"`
			DetailedMessage string `json:",omitempty"`
			Origin          string `json:",omitempty"`
		} `json:",omitempty"`
	}{
		StepName: r.StepName,
	}
	if r.AppliedTo != nil {
		s.AppliedTo = r.AppliedTo.HumanDescribe()
	}

	if r.Error != nil {
		s.Error.Message = r.Error.Error()
		s.Status = "Failure"
	} else {
		s.Status = "Success"
	}
	if r.ErrorDetail != "" {
		s.Error.DetailedMessage = r.ErrorDetail
	}

	if r.ErrorOrigin != nil {
		s.Error.Origin = r.ErrorOrigin.Localize()
	}

	return json.MarshalIndent(s, "", "    ")
}

// Array() initialize an array with the step context
func (sc stepContext) Array() stepContexts {
	return stepContexts{
		Contexts: []stepContext{sc},
	}
}

func InitStepContext(stepName string, appliedTo model.HumanDescriber, c cleanup) stepContext {
	return stepContext{
		StepName:  stepName,
		AppliedTo: appliedTo,
		CleanUp:   c,
	}
}

func InitStepContexts() *stepContexts {
	sCs := &stepContexts{}
	sCs.Contexts = make([]stepContext, 0)
	return sCs
}

func (sc *stepContexts) Add(c stepContext) {
	sc.Contexts = append(sc.Contexts, c)
}

// cleanup represents a cleanup method to rollback what has been done by a step
type cleanup func(c *InstallerContext) error

// step represents a sinlge step used to compose a process executed by the installer
type step func(c *InstallerContext) stepContexts

func noCleanUpRequired(c *InstallerContext) error {
	// Do nothing and it's okay...
	// This is just an explicit empty implementation to clearly materialize that no cleanup is required
	return nil
}

// launch runs a slice of step functions
//
// If one step in the slice returns an error then the launch process will stop and
// the cleanup will be invoked on all previously launched steps
func launch(fs []step, c *InstallerContext) ExecutionReport {

	r := &ExecutionReport{
		Context: c,
	}

	cleanups := []cleanup{}

	for _, f := range fs {
		ctx := f(c)
		for _, cs := range ctx.Contexts {
			r.Steps.Contexts = append(r.Steps.Contexts, cs)
			if cs.CleanUp != nil {
				cleanups = append(cleanups, cs.CleanUp)
			}

			e := cs.Error
			// TODO Consume the context here  in order to build/populate the report file
			if e != nil {
				cleanLaunched(cleanups, c)
				r.Error = e
				return *r
			}
		}
	}
	return *r
}

//cleanLaunched runs a slice of cleanup functions
func cleanLaunched(fs []cleanup, c *InstallerContext) (e error) {
	for _, f := range fs {
		e := f(c)
		if e != nil {
			return e
		}
	}
	return nil
}
