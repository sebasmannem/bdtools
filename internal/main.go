package internal

import (
	"context"
	"github.com/sebasmannem/bdtools/pkg/quotagroups"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
	//ctx context.Context
)

func InitLogger(logger *zap.SugaredLogger) {
	log = logger
	quotagroups.InitLogger(log)
}

func InitContext(c context.Context) {
	quotagroups.InitContext(c)
}
