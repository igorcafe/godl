package log

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "", 0)
var (
	Debugf = logger.Printf
	Infof  = logger.Printf
	Warnf  = logger.Printf
	Errorf = logger.Printf
	Fatalf = logger.Panicf
)
