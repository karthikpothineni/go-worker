package workerpool

import (
	"github.com/sirupsen/logrus"
)

var (
	log    = logrus.New().WithFields(logrus.Fields{"testworker": 1})
	worker = &Worker{Log: log}
)
