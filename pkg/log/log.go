package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger(logLevel string, logFile string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Println("Can't parse log level:", err)
		os.Exit(1)
	}
	Logger = &logrus.Logger{
		Level: level,
		Formatter: &logrus.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}
	Logger.SetLevel(level)
	Logger.Out = os.Stdout
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			Logger.Out = file
		} else {
			Logger.Errorf("Unable log to file: %s, using default stdout", logFile)
		}
	}
	Logger.Debug("log level set to: ", logLevel)
}
