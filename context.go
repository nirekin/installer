package installer

import (
	"log"

	"github.com/lagoon-platform/engine"
	"gopkg.in/yaml.v2"
)

type Buffer struct {
	Envvars   map[string]string
	Extravars map[string]string
	Param     map[string]interface{}
}

func (bu Buffer) Params() (b []byte, e error) {
	b, e = yaml.Marshal(bu.Param)
	return
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
	buffer        Buffer
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
