// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"log"
	"log/syslog"
	"time"

	"git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
)

// LogEntry is an entry in the townsourced log
type LogEntry struct {
	Time    time.Time
	Fields  map[string]interface{}
	Level   string
	Message string
}

// LogHook is a hook for logging townsourced errors in the database
type LogHook struct {
}

// Levels implements the logrus.Hook interface for LogHook
func (l *LogHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// Fire implements the logrus.Hook interface for LogHook
func (l *LogHook) Fire(entry *logrus.Entry) error {
	go func(e *logrus.Entry) {
		lEntry := &LogEntry{
			Time:    e.Time,
			Fields:  map[string]interface{}(e.Data),
			Level:   e.Level.String(),
			Message: e.Message,
		}
		err := data.Log(lEntry)
		if err != nil {
			//error writing log to database, write to syslog instead
			slog, serr := syslog.NewLogger(lEntry.syslogPriority(), log.LstdFlags)
			if serr != nil {
				return
			}
			slog.Printf("Error logging entry in database: %s, Original Entry: %s", err, lEntry.Message)
		}

	}(entry)

	return nil
}

func (le *LogEntry) syslogPriority() syslog.Priority {
	switch le.Level {
	case logrus.PanicLevel.String():
		return syslog.LOG_EMERG
	case logrus.FatalLevel.String():
		return syslog.LOG_CRIT
	case logrus.ErrorLevel.String():
		return syslog.LOG_ERR
	case logrus.WarnLevel.String():
		return syslog.LOG_WARNING
	case logrus.InfoLevel.String():
		return syslog.LOG_INFO
	case logrus.DebugLevel.String():
		return syslog.LOG_DEBUG
	default:
		return syslog.LOG_INFO
	}
}
