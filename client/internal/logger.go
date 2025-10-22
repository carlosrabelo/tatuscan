//go:build windows || linux || darwin

package internal

import "github.com/sirupsen/logrus"

// Logger variable that can be set from main package
var Log *logrus.Logger

// SetLogger sets the logger to be used by internal functions
func SetLogger(logger *logrus.Logger) {
	Log = logger
}
