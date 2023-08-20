package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Fields type, used to pass to `WithFields`.
type Fields logrus.Fields

// Entry is the final or intermediate logger logging entry.
type Entry struct {
	// Contains all the fields set by the user.
	Data Fields
}

var entryPool sync.Pool

// WithField add a single field to the Entry.
func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(Fields{key: value})
}

// WithFields add a map of fields to the Entry.
func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	for k, v := range entry.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{Data: data}
}

func newEntry() *Entry {
	entry, ok := entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return &Entry{
		// Default is five fields, give a little extra room
		Data: make(Fields, 5),
	}
}

func releaseEntry(entry *Entry) {
	entry.Data = map[string]interface{}{}
	entryPool.Put(entry)
}

// Debug debug
func (entry *Entry) Debug(args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Debug(args...)
}

// Debugf debug with format
func (entry *Entry) Debugf(format string, args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Debugf(format, args...)
}

// Info info
func (entry *Entry) Info(args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Info(args...)
}

// Infof info with format
func (entry *Entry) Infof(format string, args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Infof(format, args...)
}

// Warn warn
func (entry *Entry) Warn(args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Warn(args...)
}

// Warnf warn with format
func (entry *Entry) Warnf(format string, args ...interface{}) {
	logrusEntry := LogAccess.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Warnf(format, args...)
}

// Error error
func (entry *Entry) Error(args ...interface{}) {
	logrusEntry := LogError.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Error(args...)
}

// Errorf error with format
func (entry *Entry) Errorf(format string, args ...interface{}) {
	logrusEntry := LogError.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Errorf(format, args...)
}

// Fatal fatal
func (entry *Entry) Fatal(args ...interface{}) {
	logrusEntry := LogError.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Fatal(args...)
}

// Fatalf fatal with formatter
func (entry *Entry) Fatalf(format string, args ...interface{}) {
	logrusEntry := LogError.WithFields(logrus.Fields(entry.Data))
	SetCallFrame(logrusEntry, CallerSkip)
	logrusEntry.Fatalf(format, args...)
}

// WithField adds a field to the log entry, note that it doesn't log until you
// call Debug, Info, Warn, Error, Fatal or Panic. It only creates a log entry.
// If you want multiple fields, use `WithFields`.
func WithField(key string, value interface{}) *Entry {
	entry := newEntry()
	defer releaseEntry(entry)
	return entry.WithField(key, value)
}

// WithFields adds a struct of fields to the log entry. All it does is call
// `WithField` for each `Field`.
func WithFields(fields Fields) *Entry {
	entry := newEntry()
	defer releaseEntry(entry)
	return entry.WithFields(fields)
}
