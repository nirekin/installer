package installer

import (
	"log"

	"github.com/lagoon-platform/engine"
)

type buffer struct {
	envvars   map[string]string
	extravars map[string]string
}

type InstallerContext struct {
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
	ef            *engine.ExchangeFolder
	buffer        buffer
}

func (c *InstallerContext) SetLog(l *log.Logger) {
	c.log = l
}

func (c *InstallerContext) LogPrintln(v ...interface{}) {
	c.log.Println(v)
}

func (c *InstallerContext) LogFatal(v ...interface{}) {
	c.log.Fatal(v)
}
