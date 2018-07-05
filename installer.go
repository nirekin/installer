package installer

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/model"
	"github.com/lagoon-platform/model/params"
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
		fconsumesetuporchestrator,
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
	c.lagoon.ComponentManager().RegisterComponent(c.lagoon.Environment().Orchestrator.Component)
	return nil, nil
}

func fsetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Running setup for provider %s", p.Name)
		proEf := c.ef.Input.Children[p.Name]
		proEfIn := proEf.Input
		proEfOut := proEf.Output

		bp := params.BuilBaseParam(c.client, "", p.Name, c.sshPublicKey, c.sshPrivateKey)
		b, e := bp.Content()
		if e != nil {
			return e, nil
		}

		e = proEfIn.Write(b, engine.ParamYamlFileName)
		if e != nil {
			return e, nil
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *proEfIn)

		ev := engine.BuildExtraVars("", *proEfIn, *proEfOut)
		path := c.lagoon.ComponentManager().ComponentPath(p.Component.Id)
		module := engine.JoinPaths(path, "modules")
		if _, err := os.Stat(engine.JoinPaths(path, "modules")); err != nil {
			if os.IsNotExist(err) {
				c.log.Printf("No module located in %s", module)
				module = ""
			}
		} else {
			c.log.Printf("Module located in %s", module)
		}
		engine.LaunchPlayBook(path, "setup.yml", ev, module, *c.log)
	}

	return nil, nil
}

func fconsumesetup(c *InstallerContext) (error, cleanup) {
	buffer := Buffer{}
	buffer.Envvars = make(map[string]string)
	buffer.Extravars = make(map[string]string)
	buffer.Param = make(map[string]interface{})
	c.buffer = buffer

	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Consume setup for provider %s", p.Name)
		proEfOut := c.ef.Input.Children[p.Name].Output

		if ok, b, err := proEfOut.ContainsEnvYaml(); ok {
			c.log.Printf("Consuming %s from setup for provider %s", engine.EnvYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Envvars)
			if err != nil {
				return err, noCleanUpRequired
			}

			// TODO MOVE THIS into the init phase of the next step
			for k, v := range c.buffer.Envvars {
				// TODO don't se  os.Setenv(k, v), better to place this at the command level
				os.Setenv(k, v)
				c.log.Printf("Setting env var %s=%s", k, v)
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, b, err := proEfOut.ContainsExtraVarsYaml(); ok {
			c.log.Printf("Consuming %s from setup for provider %s", engine.ExtraVarYamlFileName, p.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Extravars)
			if err != nil {
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, _, err := proEfOut.ContainsParamYaml(); ok {
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
			proEf = c.ef.Input.Children[p.ProviderName()]

			// Node exchange folder
			nodeEf, e = proEf.Input.AddChildExchangeFolder(n.Name)
			if e != nil {
				return e, nil
			}
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.client, n.Name)
			// Provider exchange folder
			proEf = c.ef.Input.Children[p.ProviderName()]

			// Node exchange folder
			nodeEf, e = proEf.Input.AddChildExchangeFolder(n.Name)
			if e != nil {
				return e, nil
			}
			e = nodeEf.Create()
			if e != nil {
				return e, nil
			}

			bp := params.BuilBaseParam(c.client, uid, p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
			np := n.NodeParams()
			bp.AddInt("instances", np.Instances)
			bp.AddMap(np.Params)
			b, e := bp.Content()
			if e != nil {
				return e, nil
			}
			createSession.Add(n.Name, uid)
			engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

			e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
			if e != nil {
				return e, noCleanUpRequired
			}
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)

		ev := engine.BuildExtraVars("", *nodeEf.Input, *nodeEf.Output)

		path := c.lagoon.ComponentManager().ComponentPath(p.Component().Id)
		module := engine.JoinPaths(path, "modules")
		if _, err := os.Stat(engine.JoinPaths(path, "modules")); err != nil {
			if os.IsNotExist(err) {
				c.log.Printf("No module located in %s", module)
				module = ""
			}
		} else {
			c.log.Printf("Module located in %s", module)
		}
		engine.LaunchPlayBook(path, "create.yml", ev, module, *c.log)
	}

	by, e := createSession.Content()
	if e != nil {
		return e, nil
	}
	engine.SaveFile(c.log, *c.ef.Location, engine.CreationSessionFileName, by)

	return nil, nil
}

func fconsumecreate(c *InstallerContext) (error, cleanup) {
	buffer := Buffer{}
	buffer.Envvars = make(map[string]string)
	buffer.Extravars = make(map[string]string)
	buffer.Param = make(map[string]interface{})
	c.buffer = buffer

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume create for node %s", n.Name)
		p := n.Provider
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]
		doutFolder := nodeEf.Output

		if ok, b, err := doutFolder.ContainsEnvYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.EnvYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Envvars)
			if err != nil {
				return err, noCleanUpRequired
			}
			for k, v := range c.buffer.Envvars {
				// TODO don't se  os.Setenv(k, v), better to place this at the command level
				os.Setenv(k, v)
				c.log.Printf("Setting env var %s=%s", k, v)
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsExtraVarsYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.ExtraVarYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Extravars)
			if err != nil {
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.ExtraVarYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsParamYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.ParamYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Param)
			if err != nil {
				c.log.Printf("Error consuming the param %s", err.Error())
				return err, noCleanUpRequired
			}
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
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]

		eq := engine.BuildEquals(c.buffer.Extravars)
		c.log.Printf("Passing extra var to orchestrator setup: %s", eq)
		ev := engine.BuildExtraVars(eq, *nodeEf.Input, *nodeEf.Output)

		bp := params.BuilBaseParam(c.client, "", p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
		op := n.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddMap(c.buffer.Param)

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

		path := c.lagoon.ComponentManager().ComponentPath(c.lagoon.Environment().Orchestrator.Component.Id)
		module := engine.JoinPaths(path, "modules")
		if _, err := os.Stat(engine.JoinPaths(path, "modules")); err != nil {
			if os.IsNotExist(err) {
				c.log.Printf("No module located in %s", module)
				module = ""
			}
		} else {
			c.log.Printf("Module located in %s", module)
		}
		engine.LaunchPlayBook(path, "setup.yml", ev, module, *c.log)
	}

	return nil, nil
}

func fconsumesetuporchestrator(c *InstallerContext) (error, cleanup) {
	buffer := Buffer{}
	buffer.Envvars = make(map[string]string)
	buffer.Extravars = make(map[string]string)
	buffer.Param = make(map[string]interface{})
	c.buffer = buffer

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume create for node %s", n.Name)
		p := n.Provider
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]
		doutFolder := nodeEf.Output

		if ok, b, err := doutFolder.ContainsEnvYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.EnvYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Envvars)
			if err != nil {
				return err, noCleanUpRequired
			}
			for k, v := range c.buffer.Envvars {
				// TODO don't se  os.Setenv(k, v), better to place this at the command level
				os.Setenv(k, v)
				c.log.Printf("Setting env var %s=%s", k, v)
			}
		} else {
			c.log.Printf("No %s located...", engine.EnvYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsExtraVarsYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.ExtraVarYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Extravars)
			if err != nil {
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.ExtraVarYamlFileName)
		}

		if ok, b, err := doutFolder.ContainsParamYaml(); ok {
			c.log.Printf("Consuming %s from create for node %s", engine.ParamYamlFileName, n.Name)
			if err != nil {
				return err, noCleanUpRequired
			}
			err = yaml.Unmarshal([]byte(b), c.buffer.Param)
			if err != nil {
				c.log.Printf("Error consuming the param %s", err.Error())
				return err, noCleanUpRequired
			}
		} else {
			c.log.Printf("No %s located...", engine.ParamYamlFileName)
		}
	}
	return nil, nil
}

func forchestrator(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		p := n.Provider
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]

		bp := params.BuilBaseParam(c.client, "", p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
		op := n.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddMap(c.buffer.Param)
		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// TODO REMOVE HARDCODED STUFF AND BASED THIS ON THE RECEIVED ENV FILE
		os.Setenv("http_proxy", c.httpProxy)
		os.Setenv("https_proxy", c.httpsProxy)

		ev := engine.BuildExtraVars("", *nodeEf.Input, *nodeEf.Output)

		path := c.lagoon.ComponentManager().ComponentPath(c.lagoon.Environment().Orchestrator.Component.Id)
		module := engine.JoinPaths(path, "modules")
		if _, err := os.Stat(engine.JoinPaths(path, "modules")); err != nil {
			if os.IsNotExist(err) {
				c.log.Printf("No module located in %s", module)
				module = ""
			}
		} else {
			c.log.Printf("Module located in %s", module)
		}

		engine.LaunchPlayBook(path, "install.yml", ev, module, *c.log)
	}

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
