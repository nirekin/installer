package installer

import (
	"encoding/json"
)

type ExecutionReport struct {
	Error error
	Steps stepContexts
}

// Content returns the json representation of the report steps
func (er ExecutionReport) Content() (b []byte, e error) {
	b, e = json.Marshal(&er.Steps)
	return
}
