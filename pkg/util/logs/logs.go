package logs

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var isInit = false

func Init() {

	if isInit {
		return
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
		ForceColors:   false,
	})

	isInit = true
}
