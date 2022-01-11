package sam

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util/logs"
)

var sam *faas.Client

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.Background()); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
}

func NewEvent(ctx context.Context, name string) {

}

func NewRequest(ctx context.Context, name string) {

}
