package devstep

import (
	logPkg "github.com/segmentio/go-log"
	"os"
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
