package installer

import (
	"fmt"
	"log"

	"github.com/ekara-platform/engine/util"
)

func createChildExchangeFolder(parent *util.FolderPath, name string, sc *stepContext, log *log.Logger) (*util.ExchangeFolder, bool) {
	ef, e := parent.AddChildExchangeFolder(name)
	if e != nil {
		err := fmt.Errorf(ERROR_ADDING_EXCHANGE_FOLDER, name, e.Error())
		log.Printf(err.Error())
		InstallerFail(sc, err, "")
		return ef, true
	}
	e = ef.Create()
	if e != nil {
		err := fmt.Errorf(ERROR_CREATING_EXCHANGE_FOLDER, name, e.Error())
		log.Printf(err.Error())
		InstallerFail(sc, err, "")
		return ef, true
	}
	return ef, false
}
