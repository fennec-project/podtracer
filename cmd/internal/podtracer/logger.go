package podtracer

import (
	logger "log"
	"os"
)

func Log(msgLogLevel string, msg string, args ...interface{}) {
	systemLogLevel := os.Getenv("PODTRACER_LOGLEVEL")
	if systemLogLevel == "DEBUG" {
		logger.Printf("["+msgLogLevel+"] "+msg, args)
		return
	} else if msgLogLevel != "DEBUG" {
		logger.Printf("["+msgLogLevel+"] "+msg, args)
		return
	} else {
		return
	}
}
