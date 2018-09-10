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
func fSHKeys(c *InstallerContext) (error, cleanup) {
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
			return fmt.Errorf(ERROR_GENERATING_SSH_KEYS, e.Error()), nil
		}
		engine.SaveFile(c.log, *c.ef.Input, engine.SSHPuplicKeyFileName, publicKey)
		engine.SaveFile(c.log, *c.ef.Input, engine.SSHPrivateKeyFileName, privateKey)
		c.sshPublicKey = filepath.Join(c.ef.Input.Path(), engine.SSHPuplicKeyFileName)
		c.sshPrivateKey = filepath.Join(c.ef.Input.Path(), engine.SSHPrivateKeyFileName)
	}

	if c.log != nil {
		c.log.Printf(LOG_SSH_PUBLIC_KEY, c.sshPublicKey)
		c.log.Printf(LOG_SSH_PRIVATE_KEY, c.sshPrivateKey)
	}

	if generate {
		// If the keys have been generated the they should be cleaned in case
		// of subsequent errors
		return nil, func(c *InstallerContext) func(c *InstallerContext) error {
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
	} else {
		return nil, noCleanUpRequired
	}
}
