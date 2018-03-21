package main

import (
	"log"
	"os"

	"io/ioutil"
	_ "path/filepath"

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
		logger.Printf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
		os.Exit(1)
	}
	parseDescriptor(descriptor)

	fileName := "container_output.json"
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write([]byte("written from the go code running on the container"))
	if err != nil {
		panic(err)
	}
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		logger.Println(f.Name())
	}
}

func parseDescriptor(descriptor string) ([]byte, error) {
	logger.Printf(LOG_PARSING)

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
