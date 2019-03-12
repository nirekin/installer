package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/ssh"
	"github.com/ekara-platform/engine/util"
)

type NodeExtraVars struct {
	Params    map[string]string
	Instances int
}

func Run(c InstallerContext) (e error) {
	// Check if the received action is supporter by the engine
	c.Log().Println(LOG_RUNNING)
	a := os.Getenv(util.ActionEnvVariableKey)

	am := engine.CreateActionManager()
	switch a {
	case engine.ActionCreateId.String():
		c.Log().Println(LOG_ACTION_CREATE)
		// We repeat the invocation of the fillContext in each case in order
		// to be able to throw a specific error to detect an unsupported action
		e = fillContext(&c)
		if e != nil {
			return e
		}

		if e := fillSHKeys(&c); e != nil {
			return e
		}
		am.Run(engine.ActionCreateId, c)
	case engine.ActionInstallId.String():
		c.Log().Println(LOG_ACTION_INSTALL)
		e = fillContext(&c)
		if e != nil {
			return e
		}

		if e := fillSHKeys(&c); e != nil {
			return e
		}
		am.Run(engine.ActionInstallId, c)
	case engine.ActionDeployId.String():
		c.Log().Println(LOG_ACTION_DEPLOY)
		e = fillContext(&c)
		if e != nil {
			return e
		}

		if e := fillSHKeys(&c); e != nil {
			return e
		}
		am.Run(engine.ActionDeployId, c)
	case engine.ActionCheckId.String():
		c.Log().Println(LOG_ACTION_CHECK)
		e = fillContext(&c)
		if e != nil {
			return e
		}

		am.Run(engine.ActionCheckId, c)
	case engine.ActionDumpId.String():
		c.Log().Println(LOG_ACTION_DUMP)
		e = fillContext(&c)
		if e != nil {
			return e
		}

		am.Run(engine.ActionDumpId, c)
	default:
		if a == "" {
			a = LOG_NO_ACTION
		}
		// Bad luck; unsupported action!
		e = fmt.Errorf(ERROR_UNSUPORTED_ACTION, a)
	}
	return
}

func fillContext(c *InstallerContext) error {
	c.Log().Println("Filling the installer context")
	fillProxy(c)
	if e := fillQualifiedName(c); e != nil {
		return e
	}
	if e := fillExchangeFolder(c); e != nil {
		return e
	}
	if e := fillLocation(c); e != nil {
		return e
	}
	if e := fillCliparam(c); e != nil {
		return e
	}
	fillEkara(c)
	if c.ekaraError != nil {
		c.Log().Printf("Error filling the installer context:  %s", c.ekaraError.Error())
		return c.ekaraError
	}
	return nil
}

// fillProxy loads the proxy settings form the environmant variables into the
// context
func fillProxy(c *InstallerContext) {
	// We check if the proxy is well defined, the proxy can be required in order
	// to be capable to download the environment descriptor content and all its
	// related components
	c.httpProxyContent, c.httpsProxyContent, c.noProxyContent = engine.CheckProxy()
}

// fillQualifiedName extracts the qualified environment name from the
// environment variable "engine.StarterEnvQualifiedVariableKey"
func fillQualifiedName(c *InstallerContext) error {
	c.qualifiedNameContent = os.Getenv(util.StarterEnvQualifiedVariableKey)
	if c.qualifiedNameContent == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvQualifiedVariableKey)
	}
	return nil
}

func fillExchangeFolder(c *InstallerContext) error {
	var err error
	c.efolder, err = util.CreateExchangeFolder(util.InstallerVolume, "")
	if err != nil {
		return fmt.Errorf(engine.ERROR_CREATING_EXCHANGE_FOLDER, c.qualifiedNameContent, err.Error())
	}
	return nil
}

// fillLocation extracts the descriptor location and descriptor  file name from the
// environment variables "engine.StarterEnvVariableKey" and
// engine.StarterEnvNameVariableKey
func fillLocation(c *InstallerContext) error {
	c.locationContent = os.Getenv(util.StarterEnvVariableKey)
	if c.locationContent == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvVariableKey)
	}
	c.name = os.Getenv(util.StarterEnvNameVariableKey)
	if c.name == "" {
		return fmt.Errorf(ERROR_REQUIRED_ENV, util.StarterEnvNameVariableKey)
	}
	return nil
}

func fillCliparam(c *InstallerContext) error {
	ok := c.Ef().Location.Contains(util.CliParametersFileName)
	if ok {
		p, e := ansible.ParseParams(util.JoinPaths(c.Ef().Location.Path(), util.CliParametersFileName))
		if e != nil {
			return fmt.Errorf(ERROR_LOADING_CLI_PARAMETERS, e)
		}
		c.cliparams = p
		c.Log().Printf(LOG_CLI_PARAMS, c.cliparams)
	}
	return nil
}

func fillEkara(c *InstallerContext) {
	root, flavor := repositoryFlavor(c.Location())
	if c.Cliparams() != nil {
		c.Log().Printf("Creating ekara environment with parameter for templating")
		c.engine, c.ekaraError = engine.Create(c.Log(), "/var/lib/ekara", c.Cliparams())
	} else {
		c.Log().Printf("Creating ekara environment without parameter for templating")
		c.engine, c.ekaraError = engine.Create(c.Log(), "/var/lib/ekara", map[string]interface{}{})
	}

	if c.ekaraError == nil {
		c.ekaraError = c.engine.Init(root, flavor, c.name)
	}
	// Note: here we are not taking in account the "c.ekaraError != nil" to place the error
	// into the context because this error managment depends on the whole process
	// and will be treated if required by the "ffailOnEkaraError" step
}

// fSHKeys checks if the SSH keys are specified via environment variables.
//
// If:
//		YES; they will be loaded into the context
//		NOT; they will be generated and then loaded into the context
//
func fillSHKeys(c *InstallerContext) error {
	var generate bool
	if c.Ef().Input.Contains(util.SSHPuplicKeyFileName) && c.Ef().Input.Contains(util.SSHPrivateKeyFileName) {
		c.sshPublicKeyContent = filepath.Join(c.Ef().Input.Path(), util.SSHPuplicKeyFileName)
		c.sshPrivateKeyContent = filepath.Join(c.Ef().Input.Path(), util.SSHPrivateKeyFileName)
		generate = false
		c.Log().Println("SSHKeys not generation required")
	} else {
		c.Log().Println("SSHKeys generation required")
		generate = true
	}

	if generate {
		publicKey, privateKey, e := ssh.Generate()
		if e != nil {
			return fmt.Errorf(ERROR_GENERATING_SSH_KEYS, e.Error())
		}
		_, e = util.SaveFile(c.Log(), *c.Ef().Input, util.SSHPuplicKeyFileName, publicKey)
		if e != nil {
			return fmt.Errorf("An error occured saving the public key into :%v", c.Ef().Input.Path())
		}
		_, e = util.SaveFile(c.Log(), *c.Ef().Input, util.SSHPrivateKeyFileName, privateKey)
		if e != nil {
			return fmt.Errorf("An error occured saving the private key into :%v", c.Ef().Input.Path())
		}
		c.sshPublicKeyContent = filepath.Join(c.Ef().Input.Path(), util.SSHPuplicKeyFileName)
		c.sshPrivateKeyContent = filepath.Join(c.Ef().Input.Path(), util.SSHPrivateKeyFileName)

		// If the keys have been generated then they should be cleaned in case
		// of subsequent errors
		/*
			sc.CleanUp = func(c *InstallerContext) func(c *InstallerContext) error {
				return func(c *InstallerContext) (err error) {
					if c.log != nil {
						c.log.Println("Running fSHKeys cleanup")
						c.log.Printf("Cleaning %s", c.sshPublicKey)
					}

					err = os.Remove(c.sshPublicKey)
					if err != nil {
						return
					}
					if c.log != nil {
						c.log.Printf("Cleaning %s", c.sshPrivateKey)
					}

					err = os.Remove(c.sshPrivateKey)
					if err != nil {
						return
					}
					return
				}
			}(c)
		*/
	}

	if c.Log() != nil {
		c.Log().Printf(LOG_SSH_PUBLIC_KEY, c.sshPublicKeyContent)
		c.Log().Printf(LOG_SSH_PRIVATE_KEY, c.sshPrivateKeyContent)
	}
	return nil
}

// runCreate launches the environment creation
/*
func runCreate(c engine.LaunchContext) engine.ExecutionReport {
	// Stack of functions required to create an environment
	calls := []step{
	fproxy,
	fqualifiedName,
	fexchangeFoldef,
	flocation,
	fcliparam,
	fekara,
	fSHKeys,
		//freport,
		//ffailOnEkaraError,
		//fsetup,
		//fconsumesetup,
		//fcreate,
		//fconsumecreate,
		//fsetuporchestrator,
		//fconsumesetuporchestrator,
		//forchestrator,
		//fstack,
	}
	return launch(calls, c)
}

// runCheck launches the environment check
func runCheck(c engine.LaunchContext) engine.ExecutionReport {
	// Stack of functions required to check an environment
	calls := []step{
	fproxy,
	fqualifiedName,
	fexchangeFoldef,
	flocation,
	fcliparam,
	fekara,

		//flogCheck,
	}
	return launch(calls, c)
}
*/

//repositoryFlavor returns the repository flavor, branchn tag ..., based on the
// presence of '@' into the given url
func repositoryFlavor(url string) (string, string) {

	if strings.Contains(url, "@") {
		s := strings.Split(url, "@")
		return s[0], s[1]
	}
	return url, ""
}
