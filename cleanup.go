package installer

// cleanup represents a cleanup method to rollback what has been done by a step
type cleanup func(c *InstallerContext) error

func noCleanUpRequired(c *InstallerContext) error {
	// Do nothing and it's okay...
	// This is just an explicit empty implementation to clearly materialize that no cleanup is required
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
