package main

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("")
var format = logging.MustStringFormatter(
	"%{time} %{level:.4s} %{shortpkg}:%{shortfunc}: %{message}",
)

// Password is just an example type implementing the Redactor interface. Any
// time this is logged, the Redacted() function will be called.
type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

/*func foo() {
	log.Debug("format string %s", "aa")
}*/

func SetupLogging() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	backendFormatter :=
		logging.NewBackendFormatter(backend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.INFO, "")

	logging.SetBackend(backendLeveled)
	/*
		log.Debug("debug %s", Password("secret"))
		log.Info("this is an info")
		log.Notice("notice")
		log.Warning("warning")
		log.Error("err")
		log.Critical("crit")*/
}
