package installer

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
		for _, sr := range ctx.Results {
			r.Steps.Results = append(r.Steps.Results, sr)
			if sr.cleanUp != nil {
				cleanups = append(cleanups, sr.cleanUp)
			}

			e := sr.error
			if e != nil {
				cleanLaunched(cleanups, c)
				r.Error = e
				return *r
			}
		}
	}
	return *r
}
