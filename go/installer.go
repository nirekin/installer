package main

import (
	"log"
	"os"
	"path"
	_ "time"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
)

var (
	location    string
	client      string
	httpProxy   string
	httpsProxy  string
	noProxy     string
	loggerLog   *log.Logger
	loggerErr   *log.Logger
	lagoon      engine.Lagoon
	lagoonError error
	ef          engine.ExchangeFolder
)

const (
	VALIDATION_OUTPUT_FILE = "validation_output.json"
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
	a := os.Getenv(engine.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		runCreate()
	case engine.ActionCheck.String():
		runCheck()
	default:
		if a == "" {
			a = "No action specified"
		}
		loggerErr.Fatalf(ERROR_UNSUPORTED_ACTION, a)
	}
}

func runCreate() {
	getProxy()
	getClient()
	getExchangeFoldef()
	getLocation()
	createLagoon()
	failOnLagoonError()
	runCreation()
}

func runCheck() {
	getProxy()
	getExchangeFoldef()
	getLocation()
	createLagoon()
	logLagoon()
}

func runCreation() {
	// Check if a session already exists
	var createSession engine.CreationSession
	var d string

	b, s := engine.HasCreationSession(ef)
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

		//i := 1
		//for i <= 20 {
		//	loggerLog.Printf("log content for testing %d", i)
		//	i = i + 1
		//	time.Sleep(time.Millisecond * 1000)
		//}

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

func logLagoon() {
	ve := lagoonError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			loggerErr.Fatalf(ERROR_PARSING_ENVIRONMENT, ve.Error())
		} else {

			loggerLog.Printf(ve.Error())

			b, e := vErrs.JSonContent()
			if e != nil {
				loggerErr.Fatal(e)
			}
			// print both errors and warnings into the report file
			engine.SaveFile(loggerLog, ef.Output.Path(), ERROR_CREATING_REPORT_FILE, b)
			if vErrs.HasErrors() {
				// in case of validation error we stop

				loggerErr.Fatalf(ERROR_PARSING_ENVIRONMENT, ve.Error())
			}
		}
	} else {
		loggerLog.Printf(LOG_VALIDATION_SUCCESSFUL)
	}

}

func getExchangeFoldef() {
	var err error
	ef, err = engine.ClientExchangeFolder(engine.InstallerVolume, "")
	if err != nil {
		loggerErr.Fatalf(ERROR_CREATING_EXCHANGE_FOLDER, engine.ClientEnvVariableKey)
	}
}

func getLocation() {
	location = os.Getenv(engine.StarterEnvVariableKey)
	loggerLog.Printf("GetLocation \"%s\"", location)
	if location == "" {
		loggerErr.Fatalf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
	}
}

func getClient() {
	client = os.Getenv(engine.ClientEnvVariableKey)
	if client == "" {
		loggerErr.Fatalf(ERROR_REQUIRED_ENV, engine.ClientEnvVariableKey)
	}
	loggerLog.Printf(LOG_CREATION_FOR_CLIENT, client)
}

func getProxy() {
	// We check if the proxy is well defined, the proxy is required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	httpProxy, httpsProxy, noProxy = engine.CheckProxy()
}

func createLagoon() {
	// TODO CHECK THE REAL VERSION HERE ONCE IT WILL BE COMMITED BY THE COMPONENT
	lagoon, lagoonError = engine.Create(loggerLog, "/var/lib/lagoon", location, "")
}

func failOnLagoonError() {
	if lagoonError != nil {
		loggerErr.Println(lagoonError)
		loggerErr.Fatalf(ERROR_PARSING_DESCRIPTOR, lagoonError.Error())
	}
}
