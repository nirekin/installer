package main

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

func flocation() (error, cleanup) {
	location = os.Getenv(engine.StarterEnvVariableKey)
	if location == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey), nil
	}
	return nil, noCleanUpRequired
}
