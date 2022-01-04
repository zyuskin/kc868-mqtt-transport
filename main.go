package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	config := NewConfig()
	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logrus.Warnf("Wrong log level %s, set default log level %s", config.LogLevel, ConfigDefaultLogLevel)
		logLevel, _ = logrus.ParseLevel(ConfigDefaultLogLevel)
	}
	logrus.SetLevel(logLevel)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	manager := NewManger(config)
	manager.Start()
}
