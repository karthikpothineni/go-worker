package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Log - log object
var Log *logrus.Logger

// Init - initializes the logger
func Init() {

	Log = logrus.New()

	Log.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	}

	Log.Out = os.Stdout
}
