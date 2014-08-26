package devstep

import (
	"os"
	logPkg "github.com/segmentio/go-log"
)

var log *logPkg.Logger

func init() {
	log = logPkg.New(os.Stderr, logPkg.NOTICE, "")
}

func Verbose(verbose bool) {
	if verbose {
		log.Level = logPkg.DEBUG
	}
}
