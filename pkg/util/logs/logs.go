package logs

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
		ForceColors:   false,
	})
}
