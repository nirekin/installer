package installer

// cleanup represents a cleanup method to rollback what has been done by a step
type cleanup func(c *InstallerContext) error

// step represents a sinlge ste used to compose a process executed by the installer
type step func(c *InstallerContext) (error, cleanup)

func noCleanUpRequired(c *InstallerContext) error {
	// Do nothing and it's okay...
	// This is just an explicit empty implementation to clearly materialize that no cleanup is required
	return nil
}

// launch runs a slice of step functions
//
// If one step in the slice returns an error then the launch process will stop and
// the cleanup will be invoked on all previously launched steps
func launch(fs []step, c *InstallerContext) error {

	cleanups := []cleanup{}
	for _, f := range fs {
		e, clean := f(c)
		if clean != nil {
			cleanups = append(cleanups, clean)
		}
		if e != nil {
			cleanLaunched(cleanups, c)
			return e
		}
	}
	return nil
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
