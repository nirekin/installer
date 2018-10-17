package installer

import (
	"fmt"
	"os"
	"strings"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
	"github.com/ekara-platform/model"
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

func Run(c *InstallerContext) (e error) {
	// Check if the received action is supporter by the engine
	c.log.Println(LOG_RUNNING)
	a := os.Getenv(util.ActionEnvVariableKey)
	switch a {
	case engine.ActionCreate.String():
		c.log.Println(LOG_ACTION_CREATE)
		e = writeReport(runCreate(c))
	case engine.ActionCheck.String():
		c.log.Println(LOG_ACTION_CHECK)
		e = writeReport(runCheck(c))
	default:
		if a == "" {
			a = LOG_NO_ACTION
		}
		e = fmt.Errorf(ERROR_UNSUPORTED_ACTION, a)
	}
	return
}

// runCreate launches the environment creation
func runCreate(c *InstallerContext) ExecutionReport {
	// Stack of functions required to create an environment
	calls := []step{
		fproxy,
		fexchangeFoldef,
		flocation,
		fcliparam,
		fekara,
		ffailOnEkaraError,
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
	return launch(calls, c)
}

// runCheck launches the environment check
func runCheck(c *InstallerContext) ExecutionReport {
	// Stack of functions required to check an environment
	calls := []step{
		fproxy,
		fexchangeFoldef,
		flocation,
		fcliparam,
		fekara,
		flogCheck,
	}
	return launch(calls, c)
}

func fsession(c *InstallerContext) stepContexts {
	sc := InitStepContext("Checking the execution session", nil, noCleanUpRequired)
	// Check if a session already exists
	var createSession *engine.CreationSession

	b, s := engine.HasCreationSession(*c.ef)
	if !b {
		createSession = &engine.CreationSession{Client: c.ekara.Environment().QualifiedName().String(), Uids: make(map[string]string)}
	} else {
		createSession = s.CreationSession
	}

	// If needed creates the missing Uids for the nodes
	for _, n := range c.ekara.Environment().NodeSets {
		if val, ok := createSession.Uids[n.Name]; ok {
			c.log.Printf(LOG_REUSING_UID_FOR_CLIENT, val, c.ekara.Environment().QualifiedName(), n.Name)
		} else {
			uid := engine.GetUId()
			c.log.Printf(LOG_CREATING_UID_FOR_CLIENT, uid, c.ekara.Environment().QualifiedName(), n.Name)
			createSession.Add(n.Name, uid)
		}
	}
	by, e := createSession.Content()
	if e != nil {
		InstallerFail(&sc, e, fmt.Sprintf("An error occured marshalling the session content :%v", createSession))
		goto MoveOut
	}
	{
		f, e := util.SaveFile(c.log, *c.ef.Location, util.CreationSessionFileName, by)
		if e != nil {
			InstallerFail(&sc, e, fmt.Sprintf("An error occured saving the session file into :%v", c.ef.Location.Path()))
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
	for _, p := range c.ekara.Environment().Providers {
		sc := InitStepContext("Running the setup phase", p, noCleanUpRequired)
		c.log.Printf(LOG_RUNNING_SETUP_FOR, p.Name)

		// Provider setup exchange folder
		setupProviderEf, ko := createChildExchangeFolder(c.ef.Input, "setup_provider_"+p.Name, &sc, c.log)
		if ko {
			sCs.Add(sc)
			continue
		}

		setupProviderEfIn := setupProviderEf.Input
		setupProviderEfOut := setupProviderEf.Output

		// Prepare parameters
		bp := c.BuildBaseParam("", p.Name)
		bp.AddNamedMap("params", p.Parameters)

		if ko := saveBaseParams(bp, c, setupProviderEfIn, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// This is the first "real" step of the process so the used buffer is empty
		emptyBuff := ansible.CreateBuffer()

		// Prepare components map
		if ko := saveComponentMap(c, setupProviderEfIn, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", *setupProviderEfIn, *setupProviderEfOut, emptyBuff)

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)

		// Adding the environment variables from the provider
		for envK, envV := range p.EnvVars {
			env.Add(envK, envV)
		}

		// We launch the playbook
		err, code := c.ekara.AnsibleManager().Execute(p.Component.Resolve(), "setup.yml", exv, env, "")
		if err != nil {
			pEo := playBookErrorOrigin{
				Playbook:  "setup.yml",
				Compoment: p.Component.Resolve().Id,
				Code:      code,
			}
			sc.Error = err
			sc.ErrorDetail = "An error occured executing the playbook"
			sc.ErrorOrigin = pEo
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumesetup(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, p := range c.ekara.Environment().Providers {
		sc := InitStepContext("Consuming the setup phase", p, noCleanUpRequired)
		c.log.Printf("Consume setup for provider %s", p.Name)
		setupProviderEfOut := c.ef.Input.Children["setup_provider_"+p.Name].Output
		err, buffer := ansible.GetBuffer(setupProviderEfOut, c.log, "provider:"+p.Name)
		if err != nil {
			InstallerFail(&sc, err, fmt.Sprintf("An error occured getting the buffer"))
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
	for _, n := range c.ekara.Environment().NodeSets {
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
		nodeCreateEf, ko := createChildExchangeFolder(c.ef.Input, "create_"+n.Name, &sc, c.log)
		if ko {
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		bp := c.BuildBaseParam(uid, p.Name)
		bp.AddInt("instances", n.Instances)
		bp.AddNamedMap("params", p.Parameters)
		bp.AddInterface("volumes", n.Volumes)
		bp.AddBuffer(buffer)

		if ko := saveBaseParams(bp, c, nodeCreateEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		if ko := saveComponentMap(c, nodeCreateEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", *nodeCreateEf.Input, *nodeCreateEf.Output, buffer)

		inventory := ""
		if len(buffer.Inventories) > 0 {
			inventory = buffer.Inventories["inventory_path"]
		}

		// We launch the playbook
		err, code := c.ekara.AnsibleManager().Execute(p.Component.Resolve(), "create.yml", exv, env, inventory)
		if err != nil {
			pEo := playBookErrorOrigin{
				Playbook:  "create.yml",
				Compoment: p.Component.Resolve().Id,
				Code:      code,
			}
			sc.Error = err
			sc.ErrorDetail = "An error occured executing the playbook"
			sc.ErrorOrigin = pEo
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumecreate(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.ekara.Environment().NodeSets {
		sc := InitStepContext("Consuming the create phase", n, noCleanUpRequired)
		c.log.Printf("Consume create for node %s", n.Name)
		nodeCreateEf := c.ef.Input.Children["create_"+n.Name].Output
		err, buffer := ansible.GetBuffer(nodeCreateEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			InstallerFail(&sc, err, fmt.Sprintf("An error occured getting the buffer"))
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
	for _, n := range c.ekara.Environment().NodeSets {
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
		setupOrcherstratorEf, ko := createChildExchangeFolder(c.ef.Input, "setup_orchestrator_"+n.Name, &sc, c.log)
		if ko {
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		//bp := engine.BuilBaseParam(c.ekara.Environment().QualifiedName(), uid, p.Name, c.sshPublicKey, c.sshPrivateKey)
		bp := c.BuildBaseParam(uid, p.Name)
		op := n.Orchestrator.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddBuffer(buffer)

		if ko := saveBaseParams(bp, c, setupOrcherstratorEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		if ko := saveComponentMap(c, setupOrcherstratorEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// ugly but .... TODO change this
		env.AddBuffer(bufferPro)

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", *setupOrcherstratorEf.Input, *setupOrcherstratorEf.Output, buffer)

		inventory := ""
		if len(bufferPro.Inventories) > 0 {
			inventory = bufferPro.Inventories["inventory_path"]
		}

		// We launch the playbook
		err, code := c.ekara.AnsibleManager().Execute(c.ekara.Environment().Orchestrator.Component.Resolve(), "setup.yml", exv, env, inventory)
		if err != nil {
			pEo := playBookErrorOrigin{
				Playbook:  "setup.yml",
				Compoment: p.Component.Resolve().Id,
				Code:      code,
			}
			sc.Error = err
			sc.ErrorDetail = "An error occured executing the playbook"
			sc.ErrorOrigin = pEo
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}

	return *sCs
}

func fconsumesetuporchestrator(c *InstallerContext) stepContexts {
	sCs := InitStepContexts()
	for _, n := range c.ekara.Environment().NodeSets {
		sc := InitStepContext("Consuming the orchestrator setup phase", n, noCleanUpRequired)
		c.log.Printf("Consume orchestrator setup for node %s", n.Name)
		setupOrcherstratorEf := c.ef.Input.Children["setup_orchestrator_"+n.Name].Output
		err, buffer := ansible.GetBuffer(setupOrcherstratorEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			InstallerFail(&sc, err, fmt.Sprintf("An error occured getting the buffer"))
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
	for _, n := range c.ekara.Environment().NodeSets {
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

		installOrcherstratorEf, ko := createChildExchangeFolder(c.ef.Input, "install_orchestrator_"+n.Name, &sc, c.log)
		if ko {
			sCs.Add(sc)
			continue
		}

		// Prepare parameters
		bp := c.BuildBaseParam(uid, p.Name)
		bp.AddNamedMap("orchestrator", n.Orchestrator.OrchestratorParams())

		// TODO removed this hardcoded AWS
		pr := c.ekara.Environment().Providers["aws"].Proxy
		bp.AddInterface("proxy", pr)
		bp.AddBuffer(buffer)

		if ko := saveBaseParams(bp, c, installOrcherstratorEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare components map
		if ko := saveComponentMap(c, installOrcherstratorEf.Input, &sc); ko {
			sCs.Add(sc)
			continue
		}

		// Prepare environment variables
		env := ansible.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer)

		// ugly but .... TODO change this
		env.AddBuffer(bufferPro)

		// Prepare extra vars
		exv := ansible.BuildExtraVars("", *installOrcherstratorEf.Input, *installOrcherstratorEf.Output, buffer)

		inventory := ""
		if len(bufferPro.Inventories) > 0 {
			inventory = bufferPro.Inventories["inventory_path"]
		}

		// We launch the playbook
		err, code := c.ekara.AnsibleManager().Execute(c.ekara.Environment().Orchestrator.Component.Resolve(), "install.yml", exv, env, inventory)
		if err != nil {
			pEo := playBookErrorOrigin{
				Playbook:  "install.yml",
				Compoment: p.Component.Resolve().Id,
				Code:      code,
			}
			sc.Error = err
			sc.ErrorDetail = "An error occured executing the playbook"
			sc.ErrorOrigin = pEo
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func flogCheck(c *InstallerContext) stepContexts {
	sc := InitStepContext("Validating the environment content", nil, noCleanUpRequired)
	ve := c.ekaraError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			DescriptorFail(&sc, fmt.Errorf(ERROR_PARSING_ENVIRONMENT, ve.Error()), "")
		} else {
			c.log.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				DescriptorFail(&sc, fmt.Errorf(ERROR_GENERIC, e), "")
			}
			// print both errors and warnings into the report file
			path, err := util.SaveFile(c.log, *c.ef.Output, VALIDATION_OUTPUT_FILE, b)
			if err != nil {
				// in case of error writing the report file
				DescriptorFail(&sc, fmt.Errorf(ERROR_CREATING_REPORT_FILE, path), "")
			}

			if vErrs.HasErrors() {
				// in case of validation error we stop
				DescriptorFail(&sc, fmt.Errorf(ERROR_PARSING_DESCRIPTOR, ve.Error()), "")
			}
		}
	} else {
		c.log.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return sc.Array()
}

func fekara(c *InstallerContext) stepContexts {
	sc := InitStepContext("Creating the environment based on the descriptor", nil, noCleanUpRequired)
	root, flavor := repositoryFlavor(c.location)
	if c.cliparams != nil {
		c.log.Printf("Creating ekara environment with parameter for templating")
		c.ekara, c.ekaraError = engine.Create(c.log, "/var/lib/ekara", c.cliparams)
	} else {
		c.log.Printf("Creating ekara environment without parameter for templating")
		c.ekara, c.ekaraError = engine.Create(c.log, "/var/lib/ekara", map[string]interface{}{})
	}

	if c.ekaraError == nil {
		c.ekaraError = c.ekara.Init(root, flavor)
	}
	// Note: here we are not taking in account the "c.ekaraError != nil" to place the error
	// into the context because this error managment depends on the whole process
	// and will be treated if required by the "ffailOnEkaraError" step
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

func ffailOnEkaraError(c *InstallerContext) stepContexts {
	sc := InitStepContext("Stopping the process in case of validation errors", nil, noCleanUpRequired)
	if c.ekaraError != nil {
		vErrs, ok := c.ekaraError.(model.ValidationErrors)
		if ok {
			if vErrs.HasErrors() {
				// in case of validation error we stop
				c.log.Println(c.ekaraError)
				DescriptorFail(&sc, fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.ekaraError.Error()), "")
				goto MoveOut
			}
		} else {
			DescriptorFail(&sc, fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.ekaraError.Error()), "")
			goto MoveOut
		}
	}
MoveOut:
	return sc.Array()
}
