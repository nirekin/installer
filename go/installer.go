package main

import (
	"log"
	"os"
	"path"
	_ "time"

	"github.com/lagoon-platform/engine"
)

var (
	location  string
	loggerLog *log.Logger
	loggerErr *log.Logger
	lagoon    engine.Lagoon
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

// main starts the process of the installer.
//
// This method is supposed to be launched via an entrypoint through the Dockerfile
// used to generate the image.
func main() {
	loggerLog = log.New(os.Stdout, engine.InstallerLogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	loggerLog.Println(LOG_STARTING)

	// Get the env var telling if we are in create or update mode
	create, err := engine.CheckCUMode()

	if err != nil {
		loggerErr.Fatal(err)
	}
	loggerLog.Printf(LOG_INSTALLER_MODE, create)

	// We check if the proxy is well defined, the proxy is required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	httpProxy, httpsProxy, noProxy := engine.CheckProxy()

	var client string
	// If we are in creation mode we check if the environment descriptor is well
	// specified
	if create {
		client = os.Getenv(engine.ClientEnvVariableKey)
		if client == "" {
			loggerErr.Fatalf(ERROR_REQUIRED_ENV, engine.ClientEnvVariableKey)
		}
		loggerLog.Printf(LOG_CREATION_FOR_CLIENT, client)
	}

	// 	Get all the content to create.
	location = os.Getenv(engine.StarterEnvVariableKey)
	if location == "" {
		loggerErr.Fatalf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
	}

	// TODO CHECK THE REAL VERSION HERE ONCE IT WILL BE COMMITED BY THE COMPONENT
	lagoon, e := engine.Create(loggerLog, "/var/lib/lagoon", location, "")
	if e != nil {
		loggerErr.Println(e)
		loggerErr.Fatalf(ERROR_PARSING_DESCRIPTOR, e.Error())
	}
	// Check if a session already exists
	var createSession engine.CreationSession

	var d string
	b, s := engine.HasCreationSession(engine.InstallerVolume)
	if !b {
		createSession = engine.CreationSession{Client: client, Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	for _, n := range lagoon.Environment().NodeSets {
		loggerLog.Printf(LOG_PROCESSING_NODE, n.Name)

		if val, ok := createSession.Uids[n.Name]; ok {
			loggerLog.Printf(LOG_REUSING_UID_FOR_CLIENT, val, client, n.Name)
			d = path.Join(engine.InstallerVolume, val)
		} else {
			uid := engine.GetUId()
			loggerLog.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, client, n.Name)

			b, e := n.ExtraVars(client, uid, engine.InstallerVolume)
			if e != nil {
				loggerErr.Fatal(e)
			}
			createSession.Add(n.Name, uid)

			d = path.Join(engine.InstallerVolume, uid)
			engine.SaveFile(loggerLog, d, engine.NodeConfigFileName, b)

			b, e = n.DockerVars()
			if e != nil {
				loggerErr.Fatal(e)
			}
			engine.SaveFile(loggerLog, d, engine.NodeDockerFileName, b)

			if httpProxy != "" {
				engine.SaveProxy(loggerLog, d, httpProxy, httpsProxy, noProxy)
			}
		}
		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("ANSIBLE_INVENTORY", "/opt/lagoon/ansible/aws-provider/scripts/ec2.py")
		os.Setenv("EC2_INI_PATH", "/opt/lagoon/ansible/aws-provider/scripts/ec2.ini")
		os.Setenv("http_proxy", httpProxy)
		os.Setenv("https_proxy", httpsProxy)
		// TODO ENV VARIABLES SHOULD BE PASSED TO THE PLAYBOOK LAUNCHER

		// TODO WAIT FOR THE END OF THE NEW COMPONENT SPECIFICATIONS
		// AND ADAPT THE PLAYBOOK NAME AND TAKE IN ACCOUNT THE HOOKS
		// TODO USE THE REAL COMPONENT LOCATION COMMING FROM THE COMPONENTS MANAGER
		engine.LaunchPlayBook("/opt/lagoon/ansible/aws-provider", "provisioning-stack.yml", "config_dir="+d, *loggerLog)
	}

	by, e := createSession.Content()
	if e != nil {
		loggerErr.Fatal(e)
	}
	engine.SaveFile(loggerLog, engine.InstallerVolume, engine.CreationSessionFileName, by)

	// Dummy log line just for testing purposes
	//loggerLog.Println("Last Super log from installer ")
}
