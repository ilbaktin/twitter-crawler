package log

import "fmt"

type Logger struct {
	Name   string
	prefix string
}

func NewLogger(name string) *Logger {
	prefix := fmt.Sprintf("%s: ", name)
	return &Logger{
		name,
		prefix,
	}
}

func (l *Logger) LogInfo(msg string, args ...interface{}) {
	LogInfo(l.prefix+msg, args...)
}

func (l *Logger) LogWarning(msg string, args ...interface{}) {
	LogWarning(l.prefix+msg, args...)
}

func (l *Logger) LogError(msg string, args ...interface{}) {
	LogError(l.prefix+msg, args...)
}
