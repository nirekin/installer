package installer

import (
	"fmt"
)

type EngineError struct {
	Message         string `json:",omitempty"`
	DetailedMessage string `json:",omitempty"`
	Origin          string `json:",omitempty"`
}

type ErrorLocalizer interface {
	Localize() string
}

const (
	originLagoonInstaller       simpleErrorOrigin = "Lagoon Installer"
	originEnvironmentDescriptor simpleErrorOrigin = "Environment descriptor"
)

var InstallerFail = failOn(originLagoonInstaller)
var DescriptorFail = failOn(originEnvironmentDescriptor)

func failOn(log ErrorLocalizer) func(sc *stepContext, err error, detail string) {
	return func(sc *stepContext, err error, detail string) {
		sc.Error = err
		sc.ErrorDetail = detail
		sc.ErrorOrigin = log
	}
}

type simpleErrorOrigin string

func (c simpleErrorOrigin) Localize() string {
	return string(c)
}

type playBookErrorOrigin struct {
	Playbook  string
	Compoment string
	Code      int
}

func (p playBookErrorOrigin) Localize() string {
	return fmt.Sprintf("Playbook: %s in component %s returned %d", p.Playbook, p.Compoment, p.Code)
}
