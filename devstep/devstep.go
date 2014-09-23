package devstep

import (
	logPkg "github.com/segmentio/go-log"
	"os"
	"strings"
)

var log *logPkg.Logger
var LogLevel string

func init() {
	log = logPkg.New(os.Stderr, logPkg.NOTICE, "")
}

func SetLogLevel(level string) error {
	if err := log.SetLevelString(level); err != nil {
		return err
	} else {
		LogLevel = strings.ToUpper(level)
		return nil
	}
}
