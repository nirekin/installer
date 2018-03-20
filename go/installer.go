package main

import (
	"log"
	"os"

	"github.com/lagoon-platform/engine"
)

const (

	// The environment variable used to pass the environment descriptor
	// content to the installer image.
	starterEnvVariableKey string = "LAGOON_ENV_DESCR"
)

var (
	descriptor string
	logger     *log.Logger
)

func main() {

	logger = log.New(os.Stdout, "Lagoon INSTALLER: ", log.Ldate|log.Ltime)

	descriptor = os.Getenv(starterEnvVariableKey)
	if descriptor == "" {
		log.Printf(ERROR_REQUIRED_ENV, starterEnvVariableKey)
		os.Exit(1)
	}
	log.Printf("Ok passed")
	parseDescriptor(descriptor)
}

func parseDescriptor(descriptor string) ([]byte, error) {
	log.Printf(LOG_PARSING)

	lagoon, e := engine.CreateFromContent(logger, []byte(descriptor))
	if e != nil {
		return nil, e
	}

	content, err := lagoon.GetContent()
	if err != nil {
		return nil, err
	}
	return content, nil
}
