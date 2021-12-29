package logs

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})
}
