package log

import (
	"errors"
	"net"
	"net/url"
	"os"
	
	"github.com/sirupsen/logrus"
	logrusagent "github.com/tengattack/logrus-agent-hook"
	"golang.org/x/term"
)

// Config is logging config.
type Config struct {
	Format      string      `yaml:"format"`
	AccessLog   string      `yaml:"access_log"`
	AccessLevel string      `yaml:"access_level"`
	ErrorLog    string      `yaml:"error_log"`
	ErrorLevel  string      `yaml:"error_level"`
	Agent       AgentConfig `yaml:"agent"`
}

// AgentConfig is sub section of LogConfig.
type AgentConfig struct {
	Enabled    bool   `yaml:"enabled"`
	DSN        string `yaml:"dsn"`
	AppID      string `yaml:"app_id"`
	Host       string `yaml:"host"`
	InstanceID string `yaml:"instance_id"`
	Category   string `yaml:"category"`
}

var (
	// IsTerm instructs current stdout whether is terminal
	IsTerm bool
	// LogAccess is log access log
	LogAccess *logrus.Logger
	// LogError is log error log
	LogError *logrus.Logger
	// conf package config
	conf *Config
)

// DefaultConfig is default configuration
var DefaultConfig = &Config{
	Format:      "string",
	AccessLog:   "stdout",
	AccessLevel: "info",
	ErrorLog:    "stderr",
	ErrorLevel:  "error",
	Agent: AgentConfig{
		Enabled: false,
	},
}

func init() {
	IsTerm = term.IsTerminal(int(os.Stdout.Fd()))
}

// GetLogConfig return current log config
func GetLogConfig() *Config {
	return conf
}

// InitLog use for initial log module
func InitLog(logConf *Config) error {
	var err error

	if logConf != nil {
		conf = logConf
	} else {
		conf = DefaultConfig
	}
	// get default host and instance id from environment variables or hostname
	if conf.Agent.Host == "" || conf.Agent.InstanceID == "" {
		hostname, _ := os.Hostname()
		if conf.Agent.Host == "" {
			host := os.Getenv("HOST")
			if host == "" {
				host = hostname
			}
			conf.Agent.Host = host
		}
		if conf.Agent.InstanceID == "" {
			instanceID := os.Getenv("INSTANCE_ID")
			if instanceID == "" {
				instanceID = hostname
			}
			conf.Agent.InstanceID = instanceID
		}
	}

	// init logger
	LogAccess = logrus.New()
	LogError = logrus.New()

	LogAccess.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006/01/02 - 15:04:05",
		FullTimestamp:   true,
	}

	LogError.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006/01/02 - 15:04:05",
		FullTimestamp:   true,
	}

	// set logger
	if err = SetLogLevel(LogAccess, conf.AccessLevel); err != nil {
		return errors.New("Set access log level error: " + err.Error())
	}

	if err = SetLogLevel(LogError, conf.ErrorLevel); err != nil {
		return errors.New("Set error log level error: " + err.Error())
	}

	if err = SetLogOut(LogAccess, conf.AccessLog); err != nil {
		return errors.New("Set access log path error: " + err.Error())
	}

	if err = SetLogOut(LogError, conf.ErrorLog); err != nil {
		return errors.New("Set error log path error: " + err.Error())
	}

	return nil
}

// SetLogOut provide log stdout and stderr output
func SetLogOut(log *logrus.Logger, outString string) error {
	switch outString {
	case "stdout":
		log.Out = os.Stdout
	case "stderr":
		log.Out = os.Stderr
	default:
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			return err
		}

		log.Out = f
	}

	if conf != nil && conf.Agent.Enabled {
		// configure log agent (logstash) hook
		u, err := url.Parse(conf.Agent.DSN)
		if err != nil {
			return err
		}
		conn, err := net.Dial(u.Scheme, u.Host)
		if err != nil {
			return err
		}
		fields := logrus.Fields{
			"app_id":      conf.Agent.AppID,
			"host":        conf.Agent.Host,
			"instance_id": conf.Agent.InstanceID,
		}
		if conf.Agent.Category != "" {
			fields["category"] = conf.Agent.Category
		}
		hook := logrusagent.New(
			conn, logrusagent.DefaultFormatter(fields))
		log.Hooks.Add(hook)
	}

	return nil
}

// SetLogLevel is define log level what you want
// log level: panic, fatal, error, warn, info and debug
func SetLogLevel(log *logrus.Logger, levelString string) error {
	level, err := logrus.ParseLevel(levelString)

	if err != nil {
		return err
	}

	log.Level = level

	return nil
}
