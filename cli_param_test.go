package installer

import (
	"log"
	"os"
	"testing"

	"github.com/ekara-platform/engine/util"
	"github.com/stretchr/testify/assert"
)

func TestReadingParam(t *testing.T) {

	ef, e := util.CreateExchangeFolder("./", "testFolfer")
	assert.Nil(t, e)
	assert.NotNil(t, ef)
	defer ef.Delete()

	e = ef.Create()
	assert.Nil(t, e)

	pContent := `key1: value1`

	e = ef.Location.Write([]byte(pContent), util.CliParametersFileName)
	assert.Nil(t, e)

	c := &InstallerContext{
		ef:  ef,
		log: log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds),
	}

	sc := fcliparam(c)
	assert.Nil(t, sc.Results[0].error)
	cParam := c.cliparams
	assert.NotNil(t, cParam)
	assert.Equal(t, len(cParam), 1)

}
