package main

import (
	"fmt"
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
	e := run()
	if e != nil {
		loggerLog.Fatal(e)
	}
}

func run() (e error) {
	// Check if the received action is supporter by the engine
	a := os.Getenv(engine.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		e = runCreate()
	case engine.ActionCheck.String():
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
	calls := []func() error{
		fproxy,
		fclient,
		fexchangeFoldef,
		flocation,
		flagoon,
		ffailOnLagoonError,
		fcreate,
	}
	e = launch(calls)
	if e != nil {
		return
	}
	return
}

// runCheck launches the envinronemt check
func runCheck() (e error) {
	// Stack of functions required to check an envinronemnt
	calls := []func() error{
		fproxy,
		fexchangeFoldef,
		flocation,
		flagoon,
		flogLagoon,
	}
	e = launch(calls)
	if e != nil {
		return
	}
	return
}

//launch runs a slice of functions
func launch(fs []func() error) error {
	for _, f := range fs {
		e := f()
		if e != nil {
			return e
		}
	}
	return nil
}

func fcreate() error {
	// Check if a session already exists
	var createSession engine.CreationSession
	var d string
	providerFolders, err := enrichExchangeFolder(ef, lagoon.Environment())

	if err != nil {
		return (err)
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
			b, e := n.ExtraVars(client, uid, providerFolders[p].Output.Path())
			if e != nil {
				return (e)
			}
			createSession.Add(n.Name, uid)

			engine.SaveFile(loggerLog, din, engine.NodeConfigFileName, b)

			b, e = n.DockerVars()
			if e != nil {
				return (e)
			}
			engine.SaveFile(loggerLog, din, engine.NodeDockerFileName, b)

			if httpProxy != "" {
				engine.SaveProxy(loggerLog, din, httpProxy, httpsProxy, noProxy)
			}
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("ANSIBLE_INVENTORY", "/opt/lagoon/ansible/aws-provider/scripts/ec2.py")
		os.Setenv("EC2_INI_PATH", "/opt/lagoon/ansible/aws-provider/scripts/ec2.ini")
		os.Setenv("PROVIDER_PATH", "/opt/lagoon/ansible/aws-provider")

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
		//engine.LaunchPlayBook("/opt/lagoon/ansible/aws-provider", "provisioning-stack.yml", "config_dir="+d, *loggerLog)
	}

	by, e := createSession.Content()
	if e != nil {
		return e
	}
	engine.SaveFile(loggerLog, engine.InstallerVolume, engine.CreationSessionFileName, by)

	// Dummy log line just for testing purposes
	//loggerLog.Println("Last Super log from installer ")
	return nil
}

func flogLagoon() error {
	ve := lagoonError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error())
		} else {
			loggerLog.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				return fmt.Errorf(ERROR_GENERIC, e)
			}
			// print both errors and warnings into the report file
			engine.SaveFile(loggerLog, ef.Output.Path(), ERROR_CREATING_REPORT_FILE, b)
			if vErrs.HasErrors() {
				// in case of validation error we stop

				return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error())
			}
		}
	} else {
		loggerLog.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return nil
}

func fexchangeFoldef() error {
	var err error
	ef, err = engine.CreateExchangeFolder(engine.InstallerVolume, "")
	if err != nil {
		return fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, engine.ClientEnvVariableKey)
	}
	return nil
}

func flocation() error {
	location = os.Getenv(engine.StarterEnvVariableKey)
	if location == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.StarterEnvVariableKey)
	}
	return nil
}

func fclient() error {
	client = os.Getenv(engine.ClientEnvVariableKey)
	if client == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, engine.ClientEnvVariableKey)
	}
	loggerLog.Printf(LOG_CREATION_FOR_CLIENT, client)
	return nil
}

func fproxy() error {
	// We check if the proxy is well defined, the proxy is required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	httpProxy, httpsProxy, noProxy = engine.CheckProxy()
	return nil
}

func flagoon() error {
	// TODO CHECK THE REAL VERSION HERE ONCE IT WILL BE COMMITED BY THE COMPONENT
	lagoon, lagoonError = engine.Create(loggerLog, "/var/lib/lagoon", location, "")
	return nil
}

func ffailOnLagoonError() error {
	if lagoonError != nil {
		loggerLog.Println(lagoonError)
		return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, lagoonError.Error())
	}
	return nil
}

//enrichExchangeFolder adds a sub level of ExchangeFolder to the received one.
//This will add a "sub ExchangeFolder" for each provider willing to be in charge
// of a nodeset creation.
func enrichExchangeFolder(f engine.ExchangeFolder, e model.Environment) (r map[string]engine.ExchangeFolder, err error) {
	r = make(map[string]engine.ExchangeFolder)
	for _, n := range lagoon.Environment().NodeSets {
		p := n.Provider.ProviderName()
		if !f.Input.Contains(p) {
			var pEf engine.ExchangeFolder
			pEf, err = engine.CreateExchangeFolder(f.Input.Path(), p)
			if err != nil {
				return
			}
			err = pEf.Create()
			if err != nil {
				return
			}
			r[p] = pEf
		}
	}
	return
}
