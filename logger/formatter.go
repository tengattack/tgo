package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	logrusagent "github.com/tengattack/logrus-agent-hook"
)

// LogFileFormatter defines the format for log file
type LogFileFormatter struct {
	logrus.TextFormatter

	MinimumCallerDepth int
}

// LogstashFormatter defines the format for Logstash
type LogstashFormatter struct {
	logrusagent.LogAgentFormatter

	FieldKeyTime       string
	FieldKeyMsg        string
	FieldKeyLevel      string
	FieldKeyCategory   string
	TimestampFormat    string
	MinimumCallerDepth int
	DisableSorting     bool
}

var (
	// Using a pool to re-use of old entries when formatting Logstash messages.
	// It is used in the Fire function.
	logstashEntryPool = sync.Pool{
		New: func() interface{} {
			return &logrus.Entry{}
		},
	}
	logstashFields = logrus.Fields{"@version": "1"}
)

// NewLogFileFormatter return the log format for log file
// eg: 2019-01-31T04:48:20 [info] [controllers/aibf/character.go:99] foo key=value
func NewLogFileFormatter(projectName string) *LogFileFormatter {
	currentProjectName = projectName
	return &LogFileFormatter{
		TextFormatter: logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05",
			FullTimestamp:   true,
		},
		MinimumCallerDepth: 0,
	}
}

// Format renders a single log entry for log file
// the original file log format is defined here: github.com/sirupsen/logrus/text_formatter.TextFormatter{}.Format()
func (f *LogFileFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(Fields)
	for k, v := range entry.Data {
		data[k] = v
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		if nil != f.SortingFunc {
			f.SortingFunc(keys)
		} else {
			sort.Strings(keys)
		}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	b.WriteString(fmt.Sprintf("%s [%s]", entry.Time.Format(timestampFormat), entry.Level.String()))

	if entry.Context != nil {
		caller, _ := entry.Context.Value(keyCaller).(*runtime.Frame)
		if caller != nil {
			b.WriteString(fmt.Sprintf(" [%s:%d]", caller.File, caller.Line))
		}
	}

	if "" != entry.Message {
		b.WriteString(" " + entry.Message)
	}
	for _, key := range keys {
		value := data[key]
		appendKeyValue(b, key, value, f.QuoteEmptyFields)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// NewLogstashFormatter return the log format for Logstash
//
//	eg: {"@timestamp":"2019-01-31T04:48:20.259Z","@version":"1",\
//	  "app_id":"missevan-go","host":"DESKTOP-Q2ANV74","instance_id":"DESKTOP-Q2ANV74",\
//	  "level":"INFO","message":"[controllers/aibf/character.go:99] foo key=value"}
func NewLogstashFormatter(fields logrus.Fields) *LogstashFormatter {
	for k, v := range logstashFields {
		if _, ok := fields[k]; !ok {
			fields[k] = v
		}
	}

	return &LogstashFormatter{
		LogAgentFormatter: logrusagent.LogAgentFormatter{
			Fields: fields,
		},
		FieldKeyTime:       "@timestamp",
		FieldKeyMsg:        "message",
		FieldKeyLevel:      "level",
		FieldKeyCategory:   "category",
		MinimumCallerDepth: 0,
		TimestampFormat:    "2006-01-02T15:04:05.000Z",
	}
}

// Format renders a single log entry for Logstash
// the original logstash log format is defined here: github.com/tengattack/logrus-agent-hook/hook.LogAgentFormatter{}.Format()
func (f *LogstashFormatter) Format(e *logrus.Entry) ([]byte, error) {
	entry := copyEntry(e, f.Fields)
	defer logstashEntryPool.Put(entry)
	data := make(logrus.Fields, len(entry.Data)+4)
	extras := make(logrus.Fields)
	for k, v := range entry.Data {
		if _, ok := f.Fields[k]; ok || k == f.FieldKeyCategory {
			switch v := v.(type) {
			case error:
				data[k] = v.Error()
			default:
				data[k] = v
			}
		} else {
			switch v := v.(type) {
			case error:
				extras[k] = v.Error()
			default:
				extras[k] = v
			}
		}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}
	data[f.FieldKeyTime] = entry.Time.UTC().Format(timestampFormat)
	data[f.FieldKeyLevel] = getLevelString(entry.Level)

	var message string
	if entry.Context != nil {
		caller, _ := entry.Context.Value(keyCaller).(*runtime.Frame)
		if caller != nil {
			message = fmt.Sprintf("[%s:%d]", caller.File, caller.Line)
		}
	}
	if "" != entry.Message {
		message += " " + entry.Message
	}
	if len(extras) > 0 {
		b := &bytes.Buffer{}
		if !f.DisableSorting {
			extraKeys := make([]string, 0, len(extras))
			for k := range extras {
				extraKeys = append(extraKeys, k)
			}
			sort.Strings(extraKeys)
			for _, k := range extraKeys {
				appendKeyValue(b, k, extras[k], f.QuoteEmptyFields)
			}
		} else {
			for k, v := range extras {
				appendKeyValue(b, k, v, f.QuoteEmptyFields)
			}
		}
		message += " " + b.String()
	}
	data[f.FieldKeyMsg] = message

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	dataBytes := append(serialized, '\n')
	return dataBytes, nil
}

// appendKeyValue append value with key to data that to be appended to log file
func appendKeyValue(b *bytes.Buffer, key string, value interface{}, QuoteEmptyFields bool) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	appendValue(b, value, QuoteEmptyFields)
}

// appendValue append value to data used for method appendKeyValue
func appendValue(b *bytes.Buffer, value interface{}, QuoteEmptyFields bool) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !needsQuoting(stringVal, QuoteEmptyFields) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

// needsQuoting check where text needs to be quoted
func needsQuoting(text string, QuoteEmptyFields bool) bool {
	if QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

// Convert the Level to a string. E.g. ErrorLevel becomes "ERROR".
func getLevelString(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "DEBUG"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARN"
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.FatalLevel:
		return "FATAL"
	case logrus.PanicLevel:
		return "PANIC"
	}

	return "UNKNOWN"
}

// copyEntry copies the entry `e` to a new entry and then adds all the fields in `fields`\
//   that are missing in the new entry data.
// It uses `logstashEntryPool` to re-use allocated entries.
func copyEntry(e *logrus.Entry, fields logrus.Fields) *logrus.Entry {
	ne := logstashEntryPool.Get().(*logrus.Entry)
	ne.Context = e.Context
	ne.Message = e.Message
	ne.Level = e.Level
	ne.Time = e.Time
	ne.Data = logrus.Fields{}
	for k, v := range fields {
		ne.Data[k] = v
	}
	for k, v := range e.Data {
		ne.Data[k] = v
	}
	return ne
}
