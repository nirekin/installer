package installer

import (
	"fmt"
	"os"
	"strings"

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
		flocation,
		flagoon,
		ffailOnLagoonError,
		fdownloadcore,
		fenrichExchangeFolder,
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
	engine.SaveFile(c.log, *c.ef.Location, engine.CreationSessionFileName, by)
	return nil, nil
}

func fsetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Running setup for provider %s", p.Name)

		proEf := c.ef.Input.Children[p.Name]
		proEfIn := proEf.Input
		proEfOut := proEf.Output

		bp := engine.BuilBaseParam(c.client, "", p.Name, c.sshPublicKey, c.sshPrivateKey)
		b, e := bp.Content()
		if e != nil {
			return e, nil
		}

		e = proEfIn.Write(b, engine.ParamYamlFileName)
		if e != nil {
			return e, nil
		}

		// This is the first "real" step of the process so the used buffer is empty
		emptyBuff := engine.CreateBuffer()

		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *proEfIn)

		exv := engine.BuildExtraVars("", *proEfIn, *proEfOut, emptyBuff)

		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)

		engine.LaunchPlayBook(c.lagoon.ComponentManager(), p.Component, "setup.yml", exv, env, *c.log)
	}
	return nil, nil
}

func fconsumesetup(c *InstallerContext) (error, cleanup) {
	for _, p := range c.lagoon.Environment().Providers {
		c.log.Printf("Consume setup for provider %s", p.Name)
		proEfOut := c.ef.Input.Children[p.Name].Output
		err, buffer := engine.GetBuffer(proEfOut, c.log, "provider:"+p.Name)
		if err != nil {
			return err, nil
		}
		// Keep a reference on the buffer based on the output folder
		c.buffer[proEfOut.Path()] = buffer
	}
	return nil, nil
}

func fcreate(c *InstallerContext) (error, cleanup) {
	var proEf *engine.ExchangeFolder
	var nodeEf *engine.ExchangeFolder
	var e error

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		// unique id of the nodeset
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider

		// Provider exchange folder
		proEf = c.ef.Input.Children[p.ProviderName()]

		// We check if we have a buffer corresponding to the provider output
		buffer := c.getBuffer(proEf.Output)

		// Node exchange folder
		nodeEf, e = proEf.Input.AddChildExchangeFolder(n.Name)
		if e != nil {
			return e, nil
		}
		e = nodeEf.Create()
		if e != nil {
			return e, nil
		}

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
		np := n.NodeParams()
		bp.AddInt("instances", np.Instances)
		bp.AddNamedMap("params", np.Params)
		bp.AddInterface("volumes", n.Provider.Volumes())
		bp.AddBuffer(buffer) // We consume the potentials params comming from the buffer

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer) // We consume the potentials environment variables comming from the buffer

		// Prepare extra vars
		ev := engine.BuildExtraVars("", *nodeEf.Input, *nodeEf.Output, buffer)

		// We launch the playbook
		engine.LaunchPlayBook(c.lagoon.ComponentManager(), p.Component(), "create.yml", ev, env, *c.log)
	}

	return nil, nil
}

func fconsumecreate(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume create for node %s", n.Name)
		p := n.Provider
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]
		doutFolder := nodeEf.Output

		err, buffer := engine.GetBuffer(doutFolder, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			return err, nil
		}
		c.buffer[doutFolder.Path()] = buffer
	}
	return nil, nil
}

func fsetuporchestrator(c *InstallerContext) (error, cleanup) {

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		// unique id of the nodeset
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider

		// Provider exchange folder
		proEf := c.ef.Input.Children[p.ProviderName()]
		// Node exchange folder
		nodeEf := proEf.Input.Children[n.Name]

		// We check if we have a buffer corresponding to the node output
		buffer := c.getBuffer(nodeEf.Output)

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
		op := n.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddBuffer(buffer) // We consume the potentials params comming from the buffer

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer) // We consume the potentials environment variables comming from the buffer

		// Prepare extra vars
		exv := engine.BuildExtraVars("", *nodeEf.Input, *nodeEf.Output, buffer)

		// We launch the playbook
		engine.LaunchPlayBook(c.lagoon.ComponentManager(), c.lagoon.Environment().Orchestrator.Component, "setup.yml", exv, env, *c.log)
	}

	return nil, nil
}

func fconsumesetuporchestrator(c *InstallerContext) (error, cleanup) {
	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf("Consume create for node %s", n.Name)
		p := n.Provider
		proEf := c.ef.Input.Children[p.ProviderName()]
		nodeEf := proEf.Input.Children[n.Name]
		doutFolder := nodeEf.Output
		err, buffer := engine.GetBuffer(doutFolder, c.log, "node:"+n.Name)
		// Keep a reference on the buffer based on the output folder
		if err != nil {
			return err, nil
		}
		c.buffer[doutFolder.Path()] = buffer
	}
	return nil, nil
}

func forchestrator(c *InstallerContext) (error, cleanup) {

	for _, n := range c.lagoon.Environment().NodeSets {
		c.log.Printf(LOG_PROCESSING_NODE, n.Name)
		uid := c.session.CreationSession.Uids[n.Name]

		p := n.Provider

		// Provider exchange folder
		proEf := c.ef.Input.Children[p.ProviderName()]
		// Node exchange folder
		nodeEf := proEf.Input.Children[n.Name]

		// We check if we have a buffer corresponding to the node output
		buffer := c.getBuffer(nodeEf.Output)

		// Prepare parameters
		bp := engine.BuilBaseParam(c.client, uid, p.ProviderName(), c.sshPublicKey, c.sshPrivateKey)
		op := n.OrchestratorParams()
		bp.AddNamedMap("orchestrator", op)
		bp.AddBuffer(buffer) // We consume the potentials parameters comming from the buffer

		b, e := bp.Content()
		if e != nil {
			return e, nil
		}
		engine.SaveFile(c.log, *nodeEf.Input, engine.ParamYamlFileName, b)

		// Prepare components map
		e = c.lagoon.ComponentManager().SaveComponentsPaths(c.log, c.lagoon.Environment(), *nodeEf.Input)
		if e != nil {
			return e, noCleanUpRequired
		}

		// Prepare environment variables
		env := engine.BuildEnvVars()
		env.Add("http_proxy", c.httpProxy)
		env.Add("https_proxy", c.httpsProxy)
		env.AddBuffer(buffer) // We consume the potentials environment variables comming from the buffer

		// Prepare extra vars
		exv := engine.BuildExtraVars("", *nodeEf.Input, *nodeEf.Output, buffer)

		// We launch the playbook
		engine.LaunchPlayBook(c.lagoon.ComponentManager(), c.lagoon.Environment().Orchestrator.Component, "install.yml", exv, env, *c.log)
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
	root, flavor := repositoryFlavor(c.location)
	c.lagoon, c.lagoonError = engine.Create(c.log, "/var/lib/lagoon", root, flavor, c.name)
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
		c.log.Println(c.lagoonError)
		return fmt.Errorf(ERROR_PARSING_DESCRIPTOR, c.lagoonError.Error()), noCleanUpRequired
	}

	return nil, noCleanUpRequired
}
