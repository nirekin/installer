package main

import (
	"log"
	"os"

	"github.com/lagoon-platform/engine"
)

var (
	descriptor string
	logger     *log.Logger
)

func main() {

	logger = log.New(os.Stdout, "Lagoon INSTALLER: ", log.Ldate|log.Ltime)

	descriptor = os.Getenv(engine.StarterEnvVariableKey)
	if descriptor == "" {
		log.Printf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
		os.Exit(1)
	}
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
