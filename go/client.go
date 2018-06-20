package main

import (
	"fmt"
	"os"

	"github.com/lagoon-platform/engine"
)

func fclient() (error, cleanup) {
	client = os.Getenv(engine.ClientEnvVariableKey)
	if client == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.ClientEnvVariableKey), noCleanUpRequired
	}
	loggerLog.Printf(LOG_CREATION_FOR_CLIENT, client)
	return nil, noCleanUpRequired
}
