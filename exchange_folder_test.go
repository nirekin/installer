package installer

import (
	"log"
	"os"
	"testing"

	"github.com/lagoon-platform/engine"
	"github.com/stretchr/testify/assert"
)

func TestChildExchangeFolderOk(t *testing.T) {

	ef, e := engine.CreateExchangeFolder("./", "testFolder")
	assert.Nil(t, e)
	assert.NotNil(t, ef)
	defer ef.Delete()

	e = ef.Create()
	assert.Nil(t, e)

	log := log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	sc := InitStepContext("DummyStep", nil, noCleanUpRequired)

	subEf, ko := createChildExchangeFolder(ef.Input, "subTestFolder", &sc, log)
	assert.False(t, ko)
	assert.NotNil(t, subEf)
	assert.Equal(t, sc.ErrorDetail, "")
	assert.Nil(t, sc.Error)
	assert.Nil(t, sc.ErrorOrigin)

	_, err := os.Stat(subEf.Location.Path())
	assert.Nil(t, err)
}

func TestChildExchangeFolderKo(t *testing.T) {

	ef, e := engine.CreateExchangeFolder("./", "testFolder")
	assert.Nil(t, e)
	assert.NotNil(t, ef)
	defer ef.Delete()

	// We are not calling ef.Create() in order to get an error creating the child

	log := log.New(os.Stdout, "Test", log.Ldate|log.Ltime|log.Lmicroseconds)
	sc := InitStepContext("DummyStep", nil, noCleanUpRequired)

	subEf, ko := createChildExchangeFolder(ef.Input, "subTestFolfer", &sc, log)
	assert.True(t, ko)
	assert.NotNil(t, subEf)
	assert.Equal(t, sc.ErrorDetail, "")
	assert.Equal(t, sc.ErrorOrigin, originLagoonInstaller)
	assert.NotNil(t, sc.Error)

}
