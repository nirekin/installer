package main

import (
	"log"

	"github.com/lagoon-platform/engine"
)

type installerContext struct {
	location      string
	client        string
	sshPublicKey  string
	sshPrivateKey string
	httpProxy     string
	httpsProxy    string
	noProxy       string
	log           *log.Logger
	lagoon        engine.Lagoon
	lagoonError   error
	ef            engine.ExchangeFolder
}
