package installer

import (
	"log"

	"github.com/lagoon-platform/engine"
	"github.com/lagoon-platform/engine/ansible"
	"github.com/lagoon-platform/engine/util"
)

type InstallerContext struct {
	// The environment descriptor location
	location string
	// The environment descriptor name
	name          string
	sshPublicKey  string
	sshPrivateKey string
	httpProxy     string
	httpsProxy    string
	noProxy       string
	log           *log.Logger
	lagoon        engine.Engine
	lagoonError   error
	ef            *util.ExchangeFolder
	session       *engine.EngineSession
	buffer        map[string]ansible.Buffer
	cliparams     ansible.ParamContent
}

func CreateContext(l *log.Logger) *InstallerContext {
	c := &InstallerContext{}
	c.buffer = make(map[string]ansible.Buffer)
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

func (c *InstallerContext) getBuffer(p *util.FolderPath) ansible.Buffer {
	// We check if we have a buffer corresponding to the provided folder path
	if val, ok := c.buffer[p.Path()]; ok {
		return val
	}
	return ansible.CreateBuffer()
}

func (c *InstallerContext) BuildBaseParam(nodeSetId string, provider string) ansible.BaseParam {
	return ansible.BuildBaseParam(c.lagoon.Environment().QualifiedName(), nodeSetId, provider, c.sshPublicKey, c.sshPrivateKey)
}
