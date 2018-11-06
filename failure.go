package installer

import (
	"fmt"
)

type EngineError struct {
	Message         string `json:",omitempty"`
	DetailedMessage string `json:",omitempty"`
	Origin          string `json:",omitempty"`
	Content         string `json:",omitempty"`
	Source          string `json:",omitempty"`
}

type ErrorLocalizer interface {
	Localize() string
	Source() string
}

const (
	originEkaraInstaller        simpleErrorOrigin = "Ekara Installer"
	originEnvironmentDescriptor simpleErrorOrigin = "Environment descriptor"
)

var InstallerFail = failOn(originEkaraInstaller)
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

func (c simpleErrorOrigin) Source() string {
	return "Code"
}

type playBookErrorOrigin struct {
	Playbook  string
	Compoment string
	Code      int
}

func (p playBookErrorOrigin) Localize() string {
	return fmt.Sprintf("Playbook: %s in component %s returned %d", p.Playbook, p.Compoment, p.Code)
}

func (p playBookErrorOrigin) Source() string {
	return "playbook"
}
