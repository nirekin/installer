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

func (r ExecutionReport) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(struct {
		Steps stepContexts `json:",omitempty"`
	}{
		Steps: r.Steps,
	}, "", "    ")
}

// Content returns the json representation of the report steps
func (er ExecutionReport) Content() (b []byte, e error) {
	b, e = json.MarshalIndent(&er.Steps, "", "    ")
	return
}

func (er ExecutionReport) Generate() (string, error) {
	b, err := er.Content()
	if err != nil {
		return "", err
	}
	return engine.SaveFile(er.Context.log, *er.Context.ef.Output, REPORT_OUTPUT_FILE, b)
}

func writeReport(rep ExecutionReport) error {
	loc, e := rep.Generate()
	if e != nil {
		return e
	}
	rep.Context.log.Printf(LOG_REPORT_WRITTEN, loc)
	if rep.Error != nil {
		return rep.Error
	}
	return nil
}
