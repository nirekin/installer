package installer

import (
	"encoding/json"

	"github.com/ekara-platform/engine/util"
)

type (
	ExecutionReport struct {
		Error   error
		Steps   stepResults
		Context *InstallerContext
	}

	ReportFileContent struct {
		Results []stepResult
	}

	ReportFailures struct {
		playBookFailures []stepResult
		otherFailures    []stepResult
	}
)

func writeReport(rep ExecutionReport) error {
	loc, e := rep.generate()
	if e != nil {
		return e
	}
	rep.Context.log.Printf(LOG_REPORT_WRITTEN, loc)
	if rep.Error != nil {
		return rep.Error
	}
	return nil
}

// Content returns the json representation of the report steps
func (er ExecutionReport) Content() (b []byte, e error) {
	b, e = json.MarshalIndent(&er.Steps, "", "    ")
	return
}

func (er ExecutionReport) generate() (string, error) {
	b, err := er.Content()
	if err != nil {
		return "", err
	}
	return util.SaveFile(er.Context.log, *er.Context.ef.Output, REPORT_OUTPUT_FILE, b)
}

func (rfc ReportFileContent) hasFailure() (bool, ReportFailures) {
	r := ReportFailures{}

	for _, v := range rfc.Results {
		if v.Status == STEP_STATUS_FAILURE {

			switch v.FailureCause {
			case CODE_FAILURE, DESCRIPTOR_FAILURE:
				r.otherFailures = append(r.otherFailures, v)
				break

			case PLAYBOOK_FAILURE:
				r.playBookFailures = append(r.playBookFailures, v)
				break

			}
		}
	}
	return len(r.playBookFailures) > 0 || len(r.otherFailures) > 0, r
}
