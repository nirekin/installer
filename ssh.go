package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/engine/ssh"
)

// fSHKeys checks if the SSH keys are specified via environment variables.
//
// If:
//		YES; they will be loaded into the context
//		NOT; they will be generated and then loaded into the context
//
func fSHKeys(c *InstallerContext) stepContexts {
	sc := InitStepContext("Generation the SSH keys", nil, noCleanUpRequired)
	var generate bool
	if c.ef.Input.Contains(engine.SSHPuplicKeyFileName) && c.ef.Input.Contains(engine.SSHPrivateKeyFileName) {
		c.sshPublicKey = filepath.Join(c.ef.Input.Path(), engine.SSHPuplicKeyFileName)
		c.sshPrivateKey = filepath.Join(c.ef.Input.Path(), engine.SSHPrivateKeyFileName)
		generate = false
		c.log.Println("SSHKeys not generation required")
	} else {
		c.log.Println("SSHKeys generation required")
		generate = true
	}

	if generate {
		publicKey, privateKey, e := ssh.Generate()
		if e != nil {
			sc.Err = fmt.Errorf(ERROR_GENERATING_SSH_KEYS, e.Error())
			sc.ErrDetail = "An error occured generating the keys :v"
			goto MoveOut
		}
		_, e = engine.SaveFile(c.log, *c.ef.Input, engine.SSHPuplicKeyFileName, publicKey)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the public key into :%v", c.ef.Input.Path())
			goto MoveOut
		}
		_, e = engine.SaveFile(c.log, *c.ef.Input, engine.SSHPrivateKeyFileName, privateKey)
		if e != nil {
			sc.Err = e
			sc.ErrDetail = fmt.Sprintf("An error occured saving the private key into :%v", c.ef.Input.Path())
			goto MoveOut
		}
		c.sshPublicKey = filepath.Join(c.ef.Input.Path(), engine.SSHPuplicKeyFileName)
		c.sshPrivateKey = filepath.Join(c.ef.Input.Path(), engine.SSHPrivateKeyFileName)

	MoveOut:
		// If the keys have been generated then they should be cleaned in case
		// of subsequent errors
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

		return sc.Array()
	}

	if c.log != nil {
		c.log.Printf(LOG_SSH_PUBLIC_KEY, c.sshPublicKey)
		c.log.Printf(LOG_SSH_PRIVATE_KEY, c.sshPrivateKey)
	}
	return sc.Array()
}
