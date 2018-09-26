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
	rep := launch(calls, c)
	e = rep.Error
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
		flogCheck,
	}
	rep := launch(calls, c)
	e = rep.Error
	return
}

func fsession(c *InstallerContext) stepContexts {
	sc := InitStepContext("Checking the execution session", nil, noCleanUpRequired)
	// Check if a session already exists
	var createSession *engine.CreationSession

	b, s := engine.HasCreationSession(*c.ef)
	if !b {
		createSession = &engine.CreationSession{Client: c.lagoon.Environment().QualifiedName().String(), Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	// If needed creates the missing Uids for the nodes
	for _, n := range c.lagoon.Environment().NodeSets {
		if val, ok := createSession.Uids[n.Name]; ok {
			c.log.Printf(LOG_REUSING_UID_FOR_CLIENT, val, c.lagoon.Environment().QualifiedName(), n.Name)
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.lagoon.Environment().QualifiedName(), n.Name)
			createSession.Add(n.Name, uid)
		}
	}
	by, e := createSession.Content()
	if e != nil {
		sc.Err = e
		sc.ErrDetail = fmt.Sprintf("An error occured marshalling the session content :%v", createSession)
		goto MoveOut
	}
	{
		f, e := engine.SaveFile(c.log, *c.ef.Location, engine.CreationSessionFileName, by)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the session file into :%v", c.ef.Location.Path())
			goto MoveOut
		}
		c.session = &engine.EngineSession{
			CreationSession: createSession,
			File:            f,
		}
	}
MoveOut:
	return sc.Array()
}

func fsetup(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, p := range c.lagoon.Environment().Providers {
		sc := InitStepContext("Running the setup phase", p, noCleanUpRequired)
		c.log.Printf(LOG_RUNNING_SETUP_FOR, p.Name)

		// Provider setup exchange folder
		setupProviderEf, e := c.ef.Input.AddChildExchangeFolder("setup_provider_" + p.Name)
		if e != nil {
			err := fmt.Errorf(ERROR_ADDING_EXCHANGE_FOLDER, "setup_provider_"+p.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}
		e = setupProviderEf.Create()
		if e != nil {
			err := fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "setup_provider_"+p.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}
		setupProviderEfIn := setupProviderEf.Input
		setupProviderEfOut := setupProviderEf.Output

		// Prepare parameters
		bp := c.BuildBaseParam("", p.Name)
		bp.AddNamedMap("params", p.Parameters)
		b, e := bp.Content()
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured creating the base parameters")
			sCs.Add(sc)
			continue
		}

		_, e = engine.SaveFile(c.log, *setupProviderEfIn, engine.ParamYamlFileName, b)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the parameter file into :%v", setupProviderEfIn.Path())
			sCs.Add(sc)
			continue
		}

		// This is the first "real" step of the process so the used buffer is empty
		emptyBuff := engine.CreateBuffer()

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *setupProviderEfIn)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the components file into :%v", setupProviderEfIn.Path())
			sCs.Add(sc)
			continue
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
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumesetup(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, p := range c.lagoon.Environment().Providers {
		sc := InitStepContext("Consuming the setup phase", p, noCleanUpRequired)
		c.log.Printf("Consume setup for provider %s", p.Name)
		setupProviderEfOut := c.ef.Input.Children["setup_provider_"+p.Name].Output
		err, buffer := engine.GetBuffer(setupProviderEfOut, c.log, "provider:"+p.Name)
		if err != nil {
			sc.Err = err
			sc.ErrDetail = fmt.Sprintf("An error occured getting the buffer")
			sCs.Add(sc)
			continue
		}
		// Keep a reference on the buffer based on the output folder
		c.buffer[setupProviderEfOut.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func fcreate(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.lagoon.Environment().NodeSets {
		sc := InitStepContext("Running the create phase", n, noCleanUpRequired)
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
			err := fmt.Errorf(ERROR_ADDING_EXCHANGE_FOLDER, "create_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}
		e = nodeCreateEf.Create()
		if e != nil {
			err := fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "create_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		bp := c.BuildBaseParam(uid, p.Name)
		bp.AddInt("instances", n.Instances)
		bp.AddNamedMap("params", p.Parameters)
		bp.AddInterface("volumes", n.Volumes)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured creating the base parameters")
			sCs.Add(sc)
			continue
		}
		_, e = engine.SaveFile(c.log, *nodeCreateEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the parameter file into :%v", nodeCreateEf.Input.Path())
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeCreateEf.Input)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the components file into :%v", nodeCreateEf.Input.Path())
			sCs.Add(sc)
			continue
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
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumecreate(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.lagoon.Environment().NodeSets {
		sc := InitStepContext("Consuming the create phase", n, noCleanUpRequired)
		c.log.Printf("Consume create for node %s", n.Name)
		nodeCreateEf := c.ef.Input.Children["create_"+n.Name].Output
		err, buffer := engine.GetBuffer(nodeCreateEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			sc.Err = err
			sc.ErrDetail = fmt.Sprintf("An error occured getting the buffer")
			sCs.Add(sc)
			continue
		}
		c.buffer[nodeCreateEf.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func fsetuporchestrator(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.lagoon.Environment().NodeSets {
		sc := InitStepContext("Running the orchestrator setup phase", n, noCleanUpRequired)
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
			err := fmt.Errorf(ERROR_ADDING_EXCHANGE_FOLDER, "setup_orchestrator_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue

		}
		e = setupOrcherstratorEf.Create()
		if e != nil {
			err := fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "setup_orchestrator_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		//bp := engine.BuilBaseParam(c.lagoon.Environment().QualifiedName(), uid, p.Name, c.sshPublicKey, c.sshPrivateKey)
		bp := c.BuildBaseParam(uid, p.Name)
		op := n.Orchestrator.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured creating the base parameters")
			sCs.Add(sc)
			continue
		}
		_, e = engine.SaveFile(c.log, *setupOrcherstratorEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the parameter file into :%v", setupOrcherstratorEf.Input.Path())
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *setupOrcherstratorEf.Input)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the components file into :%v", setupOrcherstratorEf.Input.Path())
			sCs.Add(sc)
			continue
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
		sCs.Add(sc)
	}

	return *sCs
}

func fconsumesetuporchestrator(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.lagoon.Environment().NodeSets {
		sc := InitStepContext("Consuming the orchestrator setup phase", n, noCleanUpRequired)
		c.log.Printf("Consume orchestrator setup for node %s", n.Name)
		setupOrcherstratorEf := c.ef.Input.Children["setup_orchestrator_"+n.Name].Output
		err, buffer := engine.GetBuffer(setupOrcherstratorEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			sc.Err = err
			sc.ErrDetail = fmt.Sprintf("An error occured getting the buffer")
			sCs.Add(sc)
			continue
		}
		c.buffer[setupOrcherstratorEf.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func forchestrator(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.lagoon.Environment().NodeSets {
		sc := InitStepContext("Running the orchestrator installation phase", n, noCleanUpRequired)
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
			err := fmt.Errorf(ERROR_ADDING_EXCHANGE_FOLDER, "install_orchestrator_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}
		e = installOrcherstratorEf.Create()
		if e != nil {
			err := fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, "install_orchestrator_"+n.Name, e.Error())
			c.log.Printf(err.Error())
			sc.Err = err
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		bp := c.BuildBaseParam(uid, p.Name)
		bp.AddNamedMap("orchestrator", n.Orchestrator.OrchestratorParams())

		// TODO removed this hardcoded AWS
		pr := c.lagoon.Environment().Providers["aws"].Proxy
		bp.AddInterface("proxy", pr)
		bp.AddBuffer(buffer)

		b, e := bp.Content()
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured creating the base parameters")
			sCs.Add(sc)
			continue
		}
		_, e = engine.SaveFile(c.log, *installOrcherstratorEf.Input, engine.ParamYamlFileName, b)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the parameter file into :%v", installOrcherstratorEf.Input.Path())
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *installOrcherstratorEf.Input)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the components file into :%v", installOrcherstratorEf.Input.Path())
			sCs.Add(sc)
			continue
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
		sCs.Add(sc)
	}
	return *sCs
}

func flogCheck(c *InstallerContext) stepContexts {
	sc := InitStepContext("Validating the environment content", nil, noCleanUpRequired)
	ve := c.lagoonError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			sc.Err = fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error())
		} else {
			c.log.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				sc.Err = fmt.Errorf(ERROR_GENERIC, e)
			}
			// print both errors and warnings into the report file
			path, err := engine.SaveFile(c.log, *c.ef.Output, VALIDATION_OUTPUT_FILE, b)
			if err != nil {
				// in case of error writing the report file
				sc.Err = fmt.Errorf(ERROR_CREATING_REPORT_FILE, path)
			}

			if vErrs.HasErrors() {
				// in case of validation error we stop
				sc.Err = fmt.Errorf(ERROR_PARSING_DESCRIPTOR, ve.Error())
			}
		}
	} else {
		c.log.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return sc.Array()
}

func flagoon(c *InstallerContext) stepContexts {
	sc := InitStepContext("Creating the environment based on the descriptor", nil, noCleanUpRequired)
	root, flavor := repositoryFlavor(c.location)
	if c.cliparams != nil {
		c.log.Printf("Creating lagoon environment with parameter for templating")
		c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", c.cliparams)
	} else {
		c.log.Printf("Creating lagoon environment without parameter for templating")
		c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", map[string]interface{}{})
	}

	if c.lagoonError == nil {
		c.lagoonError = c.lagoon.Init(root, flavor)
	}
	// Note: here we are not taking in account the "c.lagoonError != nil" to place the error
	// into the context because this error managment depends on the whole process
	// and will be treated if required by the "ffailOnLagoonError" step
	return sc.Array()
}

func fcliparam(c *InstallerContext) stepContexts {
	sc := InitStepContext("Reading substitution parameters", nil, noCleanUpRequired)
	ok := c.ef.Location.Contains(engine.CliParametersFileName)
	if ok {
		p, e := engine.ParseParams(engine.JoinPaths(c.ef.Location.Path(), engine.CliParametersFileName))
		if e != nil {
			sc.Err = fmt.Errorf(ERROR_LOADING_CLI_PARAMETERS, e)
			goto MoveOut
		}
		c.cliparams = p
		c.log.Printf(LOG_CLI_PARAMS, c.cliparams)
	}
MoveOut:
	return sc.Array()
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

func ffailOnLagoonError(c *InstallerContext) stepContexts {
	sc := InitStepContext("Stopping the process in case of validation errors", nil, noCleanUpRequired)
	if c.lagoonError != nil {
		vErrs, ok := c.lagoonError.(model.ValidationErrors)
		if ok {
			if vErrs.HasErrors() {
				// in case of validation error we stop
				c.log.Println(c.lagoonError)
				sc.Err = fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.lagoonError.Error())
				goto MoveOut
			}
		} else {
			sc.Err = fmt.Errorf(ERROR_PARSING_ENVIRONMENT, c.lagoonError.Error())
			goto MoveOut
		}
	}
MoveOut:
	return sc.Array()
}
