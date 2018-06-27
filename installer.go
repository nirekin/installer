package installer

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

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
		fdownloadcore,
		fenrichExchangeFolder,
		fsetup,
		fconsumesetup,
		fcreate,
		fconsumecreate,
		fsetuporchestrator,
		forchestrator,
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

func fdownloadcore(c *InstallerContext) (error, cleanup) {
	c.log.Printf(LOG_PLATFORM_VERSION, c.lagoon.Environment().LagoonPlatform.Version)
	c.log.Printf(LOG_PLATFORM_REPOSITORY, c.lagoon.Environment().LagoonPlatform.Repository)
	c.log.Printf(LOG_PLATFORM_COMPONENT_ID, c.lagoon.Environment().LagoonPlatform.Component.Id)
	c.lagoon.ComponentManager().RegisterComponent(c.lagoon.Environment().LagoonPlatform.Component)
	return nil, nil
}

func ProviderConfig(p string, pubK string, privK string) (b []byte, e error) {
	mcon := model.MachineConfig{}

	pcon := model.ConnectionConfig{
		Provider:          p,
		MachinePublicKey:  pubK,
		MachinePrivateKey: privK,
	}
	mcon.ConnectionConfig = pcon
	b, e = yaml.Marshal(&mcon)
	return
}

func fsetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Setup for provider %s", p.Name)
		dinFolder := c.ef.Input.Children[p.Name].Input
		doutFolder := c.ef.Input.Children[p.Name].Output

		b, e := ProviderConfig(p.Name, c.sshPublicKey, c.sshPrivateKey)
		if e != nil {
			return e, nil
		}
		e = dinFolder.Write(b, engine.ParamYamlFileName)
		if e != nil {
			return e, nil
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *dinFolder)

		ev := engine.BuildExtraVars("", *dinFolder, *doutFolder)
		engine.LaunchPlayBook(c.lagoon.ComponentManager().ComponentPath(p.Component.Id), "setup.yml", ev, *c.log)
	}
	return nil, nil
}

func fconsumesetup(c *InstallerContext) (error, cleanup) {
	buffer := buffer{}
	buffer.envvars = make(map[string]string)
	buffer.extravars = make(map[string]string)
	c.buffer = buffer

	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Consume setup for provider %s", p.Name)
		doutFolder := c.ef.Input.Children[p.Name].Output

		if ok, b, err := doutFolder.ContainsEnvYaml(); ok {
			c.log.Printf("Consuming %s from setup for provider %s", engine.EnvYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.envvars)
			if err != nil {
				return err, noCleanUpRequired
			}

			// TODO MOVE THIS into the init phase of the next step
			for k, v := range c.buffer.envvars {
				// TODO don't se  os.Setenv(k, v), better to place this at the command level
				os.Setenv(k, v)
				c.log.Printf("Setting env var %s=%s", k, v)
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsExtraVarsYaml(); ok {
			c.log.Printf("Consuming %s from setup for provider %s", engine.ExtraVarYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.extravars)
			if err != nil {
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, _, err := doutFolder.ContainsParamYaml(); ok {
			c.log.Printf("Consuming %s from setup for provider %s", engine.ParamYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}

			// TODO consume this
			// By moving it into the init phase of the next step
		} else {
			c.log.Printf("No %s located...", engine.ParamYamlFileName)
		}
	}
	return nil, nil
}

func fcreate(c *InstallerContext) (error, cleanup) {
	// Check if a session already exists
	var createSession engine.CreationSession
	var proEf *engine.ExchangeFolder
	var nodeEf *engine.ExchangeFolder
	var e error

	b, s := engine.HasCreationSession(*c.ef)
	if !b {
		createSession = engine.CreationSession{Client: c.client, Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		p := n.Provider

		if val, ok := createSession.Uids[n.Name]; ok {
			c.log.Printf(LOG_REUSING_UID_FOR_CLIENT, val, c.client, n.Name)
			// Provider exchange folder
			proEf, e = engine.CreateExchangeFolder(c.ef.Input.Path(), p.ProviderName())
			if e != nil {
				return e, nil
			}
			// Node exchange folder
			nodeEf, e = engine.CreateExchangeFolder(proEf.Input.Path(), val)
			if e != nil {
				return e, nil
			}
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.client, n.Name)
			// Provider exchange folder
			proEf, e = engine.CreateExchangeFolder(c.ef.Input.Path(), p.ProviderName())
			if e != nil {
				return e, nil
			}
			// Node exchange folder
			nodeEf, e = engine.CreateExchangeFolder(proEf.Input.Path(), n.Name)
			if e != nil {
				return e, nil
			}
			e = nodeEf.Create()
			if e != nil {
				return e, nil
			}
			b, e := n.NodeParams(c.client, uid, p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
			if e != nil {
				return e, nil
			}
			createSession.Add(n.Name, uid)
			engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

			b, e = n.OrchestratorParams()
			if e != nil {
				return e, nil
			}
			engine.SaveFile(c.log, *nodeEf.Input, engine.OrchestratorFileName, b)

			e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
			if e != nil {
				return e, noCleanUpRequired
			}
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)
		ev := engine.BuildExtraVars("", *nodeEf.Input, *proEf.Output)
		engine.LaunchPlayBook(c.lagoon.ComponentManager().ComponentPath(p.Component().Id), "create.yml", ev, *c.log)
	}

	by, e := createSession.Content()
	if e != nil {
		return e, nil
	}
	engine.SaveFile(c.log, *c.ef.Location, engine.CreationSessionFileName, by)
	return nil, nil
}

func fconsumecreate(c *InstallerContext) (error, cleanup) {
	buffer := buffer{}
	buffer.envvars = make(map[string]string)
	buffer.extravars = make(map[string]string)
	c.buffer = buffer

	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Consume create for provider %s", p.Name)
		doutFolder := c.ef.Input.Children[p.Name].Output

		if ok, b, err := doutFolder.ContainsEnvYaml(); ok {
			c.log.Printf("Consuming %s from create for provider %s", engine.EnvYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.envvars)
			if err != nil {
				return err, noCleanUpRequired
			}
			for k, v := range c.buffer.envvars {
				// TODO don't se  os.Setenv(k, v), better to place this at the command level
				os.Setenv(k, v)
				c.log.Printf("Setting env var %s=%s", k, v)
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsExtraVarsYaml(); ok {
			c.log.Printf("Consuming %s from create for provider %s", engine.ExtraVarYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.extravars)
			if err != nil {
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, _, err := doutFolder.ContainsParamYaml(); ok {
			c.log.Printf("Consuming %s from create for provider %s", engine.ParamYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}

			// TODO consume this
		} else {
			c.log.Printf("No %s located...", engine.ParamYamlFileName)
		}
	}
	return nil, nil
}

func fsetuporchestrator(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		p := n.Provider

		// Provider exchange folder
		proEf, e := engine.CreateExchangeFolder(c.ef.Input.Path(), p.ProviderName())
		if e != nil {
			return e, nil
		}
		// TODO CONSUME BUFFER HERE consume the map of extra var...
		ev := engine.BuildExtraVars("", *proEf.Input, *proEf.Output)
		engine.LaunchPlayBook(c.lagoon.ComponentManager().ComponentPath(c.lagoon.Environment().Orchestrator.Component.Id), "setup.yml", ev, *c.log)
	}
	return nil, nil
}

func forchestrator(c *InstallerContext) (error, cleanup) {

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
			engine.SaveFile(c.log, *c.ef.Output, ERROR_CREATING_REPORT_FILE, b)
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
