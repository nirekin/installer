package installer

import (
	"log"

	"github.com/ekara-platform/engine"
	"github.com/ekara-platform/engine/ansible"
	"github.com/ekara-platform/engine/util"
)

type (
	InstallerContext struct {
		// The environment descriptor name
		name                 string
		logger               *log.Logger
		efolder              *util.ExchangeFolder
		locationContent      string
		qualifiedNameContent string
		httpProxyContent     string
		httpsProxyContent    string
		noProxyContent       string
		sshPublicKeyContent  string
		sshPrivateKeyContent string
		cliparams            ansible.ParamContent
		engine               engine.Engine
		ekaraError           error
	}
)

func (c InstallerContext) Name() string {
	return c.name
}

func (c InstallerContext) Log() *log.Logger {
	return c.logger
}

func (c InstallerContext) Ef() *util.ExchangeFolder {
	return c.efolder
}

func (c InstallerContext) Ekara() engine.Engine {
	return c.engine
}

func (c InstallerContext) QualifiedName() string {
	return c.qualifiedNameContent
}

func (c InstallerContext) Location() string {
	return c.locationContent
}

func (c InstallerContext) HttpProxy() string {
	return c.httpProxyContent
}

func (c InstallerContext) HttpsProxy() string {
	return c.httpsProxyContent
}

func (c InstallerContext) NoProxy() string {
	return c.noProxyContent
}

func (c InstallerContext) SshPublicKey() string {
	return c.sshPublicKeyContent
}

func (c InstallerContext) SshPrivateKey() string {
	return c.sshPrivateKeyContent
}

func (c InstallerContext) Cliparams() ansible.ParamContent {
	return c.cliparams
}

func (c InstallerContext) Error() error {
	return c.ekaraError
}

func CreateContext(l *log.Logger) *InstallerContext {
	c := &InstallerContext{}
	c.logger = l
	return c
}
