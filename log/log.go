package log

import "fmt"

var verbosityLevel = 0

func SetVerbosityLevel(level int) {
	verbosityLevel = level
}

func Log(msg string) {
	fmt.Println(msg)
}

func LogInfo(msg string) {
	if verbosityLevel >= 2 {
		Log("INFO: " + msg)
	}
}

func LogWarning(msg string) {
	if verbosityLevel >= 1 {
		Log("WARNING: " + msg)
	}
}

func LogError(msg string) {
	Log("ERROR: " + msg)
}

