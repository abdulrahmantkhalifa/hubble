// Hubble uses a custom logger to not clutter the the global logger. If an
// application using Hubble as a library wants to disable or redirect Hubble
// logs, it can do so by changing the Hubble logger.
package logging

import (
	"log"
	"os"
)

// DefaultLogger is the logger used by all components of the Hubble application.
// You can replace or reconfigure at will, but don't set it to nil.
var DefaultLogger = log.New(os.Stdout, log.Prefix(), log.Flags())

func Println(v ...interface{}) {
	DefaultLogger.Println(v...)
}

func Print(v ...interface{}) {
	DefaultLogger.Print(v...)
}

func Printf(format string, v ...interface{}) {
	DefaultLogger.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	DefaultLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	DefaultLogger.Fatalf(format, v...)
}
