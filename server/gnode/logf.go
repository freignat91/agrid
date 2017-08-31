package gnode

import (
	"log"
	"strings"
)

const (
	LOG_ERROR = 0
	LOG_WARN  = 1
	LOG_INFO  = 2
	LOG_DEBUG = 3
)

type Logf struct {
	level int
}

var logf = Logf{level: 2}

func (l *Logf) setLevel(level string) {
	if strings.ToLower(level) == "error" {
		l.level = LOG_ERROR
	} else if strings.ToLower(level) == "warn" {
		l.level = LOG_WARN
	} else if strings.ToLower(level) == "info" {
		l.level = LOG_INFO
	} else if strings.ToLower(level) == "debug" {
		l.level = LOG_DEBUG
	}
}

func (l *Logf) levelString() string {
	switch l.level {
	case LOG_ERROR:
		return "error"
	case LOG_WARN:
		return "warn"
	case LOG_INFO:
		return "info"
	case LOG_DEBUG:
		return "debug"
	default:
		return "?"
	}
}

func (l *Logf) error(format string, args ...interface{}) {
	if l.level >= LOG_ERROR {
		log.Printf(format, args...)
	}
}

func (l *Logf) warn(format string, args ...interface{}) {
	if l.level >= LOG_WARN {
		log.Printf(format, args...)
	}
}

func (l *Logf) info(format string, args ...interface{}) {
	if l.level >= LOG_INFO {
		log.Printf(format, args...)
	}
}

func (l *Logf) debug(format string, args ...interface{}) {
	if l.level >= LOG_DEBUG {
		log.Printf(format, args...)
	}
}

func (l *Logf) debugMes(mes *AntMes, format string, args ...interface{}) {
	if mes != nil && mes.Debug {
		log.Printf(format, args...)
	}
}

func (l *Logf) printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
