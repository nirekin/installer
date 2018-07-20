package installer

import (
	"log"

	"github.com/lagoon-platform/engine"
)

type InstallerContext struct {
	// The environment descriptor location
	location string
	// The environment descriptor name
	name string
	// The client requestion the CRUD operation on the environment
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
	session       *engine.EngineSession
	buffer        map[string]engine.Buffer
}

func CreateContext(l *log.Logger) *InstallerContext {
	c := &InstallerContext{}
	c.buffer = make(map[string]engine.Buffer)
	c.log = l
	return c
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

func (c *InstallerContext) getBuffer(p *engine.FolderPath) engine.Buffer {
	// We check if we have a buffer corresponding to the provided folder path
	if val, ok := c.buffer[p.Path()]; ok {
		return val
	}
	return engine.CreateBuffer()
}
