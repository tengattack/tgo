package log

import (
	"testing"
	
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLogLevel(t *testing.T) {
	assert := assert.New(t)
	l := logrus.New()

	err := SetLogLevel(l, "debug")
	assert.Nil(err)

	err = SetLogLevel(l, "invalid")
	assert.Equal("not a valid logrus Level: \"invalid\"", err.Error())
}

func TestSetLogOut(t *testing.T) {
	assert := assert.New(t)
	l := logrus.New()

	err := SetLogOut(l, "stdout")
	assert.Nil(err)

	err = SetLogOut(l, "stderr")
	assert.Nil(err)

	// missing create logs folder.
	err = SetLogOut(l, "logs/access.log")
	assert.NotNil(err)
}

func TestInitDefaultLog(t *testing.T) {
	assert := assert.New(t)
	conf := DefaultConfig

	// no errors on default config
	assert.Nil(InitLog(conf))

	conf.AccessLevel = "invalid"

	assert.NotNil(InitLog(conf))
}

func TestAccessLevel(t *testing.T) {
	assert := assert.New(t)
	conf := DefaultConfig

	conf.AccessLevel = "invalid"

	assert.NotNil(InitLog(conf))
}

func TestErrorLevel(t *testing.T) {
	assert := assert.New(t)
	conf := DefaultConfig

	conf.ErrorLevel = "invalid"

	assert.NotNil(InitLog(conf))
}

func TestAccessLogPath(t *testing.T) {
	assert := assert.New(t)
	conf := DefaultConfig

	conf.AccessLog = "logs/access.log"

	assert.NotNil(InitLog(conf))
}

func TestErrorLogPath(t *testing.T) {
	assert := assert.New(t)
	conf := DefaultConfig

	conf.ErrorLog = "logs/error.log"

	assert.NotNil(InitLog(conf))
}
