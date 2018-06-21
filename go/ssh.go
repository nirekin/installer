package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/engine/ssh"
)

func fSHKeys(c *installerContext) (error, cleanup) {
	generate := false
	if c.ef.Input.Contains(engine.SSHPuplicKeyFileName) {
		c.sshPublicKey = filepath.Join(c.ef.Input.Path(), engine.SSHPuplicKeyFileName)
		if c.ef.Input.Contains(engine.SSHPrivateKeyFileName) {
			c.sshPrivateKey = filepath.Join(c.ef.Input.Path(), engine.SSHPrivateKeyFileName)
		} else {
			generate = true

		}
	} else {
		generate = true
	}

	if generate {
		publicKey, privateKey, e := ssh.Generate()
		if e != nil {
			return fmt.Errorf(ERROR_GENERATING_SSH_KEYS, e.Error()), nil
		}
		engine.SaveFile(c.log, c.ef.Input.Path(), engine.SSHPuplicKeyFileName, publicKey)
		engine.SaveFile(c.log, c.ef.Input.Path(), engine.SSHPrivateKeyFileName, privateKey)
		c.sshPublicKey = filepath.Join(c.ef.Input.Path(), "generated_", engine.SSHPuplicKeyFileName)
		c.sshPrivateKey = filepath.Join(c.ef.Input.Path(), "generated_", engine.SSHPrivateKeyFileName)
	}
	if c.log != nil {
		c.log.Printf(LOG_SSH_PUBLIC_KEY, c.sshPublicKey)
		c.log.Printf(LOG_SSH_PRIVATE_KEY, c.sshPrivateKey)
	}
	return nil, func(c *installerContext) func(c *installerContext) error {
		return func(c *installerContext) (err error) {
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
}
