package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
)

var (
	location      string
	client        string
	sshPublicKey  string
	sshPrivateKey string
	httpProxy     string
	httpsProxy    string
	noProxy       string
	loggerLog     *log.Logger
	lagoon        engine.Lagoon
	lagoonError   error
	ef            engine.ExchangeFolder
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

// main starts the process of the installer.
//
// This method is supposed to be launched via an entrypoint through the Dockerfile
// used to generate the image.
//
func main() {
	loggerLog = log.New(os.Stdout, engine.InstallerLogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	loggerLog.Println(LOG_STARTING)
	e := run()
	if e != nil {
		loggerLog.Fatal(e)
	}
}

func run() (e error) {
	// Check if the received action is supporter by the engine
	loggerLog.Println("Running the installer")
	a := os.Getenv(engine.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		loggerLog.Println("Action Create asked")
		e = runCreate()
	case engine.ActionCheck.String():
		loggerLog.Println("Action Check asked")
		e = runCheck()
	default:
		if a == "" {
			a = "No action specified"
		}
		e = fmt.Errorf(ERROR_UNSUPORTED_ACTION, a)
	}
	return
}

// runCreate launches the envinronemt creation
func runCreate() (e error) {
	// Stack of functions required to create an envinronemnt
	calls := []step{
		fproxy,
		fclient,
		fexchangeFoldef,
		fSHKeys,
		flocation,
		flagoon,
		ffailOnLagoonError,
		fcreate,
	}
	e = launch(calls)
	return
}

// runCheck launches the envinronemt check
func runCheck() (e error) {
	// Stack of functions required to check an envinronemnt
	calls := []step{
		fproxy,
		fexchangeFoldef,
		flocation,
		flagoon,
		flogLagoon,
	}
	e = launch(calls)
	return
}

func fcreate() (error, cleanup) {
	// Check if a session already exists
	var createSession engine.CreationSession
	var d string
	providerFolders, err := enrichExchangeFolder(ef, lagoon.Environment())

	if err != nil {
		return err, nil
	}

	b, s := engine.HasCreationSession(ef)
	if !b {
		createSession = engine.CreationSession{Client: client, Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	for _, n := range lagoon.Environment().NodeSets {
		loggerLog.Printf(LOG_PROCESSING_NODE, n.Name)
		p := n.Provider.ProviderName()

		if val, ok := createSession.Uids[n.Name]; ok {
			loggerLog.Printf(LOG_REUSING_UID_FOR_CLIENT, val, client, n.Name)
			d = path.Join(providerFolders[p].Input.Path(), val)
		} else {
			uid := engine.GetUId()
			loggerLog.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, client, n.Name)
			d = path.Join(providerFolders[p].Input.Path(), n.Name)

			din := path.Join(d, "input")

			//TODO Find a sexy way to pass the output folder to the container
			// avoiding using the config file
			// providerFolders[p].Output.Path()
			b, e := n.Config(client, uid, p, sshPublicKey, sshPrivateKey)
			if e != nil {
				return e, nil
			}
			createSession.Add(n.Name, uid)
			engine.SaveFile(loggerLog, din, engine.NodeConfigFileName, b)

			b, e = n.OrchestratorVars()
			if e != nil {
				return e, nil
			}
			engine.SaveFile(loggerLog, din, engine.OrchestratorFileName, b)

			if httpProxy != "" {
				engine.SaveProxy(loggerLog, din, httpProxy, httpsProxy, noProxy)
			}

			// TODO generate
			// component_paths:
			//  core: /opt/lagoon/ansible/core
			//  aws-provider: /opt/lagoon/ansible/aws-provider
			//  ...
			//  into a file component_paths.yaml
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("ANSIBLE_INVENTORY", "/opt/lagoon/ansible/aws-provider/scripts/ec2.py")
		os.Setenv("EC2_INI_PATH", "/opt/lagoon/ansible/aws-provider/scripts/ec2.ini")
		os.Setenv("PROVIDER_PATH", "/opt/lagoon/ansible/aws-provider")
		os.Setenv("http_proxy", httpProxy)
		os.Setenv("https_proxy", httpsProxy)

		// TODO ENV VARIABLES SHOULD BE PASSED TO THE PLAYBOOK LAUNCHER

		i := 1
		for i <= 20 {
			loggerLog.Printf("log content for testing %d", i)
			i = i + 1
			time.Sleep(time.Millisecond * 1000)
		}

		// TODO WAIT FOR THE END OF THE NEW COMPONENT SPECIFICATIONS
		// AND ADAPT THE PLAYBOOK NAME AND TAKE IN ACCOUNT THE HOOKS
		/*
			e := lagoon.ComponentManager().Ensure()
			if e != nil {
				return e
			}
			paths := lagoon.ComponentManager().ComponentsPaths()
			loggerLog.Printf("Component paths %d", paths)
			repName := lagoon.ComponentManager().ComponentPath(n.Provider.ComponentId())
			loggerLog.Printf("Launching playbook located into %s", repName)
			engine.LaunchPlayBook(repName, "provisioning-stack.yml", "config_dir="+d, *loggerLog)
		*/

		//TODO I need to get the component location fetching by its repository name
		// The trick is that currently i have just aws-provider-f9857c4ca06911bd460bb710ad23e25a5540fab6=/var/lib/lagoon/components/aws-provider-f9857c4ca06911bd460bb710ad23e25a5540fab6
		// And i need lagoon-platform/aws-provider: 0.0.1=/var/lib/lagoon/components/aws-provider-f9857c4ca06911bd460bb710ad23e25a5540fab6
		//engine.LaunchPlayBook("/opt/lagoon/ansible/aws-provider", "provisioning-stack.yml", "input_dir="+d, *loggerLog)

	}

	by, e := createSession.Content()
	if e != nil {
		return e, nil
	}
	engine.SaveFile(loggerLog, engine.InstallerVolume, engine.CreationSessionFileName, by)

	// Dummy log line just for testing purposes
	//loggerLog.Println("Last Super log from installer ")
	return nil, nil
}

func flogLagoon() (error, cleanup) {
	ve := lagoonError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error()), nil
		} else {
			loggerLog.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				return fmt.Errorf(ERROR_GENERIC, e), nil
			}
			// print both errors and warnings into the report file
			engine.SaveFile(loggerLog, ef.Output.Path(), ERROR_CREATING_REPORT_FILE, b)
			if vErrs.HasErrors() {
				// in case of validation error we stop

				return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error()), nil
			}
		}
	} else {
		loggerLog.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return nil, noCleanUpRequired
}

func flagoon() (error, cleanup) {
	// TODO CHECK THE REAL VERSION HERE ONCE IT WILL BE COMMITED BY THE COMPONENT
	lagoon, lagoonError = engine.Create(loggerLog, "/var/lib/lagoon", location, "")
	return nil, noCleanUpRequired
}

func ffailOnLagoonError() (error, cleanup) {
	if lagoonError != nil {
		loggerLog.Println(lagoonError)
		return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, lagoonError.Error()), noCleanUpRequired
	}

	return nil, noCleanUpRequired
}
