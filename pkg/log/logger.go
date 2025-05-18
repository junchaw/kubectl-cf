package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var DefaultLogger = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.WarnLevel,
}

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.WarnLevel
	}
	DefaultLogger.SetLevel(level)
}
