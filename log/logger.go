package log

import "fmt"

type Logger struct {
	Name	string
	prefix	string
}

func NewLogger(name string) *Logger {
	prefix := fmt.Sprintf("%s: ", name)
	return &Logger{
		name,
		prefix,
	}
}

func (l *Logger) LogInfo(msg string) {
	LogInfo(l.prefix + msg)
}

func (l *Logger) LogWarning(msg string) {
	LogWarning(l.prefix + msg)
}

func (l *Logger) LogError(msg string) {
	LogError(l.prefix + msg)
}
