package installer

import (
	"encoding/json"

	"github.com/lagoon-platform/engine"
)

type ExecutionReport struct {
	Error   error
	Steps   stepContexts
	Context *InstallerContext
}

// Content returns the json representation of the report steps
func (er ExecutionReport) Content() (b []byte, e error) {
	b, e = json.Marshal(&er.Steps)
	return
}

func (er ExecutionReport) Generate() (string, error) {
	b, err := er.Content()
	if b != nil {
		return "", err
	}
	return engine.SaveFile(er.Context.log, *er.Context.ef.Output, REPORT_OUTPUT_FILE, b)
}
