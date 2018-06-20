package main

import (
	"fmt"
	"path/filepath"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/engine/ssh"
)

func fSHKeys() (error, cleanup) {
	generate := false
	if ef.Input.Contains(engine.SSHPuplicKeyFileName) {
		sshPublicKey = filepath.Join(ef.Input.Path(), engine.SSHPuplicKeyFileName)
		if ef.Input.Contains(engine.SSHPrivateKeyFileName) {
			sshPrivateKey = filepath.Join(ef.Input.Path(), engine.SSHPrivateKeyFileName)
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
		engine.SaveFile(loggerLog, ef.Input.Path(), engine.SSHPuplicKeyFileName, publicKey)
		engine.SaveFile(loggerLog, ef.Input.Path(), engine.SSHPrivateKeyFileName, privateKey)
		sshPublicKey = filepath.Join(ef.Input.Path(), "generated_", engine.SSHPuplicKeyFileName)
		sshPrivateKey = filepath.Join(ef.Input.Path(), "generated_", engine.SSHPrivateKeyFileName)
	}

	loggerLog.Printf(LOG_SSH_PUBLIC_KEY, sshPublicKey)
	loggerLog.Printf(LOG_SSH_PRIVATE_KEY, sshPrivateKey)
	return nil, nil
}
