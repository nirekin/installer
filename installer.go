package installer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		fqualifiedName,
		fexchangeFoldef,
		flocation,
		fcliparam,
		freport,
		fekara,
		ffailOnEkaraError,
		fSHKeys,
		fsetup,
		fconsumesetup,
		fcreate,
		fconsumecreate,
		fsetuporchestrator,
		fconsumesetuporchestrator,
		forchestrator,
		fstack,
	}
	return launch(calls, c)
}

// runCheck launches the environment check
func runCheck(c *InstallerContext) ExecutionReport {
	// Stack of functions required to check an environment
	calls := []step{
		fproxy,
		fqualifiedName,
		fexchangeFoldef,
		flocation,
		fcliparam,
		fekara,
		flogCheck,
	}
	return launch(calls, c)
}

func fsetup(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, p := range c.ekara.ComponentManager().Environment().Providers {
		sc := InitPlaybookStepResult("Running the setup phase", p, noCleanUpRequired)
		c.log.Printf(LOG_RUNNING_SETUP_FOR, p.Name)

		// TEST FAILURE FOR THE --limit addition
		//hf, _ := c.report.hasFailure()

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
		r, err := p.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
			sCs.Add(sc)
			continue
		}
		err, code := c.ekara.AnsibleManager().Execute(r, "setup.yml", exv, env, "")

		if err != nil {
			r, err := p.Component.Resolve()
			if err != nil {
				FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
				sCs.Add(sc)
				continue
			}
			pfd := playBookFailureDetail{
				Playbook:  "setup.yml",
				Compoment: r.Id,
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occured executing the playbook", pfd)
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumesetup(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, p := range c.ekara.ComponentManager().Environment().Providers {
		sc := InitCodeStepResult("Consuming the setup phase", p, noCleanUpRequired)
		c.log.Printf("Consume setup for provider %s", p.Name)
		setupProviderEfOut := c.ef.Input.Children["setup_provider_"+p.Name].Output
		err, buffer := ansible.GetBuffer(setupProviderEfOut, c.log, "provider:"+p.Name)
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured getting the buffer"), nil)
			sCs.Add(sc)
			continue
		}
		// Keep a reference on the buffer based on the output folder
		c.buffer[setupProviderEfOut.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func fcreate(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the create phase", n, noCleanUpRequired)
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)

		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the nodeset"), nil)
			sCs.Add(sc)
			continue
		}

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
		bp := c.BuildBaseParam(n.Name, p.Name)
		bp.AddInt("instances", n.Instances)
		bp.AddInterface("labels", n.Labels)
		bp.AddNamedMap("params", p.Parameters)
		bp.AddInterface("volumes", n.Volumes.AsArray())
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
		r, err := p.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
			sCs.Add(sc)
			continue
		}
		err, code := c.ekara.AnsibleManager().Execute(r, "create.yml", exv, env, inventory)
		if err != nil {
			pfd := playBookFailureDetail{
				Playbook:  "create.yml",
				Compoment: r.Id,
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occured executing the playbook", pfd)
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fconsumecreate(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitCodeStepResult("Consuming the create phase", n, noCleanUpRequired)
		c.log.Printf("Consume create for node %s", n.Name)
		nodeCreateEf := c.ef.Input.Children["create_"+n.Name].Output
		err, buffer := ansible.GetBuffer(nodeCreateEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured getting the buffer"), nil)
			sCs.Add(sc)
			continue
		}
		c.buffer[nodeCreateEf.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func fsetuporchestrator(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the orchestrator setup phase", n, noCleanUpRequired)
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)

		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the nodeset"), nil)
			sCs.Add(sc)
			continue
		}

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
		bp := c.BuildBaseParam(n.Name, p.Name)
		op, err := n.Orchestrator.OrchestratorParams()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured getting the orchestrator parameters"), nil)
			sCs.Add(sc)
			continue
		}
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
		//TODO move this resolbe outside of the for loop
		r, err := c.ekara.ComponentManager().Environment().Orchestrator.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the orchestrator"), nil)
			sCs.Add(sc)
			continue
		}
		err, code := c.ekara.AnsibleManager().Execute(r, "setup.yml", exv, env, inventory)
		if err != nil {
			r, err := p.Component.Resolve()
			if err != nil {
				FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
				sCs.Add(sc)
				continue
			}
			pfd := playBookFailureDetail{
				Playbook:  "setup.yml",
				Compoment: r.Id,
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occured executing the playbook", pfd)
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}

	return *sCs
}

func fconsumesetuporchestrator(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitCodeStepResult("Consuming the orchestrator setup phase", n, noCleanUpRequired)
		c.log.Printf("Consume orchestrator setup for node %s", n.Name)
		setupOrcherstratorEf := c.ef.Input.Children["setup_orchestrator_"+n.Name].Output
		err, buffer := ansible.GetBuffer(setupOrcherstratorEf, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured getting the buffer"), nil)
			sCs.Add(sc)
			continue
		}
		c.buffer[setupOrcherstratorEf.Path()] = buffer
		sCs.Add(sc)
	}
	return *sCs
}

func forchestrator(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the orchestrator installation phase", n, noCleanUpRequired)
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)

		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the nodeset"), nil)
			sCs.Add(sc)
			continue
		}

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
		bp := c.BuildBaseParam(n.Name, p.Name)
		bp.AddInterface("labels", n.Labels)
		op, err := n.Orchestrator.OrchestratorParams()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured getting the orchestrator parameters"), nil)
			sCs.Add(sc)
			continue
		}
		bp.AddNamedMap("orchestrator", op)

		// TODO check how to clean all proxies
		pr := c.ekara.ComponentManager().Environment().Providers[p.Name].Proxy
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
		//TODO move this resolve outside of the for loop
		r, err := c.ekara.ComponentManager().Environment().Orchestrator.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the orchestrator"), nil)
			sCs.Add(sc)
			continue
		}
		err, code := c.ekara.AnsibleManager().Execute(r, "install.yml", exv, env, inventory)
		if err != nil {
			r, err := p.Component.Resolve()
			if err != nil {
				FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
				sCs.Add(sc)
				continue
			}
			pfd := playBookFailureDetail{
				Playbook:  "install.yml",
				Compoment: r.Id,
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occured executing the playbook", pfd)
			sCs.Add(sc)
			continue
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fstack(c *InstallerContext) stepResults {
	sCs := InitStepResults()
	for _, s := range c.ekara.ComponentManager().Environment().Stacks {
		sc := InitCodeStepResult("Starting a stack setup phase", s, noCleanUpRequired)
		c.log.Printf("Checking how to install %s", s.Name)
		// Check if the stacks holds an "install.yml" playbook
		r, err := s.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the stack"), nil)
			sCs.Add(sc)
			continue
		}
		if ok := c.ekara.AnsibleManager().Contains(r, "install.yml"); ok {
			fstackPlabook(c, s, sCs)
		} else {
			fstackCompose(c, s, sCs)
		}
		sCs.Add(sc)
	}
	return *sCs
}

func fstackPlabook(c *InstallerContext, s model.Stack, sCs *stepResults) {
	for _, n := range c.ekara.ComponentManager().Environment().NodeSets {
		sc := InitPlaybookStepResult("Running the stack playbook installation phase", model.ChainDescribable(s, n), noCleanUpRequired)

		// Stack install exchange folder
		fName := fmt.Sprintf("install_stack_%s_on_%s", s.Name, n.Name)

		p, err := n.Provider.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the nodeset"), nil)
			sCs.Add(sc)
			continue
		}

		// Provider setup exchange folder
		setupProviderEf := c.ef.Input.Children["setup_provider_"+p.Name]
		// We check if we have a buffer corresponding to the provider setup
		bufferPro := c.getBuffer(setupProviderEf.Output)

		// We use an empty buffer because no one is coming from the previous step
		buffer := ansible.CreateBuffer()

		stackEf, ko := createChildExchangeFolder(c.ef.Input, fName, &sc, c.log)
		if ko {
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
		exv := ansible.BuildExtraVars("", *stackEf.Input, *stackEf.Output, buffer)

		inventory := ""
		if len(bufferPro.Inventories) > 0 {
			inventory = bufferPro.Inventories["inventory_path"]
		}

		// We launch the playbook
		r, err := s.Component.Resolve()
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the stack"), nil)
			sCs.Add(sc)
			continue
		}
		err, code := c.ekara.AnsibleManager().Execute(r, "install.yml", exv, env, inventory)
		if err != nil {
			r, err := p.Component.Resolve()
			if err != nil {
				FailsOnCode(&sc, err, fmt.Sprintf("An error occured resolving the provider"), nil)
				sCs.Add(sc)
				continue
			}
			pfd := playBookFailureDetail{
				Playbook:  "install.yml",
				Compoment: r.Id,
				Code:      code,
			}
			FailsOnPlaybook(&sc, err, "An error occured executing the playbook", pfd)
			sCs.Add(sc)
			continue
		}
		if ko {
			sCs.Add(sc)
			continue
		}

		sCs.Add(sc)
	}
}

func fstackCompose(c *InstallerContext, s model.Stack, sCs *stepResults) {
	for _, n := range c.ekara.ComponentManager().Environment().Stacks {
		sc := InitPlaybookStepResult("Running the stack Docker Compose installation phase", model.ChainDescribable(s, n), noCleanUpRequired)
		c.log.Printf(LOG_PROCESSING_STACK_COMPOSE, s.Name, n.Name)
		err := fmt.Errorf("The Ekara plaform is not ready yet to install stacks using Docker Compose")
		FailsOnNotImplemented(&sc, err, "", nil)
		sCs.Add(sc)
	}
}

func flogCheck(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Validating the environment content", nil, noCleanUpRequired)
	ve := c.ekaraError
	if ve != nil {
		vErrs, ok := ve.(model.ValidationErrors)
		// if the error is not a "validation error" then we return it
		if !ok {
			FailsOnDescriptor(&sc, ve, fmt.Sprintf(ERROR_PARSING_ENVIRONMENT, ve.Error()), nil)
		} else {
			c.log.Printf(ve.Error())
			b, e := vErrs.JSonContent()
			if e != nil {
				FailsOnDescriptor(&sc, e, fmt.Sprintf(ERROR_GENERIC, e), nil)
			}
			// print both errors and warnings into the report file
			path, err := util.SaveFile(c.log, *c.ef.Output, VALIDATION_OUTPUT_FILE, b)
			if err != nil {
				// in case of error writing the report file
				FailsOnDescriptor(&sc, err, fmt.Sprintf(ERROR_CREATING_REPORT_FILE, path), nil)
			}

			if vErrs.HasErrors() {
				// in case of validation error we stop
				FailsOnDescriptor(&sc, ve, fmt.Sprintf(ERROR_PARSING_DESCRIPTOR, ve.Error()), nil)
			}
		}
	} else {
		c.log.Printf(LOG_VALIDATION_SUCCESSFUL)
	}
	return sc.Array()
}

func fekara(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Creating the environment based on the descriptor", nil, noCleanUpRequired)
	root, flavor := repositoryFlavor(c.location)
	if c.cliparams != nil {
		c.log.Printf("Creating ekara environment with parameter for templating")
		c.ekara, c.ekaraError = engine.Create(c.log, "/var/lib/ekara", c.cliparams)
	} else {
		c.log.Printf("Creating ekara environment without parameter for templating")
		c.ekara, c.ekaraError = engine.Create(c.log, "/var/lib/ekara", map[string]interface{}{})
	}

	if c.ekaraError == nil {
		c.ekaraError = c.ekara.Init(root, flavor, c.name)
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

func ffailOnEkaraError(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Stopping the process in case of validation errors", nil, noCleanUpRequired)
	if c.ekaraError != nil {
		vErrs, ok := c.ekaraError.(model.ValidationErrors)
		if ok {
			if vErrs.HasErrors() {
				// in case of validation error we stop
				c.log.Println(c.ekaraError)
				FailsOnDescriptor(&sc, c.ekaraError, fmt.Sprintf(ERROR_PARSING_DESCRIPTOR, c.ekaraError.Error()), nil)
				goto MoveOut
			}
		} else {
			FailsOnDescriptor(&sc, c.ekaraError, fmt.Sprintf(ERROR_PARSING_DESCRIPTOR, c.ekaraError.Error()), nil)
			goto MoveOut
		}
	}
MoveOut:
	return sc.Array()
}

// fqualifiedName extracts the qualified environment name from the
// environment variable "engine.StarterEnvQualifiedVariableKey"
func fqualifiedName(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Reading the descriptor location", nil, noCleanUpRequired)
	c.qualifiedName = os.Getenv(util.StarterEnvQualifiedVariableKey)
	if c.qualifiedName == "" {
		FailsOnCode(&sc, fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvQualifiedVariableKey), "", nil)
		goto MoveOut
	}
MoveOut:
	return sc.Array()
}

// freport reads the content of the eventually existing report file
func freport(c *InstallerContext) stepResults {
	sc := InitCodeStepResult("Reading the execution report", nil, noCleanUpRequired)

	ok := c.ef.Output.Contains(REPORT_OUTPUT_FILE)
	if ok {
		c.log.Println("A report file from a previous execution has been located")
		b, err := ioutil.ReadFile(util.JoinPaths(c.ef.Output.Path(), REPORT_OUTPUT_FILE))
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf(ERROR_READING_REPORT, REPORT_OUTPUT_FILE, err.Error()), nil)
			goto MoveOut
		}

		report := ReportFileContent{}

		err = json.Unmarshal(b, &report)
		if err != nil {
			FailsOnCode(&sc, err, fmt.Sprintf(ERROR_UNMARSHALLING_REPORT, REPORT_OUTPUT_FILE, err.Error()), nil)
			goto MoveOut
		}
		c.report = report
	} else {
		c.log.Println("Unable to locate a report file from a previous execution")
	}

MoveOut:
	return sc.Array()
}
