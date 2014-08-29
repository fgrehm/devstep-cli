package devstep

import (
	logPkg "github.com/segmentio/go-log"
	"os"
)

var log *logPkg.Logger
var LogLevel string

func init() {
	log = logPkg.New(os.Stderr, logPkg.NOTICE, "")
}

func SetLogLevel(level string) {
	switch (level) {
		case "d", "debug", "DEBUG":
			log.Level = logPkg.DEBUG
			LogLevel = "DEBUG"
		case "i", "info", "INFO":
			log.Level = logPkg.INFO
			LogLevel = "INFO"
	}
}
