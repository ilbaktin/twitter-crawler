package log

import (
	"fmt"
	"time"
)

var verbosityLevel = 0

func SetVerbosityLevel(level int) {
	verbosityLevel = level
}

func Log(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), msg)
	fmt.Printf(msg, args...)
}

func LogInfo(msg string, args ...interface{}) {
	if verbosityLevel >= 2 {
		Log("INFO: "+msg, args...)
	}
}

func LogWarning(msg string, args ...interface{}) {
	if verbosityLevel >= 1 {
		Log("WARNING: "+msg, args...)
	}
}

func LogError(msg string, args ...interface{}) {
	Log("ERROR: "+msg, args...)
}
