package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"io/ioutil"

	"github.com/lagoon-platform/engine"
)

var (
	descriptor string
	loggerLog  *log.Logger
	loggerErr  *log.Logger
	lagoon     engine.Lagoon
)

func main() {

	loggerLog = log.New(os.Stdout, "Lagoon INSTALLER LOG: ", log.Ldate|log.Ltime)
	loggerErr = log.New(os.Stderr, "Lagoon INSTALLER ERROR: ", log.Ldate|log.Ltime)

	checkProxy()
	checkDescriptor()
	checkPlaybooks()
}

func checkPlaybooks() {
	files, err := ioutil.ReadDir("/opt/lagoon/ansible")
	if err != nil {
		loggerErr.Fatal(err)
	}

	for _, f := range files {
		loggerLog.Println(f.Name())
	}

	if _, err := os.Stat("/opt/lagoon/ansible/HelloWorld.yml"); os.IsNotExist(err) {
		loggerErr.Fatalf("/opt/lagoon/ansible/HelloWorld.yml is missing")
	}

	loggerLog.Println("starting playbook 1")
	cmd := exec.Command("ansible-playbook", "HelloWorld.yml")
	cmd.Dir = "/opt/lagoon/ansible/"

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		loggerErr.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			loggerLog.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		loggerErr.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		loggerErr.Fatal(err)
	}

	if _, err := os.Stat("/tmp/testfile.txt"); os.IsNotExist(err) {
		loggerErr.Fatalf("/tmp/testfile.txt has not been created")
	}

	dat, err := ioutil.ReadFile("/tmp/testfile.txt")
	if err != nil {
		loggerErr.Fatal(err)
	}
	loggerLog.Println("------ testfile.txt - content - starts ------")
	loggerLog.Println(string(dat))
	loggerLog.Println("------ testfile.txt - content - ens ------")

}

func checkDescriptor() {
	descriptor = os.Getenv(engine.StarterEnvVariableKey)
	if descriptor == "" {
		loggerErr.Fatalf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
	}
	var err error
	lagoon, err = engine.CreateFromContent(loggerLog, []byte(descriptor))
	if err != nil {
		loggerErr.Fatalf(ERROR_PARSING_DESCRIPTOR, err.Error())
	}
	loggerLog.Printf("Number of providers to process : %d", lagoon.GetEnvironment().GetProviderDescriptions().Count())
	loggerLog.Printf("Number of nodes to create: %d", lagoon.GetEnvironment().GetNodeDescriptions().Count())
}

func checkProxy() {
	if os.Getenv("http_proxy") == "" {
		loggerErr.Fatal(fmt.Errorf(ERROR_REQUIRED_ENV, "http_proxy"))

	}
	if os.Getenv("https_proxy") == "" {
		loggerErr.Fatal(fmt.Errorf(ERROR_REQUIRED_ENV, "https_proxy"))
	}
}
