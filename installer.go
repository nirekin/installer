package installer

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

func Run(c *InstallerContext) (e error) {
	// Check if the received action is supporter by the engine
	c.log.Println("Running the installer")
	a := os.Getenv(engine.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		c.log.Println("Action Create asked")
		e = runCreate(c)
	case engine.ActionCheck.String():
		c.log.Println("Action Check asked")
		e = runCheck(c)
	default:
		if a == "" {
			a = "No action specified"
		}
		e = fmt.Errorf(ERROR_UNSUPORTED_ACTION, a)
	}
	return
}

// runCreate launches the environment creation
func runCreate(c *InstallerContext) (e error) {
	// Stack of functions required to create an environment
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
	e = launch(calls, c)
	return
}

// runCheck launches the environment check
func runCheck(c *InstallerContext) (e error) {
	// Stack of functions required to check an environment
	calls := []step{
		fproxy,
		fexchangeFoldef,
		flocation,
		flagoon,
		flogLagoon,
	}
	e = launch(calls, c)
	return
}

func fcreate(c *InstallerContext) (error, cleanup) {
	// Check if a session already exists
	var createSession engine.CreationSession
	var d string
	providerFolders, err := enrichExchangeFolder(c.ef, c.lagoon.Environment(), c)

	if err != nil {
		return err, nil
	}

	b, s := engine.HasCreationSession(c.ef)
	if !b {
		createSession = engine.CreationSession{Client: c.client, Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		p := n.Provider.ProviderName()

		if val, ok := createSession.Uids[n.Name]; ok {
			c.log.Printf(LOG_REUSING_UID_FOR_CLIENT, val, c.client, n.Name)
			d = path.Join(providerFolders[p].Input.Path(), val)
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.client, n.Name)
			d = path.Join(providerFolders[p].Input.Path(), n.Name)

			din := path.Join(d, "input")

			//TODO Find a sexy way to pass the output folder to the container
			// avoiding using the config file
			// providerFolders[p].Output.Path()
			b, e := n.Config(c.client, uid, p, c.sshPublicKey, c.sshPrivateKey)
			if e != nil {
				return e, nil
			}
			createSession.Add(n.Name, uid)
			engine.SaveFile(c.log, din, engine.NodeConfigFileName, b)

			b, e = n.OrchestratorVars()
			if e != nil {
				return e, nil
			}
			engine.SaveFile(c.log, din, engine.OrchestratorFileName, b)

			if c.httpProxy != "" {
				engine.SaveProxy(c.log, din, c.httpProxy, c.httpsProxy, c.noProxy)
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
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)

		// TODO ENV VARIABLES SHOULD BE PASSED TO THE PLAYBOOK LAUNCHER

		i := 1
		for i <= 20 {
			c.log.Printf("log content for testing %d", i)
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
			c.log.Printf("Component paths %d", paths)
			repName := lagoon.ComponentManager().ComponentPath(n.Provider.ComponentId())
			c.log.Printf("Launching playbook located into %s", repName)
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
	engine.SaveFile(c.log, engine.InstallerVolume, engine.CreationSessionFileName, by)

	// Dummy log line just for testing purposes
	//c.log.Println("Last Super log from installer ")
	return nil, nil
}

func flogLagoon(c *InstallerContext) (error, cleanup) {
	ve := c.lagoonError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error()), nil
		} else {
			c.log.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				return fmt.Errorf(ERROR_GENERIC, e), nil
			}
			// print both errors and warnings into the report file
			engine.SaveFile(c.log, c.ef.Output.Path(), ERROR_CREATING_REPORT_FILE, b)
			if vErrs.HasErrors() {
				// in case of validation error we stop

				return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error()), nil
			}
		}
	} else {
		c.log.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return nil, noCleanUpRequired
}

func flagoon(c *InstallerContext) (error, cleanup) {
	// TODO CHECK THE REAL VERSION HERE ONCE IT WILL BE COMMITED BY THE COMPONENT
	c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", c.location, "")
	return nil, noCleanUpRequired
}

func ffailOnLagoonError(c *InstallerContext) (error, cleanup) {
	if c.lagoonError != nil {
		c.log.Println(c.lagoonError)
		return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.lagoonError.Error()), noCleanUpRequired
	}

	return nil, noCleanUpRequired
}
