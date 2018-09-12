package installer

import (
	"fmt"
	"os"
	"strings"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/engine/ansible"
	"github.com/lagoon-platform/model"
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

func Run(c *InstallerContext) (e error) {
	// Check if the received action is supporter by the engine
	c.log.Println(LOG_RUNNING)
	a := os.Getenv(engine.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		c.log.Println(LOG_ACTION_CREATE)
		e = runCreate(c)
	case engine.ActionCheck.String():
		c.log.Println(LOG_ACTION_CHECK)
		e = runCheck(c)
	default:
		if a == "" {
			a = LOG_NO_ACTION
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
		flocation,
		fcliparam,
		flagoon,
		ffailOnLagoonError,
		fsession,
		fSHKeys,
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
		fcliparam,
		flagoon,
		flogLagoon,
	}
	e = launch(calls, c)
	return
}

func fsession(c *InstallerContext) (error, cleanup) {
	// Check if a session already exists
	var createSession *engine.CreationSession

	b, s := engine.HasCreationSession(*c.ef)
	if !b {
		createSession = &engine.CreationSession{Client: c.client, Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	// If needed creates the missing Uids for the nodes
	for _, n := range c.lagoon.Environment().NodeSets {
		if val, ok := createSession.Uids[n.Name]; ok {
			c.log.Printf(LOG_REUSING_UID_FOR_CLIENT, val, c.client, n.Name)
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.client, n.Name)
			createSession.Add(n.Name, uid)
		}
	}
	by, e := createSession.Content()
	if e != nil {
		return e, nil
	}
	f, e := engine.SaveFile(c.log, *c.ef.Location, engine.CreationSessionFileName, by)
	if e != nil {
		return e, nil
	}
	c.session = &engine.EngineSession{
		CreationSession: createSession,
		File:            f,
	}
	return nil, nil
}

func fsetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf(LOG_RUNNING_SETUP_FOR, p.Name)

		// Provider setup exchange folder
		setupProviderEf, e := c.ef.Input.AddChildExchangeFolder("setup_provider_" + p.Name)
		if e != nil {
			c.log.Printf(ERROR_ADDING_EXCHANGE_FOLDER, "setup_provider_"+p.Name, e.Error())
			return e, nil
		}
		e = setupProviderEf.Create()
		if e != nil {
			c.log.Printf(ERROR_CREATING_EXCHANGE_FOLDER, "setup_provider_"+p.Name, e.Error())
			return e, nil
		}
		setupProviderEfIn := setupProviderEf.Input
		setupProviderEfOut := setupProviderEf.Output

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, "", p.Name, c.sshPublicKey, c.sshPrivateKey)
		bp.AddNamedMap("params", p.Parameters)
		b, e := bp.Content()
		if e != nil {
			return e, nil
		}

		_, e = engine.SaveFile(c.log, *setupProviderEfIn, engine.ParamYamlFileName, b)
		if e != nil {
			return e, nil
		}

		// This is the first "real" step of the process so the used buffer is empty
		emptyBuff := engine.CreateBuffer()

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *setupProviderEfIn)
		if e != nil {
			return e, nil
		}

		// Prepare extra vars
		exv := engine.BuildExtraVars("", *setupProviderEfIn, *setupProviderEfOut, emptyBuff)

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)

		// Adding the environment variables from the provider
		for envK, envV := range p.EnvVars {
			env.Add(envK, envV)
		}

		// We launch the playbook
		ansible.LaunchPlayBook(c.lagoon.ComponentManager(), p.Component.Resolve(), "setup.yml", exv, env, *c.log)
	}
	return nil, nil
}

func fconsumesetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Consume setup for provider %s", p.Name)
		setupProviderEfOut := c.ef.Input.Children["setup_provider_"+p.Name].Output
		err, buffer := engine.GetBuffer(setupProviderEfOut, c.log, "provider:"+p.Name)
		if err != nil {
			return err, nil
		}
		// Keep a reference on the buffer based on the output folder
		c.buffer[setupProviderEfOut.Path()] = buffer
	}
	return nil, nil
}

func fcreate(c *InstallerContext) (error, cleanup) {

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		// unique id of the nodeset
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider.Resolve()

		// Provider setup exchange folder
		setupProviderEf := c.ef.Input.Children["setup_provider_"+p.Name]
		// We check if we have a buffer corresponding to the provider setup output
		buffer := c.getBuffer(setupProviderEf.Output)

		// Node creation exchange folder
		nodeCreateEf, e := c.ef.Input.AddChildExchangeFolder("create_" + n.Name)
		if e != nil {
			c.log.Printf(ERROR_ADDING_EXCHANGE_FOLDER, "create_"+n.Name, e.Error())
			return e, nil
		}
		e = nodeCreateEf.Create()
		if e != nil {
			c.log.Printf(ERROR_CREATING_EXCHANGE_FOLDER, "create_"+n.Name, e.Error())
			return e, nil
		}

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.Name, c.sshPublicKey, c.sshPrivateKey)
		bp.AddInt("instances", n.Instances)
		bp.AddNamedMap("params", p.Parameters)
		bp.AddInterface("volumes", n.Volumes)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		_, e = engine.SaveFile(c.log, *nodeCreateEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			return e, nil
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeCreateEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// Prepare extra vars
		ev := engine.BuildExtraVars("", *nodeCreateEf.Input, *nodeCreateEf.Output, buffer)

		// We launch the playbook
		ansible.LaunchPlayBook(c.lagoon.ComponentManager(), p.Component.Resolve(), "create.yml", ev, env, *c.log)
	}
	return nil, nil
}

func fconsumecreate(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume create for node %s", n.Name)
		nodeCreateEf := c.ef.Input.Children["create_"+n.Name].Output
		err, buffer := engine.GetBuffer(nodeCreateEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			return err, nil
		}
		c.buffer[nodeCreateEf.Path()] = buffer
	}
	return nil, nil
}

func fsetuporchestrator(c *InstallerContext) (error, cleanup) {

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		// unique id of the nodeset
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider.Resolve()

		// Provider setup exchange folder
		setupProviderEf := c.ef.Input.Children["setup_provider_"+p.Name]
		// We check if we have a buffer corresponding to the provider setup
		bufferPro := c.getBuffer(setupProviderEf.Output)

		// Node exchange folder
		nodeCreationEf := c.ef.Input.Children["create_"+n.Name]
		// We check if we have a buffer corresponding to the node output
		buffer := c.getBuffer(nodeCreationEf.Output)

		// Orchestrator setup exchange folder
		setupOrcherstratorEf, e := c.ef.Input.AddChildExchangeFolder("setup_orchestrator_" + n.Name)
		if e != nil {
			c.log.Printf(ERROR_ADDING_EXCHANGE_FOLDER, "setup_orchestrator_"+n.Name, e.Error())
			return e, nil
		}
		e = setupOrcherstratorEf.Create()
		if e != nil {
			c.log.Printf(ERROR_CREATING_EXCHANGE_FOLDER, "setup_orchestrator_"+n.Name, e.Error())
			return e, nil
		}

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.Name, c.sshPublicKey, c.sshPrivateKey)
		op := n.Orchestrator.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		_, e = engine.SaveFile(c.log, *setupOrcherstratorEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			return e, nil
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *setupOrcherstratorEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// ugly but .... TODO change this
		env.AddBuffer(bufferPro)

		// Prepare extra vars
		exv := engine.BuildExtraVars("", *setupOrcherstratorEf.Input, *setupOrcherstratorEf.Output, buffer)

		// We launch the playbook
		ansible.LaunchPlayBook(c.lagoon.ComponentManager(), c.lagoon.Environment().Orchestrator.Component.Resolve(), "setup.yml", exv, env, *c.log)
	}

	return nil, nil
}

func fconsumesetuporchestrator(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume orchestrator setup for node %s", n.Name)
		setupOrcherstratorEf := c.ef.Input.Children["setup_orchestrator_"+n.Name].Output
		err, buffer := engine.GetBuffer(setupOrcherstratorEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			return err, nil
		}
		c.buffer[setupOrcherstratorEf.Path()] = buffer
	}
	return nil, nil
}

func forchestrator(c *InstallerContext) (error, cleanup) {

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider.Resolve()

		// Provider setup exchange folder
		setupProviderEf := c.ef.Input.Children["setup_provider_"+p.Name]
		// We check if we have a buffer corresponding to the provider setup
		bufferPro := c.getBuffer(setupProviderEf.Output)

		// Orchestrator setup exchange folder
		setupOrcherstratorEf := c.ef.Input.Children["setup_orchestrator_"+n.Name]
		// We check if we have a buffer corresponding to the orchestrator setup
		buffer := c.getBuffer(setupOrcherstratorEf.Output)

		installOrcherstratorEf, e := c.ef.Input.AddChildExchangeFolder("install_orchestrator_" + n.Name)
		if e != nil {
			c.log.Printf(ERROR_ADDING_EXCHANGE_FOLDER, "install_orchestrator_"+n.Name, e.Error())
			return e, nil
		}
		e = installOrcherstratorEf.Create()
		if e != nil {
			c.log.Printf(ERROR_CREATING_EXCHANGE_FOLDER, "install_orchestrator_"+n.Name, e.Error())
			return e, nil
		}

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.Name, c.sshPublicKey, c.sshPrivateKey)
		bp.AddNamedMap("orchestrator", n.Orchestrator.OrchestratorParams())

		pr := c.lagoon.Environment().Providers["aws"].Proxy
		bp.AddInterface("proxy", pr)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		_, e = engine.SaveFile(c.log, *installOrcherstratorEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			return e, nil
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *installOrcherstratorEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// ugly but .... TODO change this
		env.AddBuffer(bufferPro)

		// Prepare extra vars
		exv := engine.BuildExtraVars("", *installOrcherstratorEf.Input, *installOrcherstratorEf.Output, buffer)

		// We launch the playbook
		ansible.LaunchPlayBook(c.lagoon.ComponentManager(), c.lagoon.Environment().Orchestrator.Component.Resolve(), "install.yml", exv, env, *c.log)
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
			path, err := engine.SaveFile(c.log, *c.ef.Output, VALIDATION_OUTPUT_FILE, b)
			if err != nil {
				// in case of error writing the report file
				return fmt.Errorf(ERROR_CREATING_REPORT_FILE, path), nil
			}

			if vErrs.HasErrors() {
				// in case of validation error we stop
				return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, ve.Error()), nil
			}
		}
	} else {
		c.log.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return nil, noCleanUpRequired
}

func flagoon(c *InstallerContext) (error, cleanup) {
	root, flavor := repositoryFlavor(c.location)
	if c.cliparams != nil {
		c.log.Printf("Creating lagoon environment with parameter for templating")
		c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", c.cliparams)
	} else {
		c.log.Printf("Creating lagoon environment without parameter for templating")
		c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", map[string]interface{}{})
	}

	if c.lagoonError == nil {
		c.lagoonError = c.lagoon.Init(root, flavor) // FIXME: really need custom descriptor name ?
	}
	c.log.Printf("Created environment providers %v \n", c.lagoon.Environment().Providers)
	return nil, noCleanUpRequired
}

func fcliparam(c *InstallerContext) (error, cleanup) {
	ok := c.ef.Location.Contains(engine.CliParametersFileName)
	if ok {

		p, e := engine.ParseParams(engine.JoinPaths(c.ef.Location.Path(), engine.CliParametersFileName))
		if e != nil {
			return fmt.Errorf(ERROR_LOADING_CLI_PARAMETERS, e), nil
		}
		c.cliparams = p
		c.log.Printf(LOG_CLI_PARAMS, c.cliparams)
	}
	return nil, noCleanUpRequired
}

//repositoryFlavor returns the repository flavor, branchn tag ..., based on the
// presence of '@' into the given url
func repositoryFlavor(url string) (string, string) {

	if strings.Contains(url, "@") {
		s := strings.Split(url, "@")
		return s[0], s[1]
	}
	return url, ""
}

func ffailOnLagoonError(c *InstallerContext) (error, cleanup) {
	if c.lagoonError != nil {
		vErrs, ok := c.lagoonError.(model.ValidationErrors)
		if ok {
			if vErrs.HasErrors() {
				// in case of validation error we stop
				c.log.Println(c.lagoonError)
				return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.lagoonError.Error()), noCleanUpRequired
			}
		} else {
			return fmt.Errorf(ERROR_PARSING_ENVIRONMENT, c.lagoonError.Error()), nil
		}
	}
	return nil, noCleanUpRequired
}
