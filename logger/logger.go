package logger

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	logrusagent "github.com/tengattack/logrus-agent-hook"
	"github.com/tengattack/tgo/log"
)

type contextKey int

const (
	keyCaller contextKey = iota
)

var (
	// LogAccess is log server request log
	LogAccess *logrus.Logger
	// LogError is log server error log
	LogError *logrus.Logger

	// CallerSkip .
	CallerSkip = 1
)

var currentProjectName string

// InitLog inits the logger in this package
func InitLog(projectName string, logConf *log.Config) error {
	err := log.InitLog(logConf)
	if err != nil {
		return err
	}
	currentProjectName = projectName
	LogAccess = log.LogAccess
	LogError = log.LogError

	conf := log.GetLogConfig()
	logFileFormatter := NewLogFileFormatter(conf.Agent.AppID)
	if logConf.AccessLog != "" {
		LogAccess.SetFormatter(logFileFormatter)
	}
	if logConf.ErrorLog != "" {
		LogError.SetFormatter(logFileFormatter)
	}

	// configure logstash
	LogAccess.ReplaceHooks(make(logrus.LevelHooks))
	LogError.ReplaceHooks(make(logrus.LevelHooks))
	if conf != nil && conf.Agent.Enabled {
		_, err := url.Parse(conf.Agent.DSN)
		if err != nil {
			return fmt.Errorf("parse dsn error: %v", err)
		}

		var opt logrusagent.Options
		opt.ChannelSize = conf.Agent.ChannelSize

		fields := logrus.Fields{
			"app_id":      conf.Agent.AppID,
			"host":        conf.Agent.Host,
			"instance_id": conf.Agent.InstanceID,
		}
		if conf.Agent.Category != "" {
			fields["category"] = conf.Agent.Category
		}

		agentFormatter := NewLogstashFormatter(fields)
		hook, _ := logrusagent.New(conf.Agent.DSN, agentFormatter, opt)
		LogAccess.Hooks.Add(hook)
		LogError.Hooks.Add(hook)
	}
	return nil
}

// getRelativePath return the relative path of file in current project
func getRelativePath(filePath string) string {
	items := strings.SplitN(filePath, "/"+currentProjectName+"/", 2)
	return items[len(items)-1]
}

// SetCallFrame .
func SetCallFrame(entry *logrus.Entry, skip int) {
	_, file, line, _ := runtime.Caller(skip + 1)
	entry.Context = context.WithValue(context.Background(), keyCaller, &runtime.Frame{
		File: getRelativePath(file),
		Line: line,
	})
}

// Debug log as debug level
func Debug(args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Debug(args...)
}

// Debugf log as debug level with format
func Debugf(format string, args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Debugf(format, args...)
}

// Info log as info level
func Info(args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Info(args...)
}

// Infof log as info level with format
func Infof(format string, args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Infof(format, args...)
}

// Warn log as warn level
func Warn(args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Warn(args...)
}

// Warnf log as warn level with format
func Warnf(format string, args ...interface{}) {
	entry := logrus.NewEntry(LogAccess)
	SetCallFrame(entry, CallerSkip)
	entry.Warnf(format, args...)
}

// Error log as error level
func Error(args ...interface{}) {
	entry := logrus.NewEntry(LogError)
	SetCallFrame(entry, CallerSkip)
	entry.Error(args...)
}

// Errorf log as error level with format
func Errorf(format string, args ...interface{}) {
	entry := logrus.NewEntry(LogError)
	SetCallFrame(entry, CallerSkip)
	entry.Errorf(format, args...)
}

// Fatal log as fatal level and exit
func Fatal(args ...interface{}) {
	entry := logrus.NewEntry(LogError)
	SetCallFrame(entry, CallerSkip)
	entry.Fatal(args...)
}

// Fatalf log as fatal level with format and exit
func Fatalf(format string, args ...interface{}) {
	entry := logrus.NewEntry(LogError)
	SetCallFrame(entry, CallerSkip)
	entry.Fatalf(format, args...)
}
