package log_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tengattack/tgo/log"
)

func TestSetLogLevel(t *testing.T) {
	assert := assert.New(t)
	l := logrus.New()

	err := log.SetLogLevel(l, "debug")
	assert.Nil(err)

	err = log.SetLogLevel(l, "invalid")
	assert.Equal("not a valid logrus Level: \"invalid\"", err.Error())
}

func TestSetLogOut(t *testing.T) {
	assert := assert.New(t)
	l := logrus.New()

	err := log.SetLogOut(l, "stdout")
	assert.Nil(err)

	err = log.SetLogOut(l, "stderr")
	assert.Nil(err)

	// missing create logs folder.
	err = log.SetLogOut(l, "logs/access.log")
	assert.NotNil(err)
}

func TestInitDefaultLog(t *testing.T) {
	assert := assert.New(t)
	conf := log.DefaultConfig

	// no errors on default config
	assert.Nil(log.InitLog(conf))

	conf.AccessLevel = "invalid"

	assert.NotNil(log.InitLog(conf))
}

func TestAccessLevel(t *testing.T) {
	assert := assert.New(t)
	conf := log.DefaultConfig

	conf.AccessLevel = "invalid"

	assert.NotNil(log.InitLog(conf))
}

func TestErrorLevel(t *testing.T) {
	assert := assert.New(t)
	conf := log.DefaultConfig

	conf.ErrorLevel = "invalid"

	assert.NotNil(log.InitLog(conf))
}

func TestAccessLogPath(t *testing.T) {
	assert := assert.New(t)
	conf := log.DefaultConfig

	conf.AccessLog = "logs/access.log"

	assert.NotNil(log.InitLog(conf))
}

func TestErrorLogPath(t *testing.T) {
	assert := assert.New(t)
	conf := log.DefaultConfig

	conf.ErrorLog = "logs/error.log"

	assert.NotNil(log.InitLog(conf))
}
