package main

import (
	"context"
	"github.com/sebasmannem/bdtools/internal"
)

var (
	ctx context.Context
	//ctxCancelFunc context.CancelFunc
	config internal.Config
)

func initContext() {
	ctx, _ = config.GetTimeoutContext(context.Background())
	internal.InitContext(ctx)
}

func main() {
	var err error
	if config, err = internal.NewConfig(); err != nil {
		initLogger("")
		log.Fatal(err)
	} else {
		initLogger(config.LogFile)
		initRemoteLoggers()
		enableDebug(config.Debug)
		log.Debug("initializing config object")
		defer log.Sync() //nolint:errcheck
		log.Debug("checking if patching is required")
		initContext()
		if err = config.QuotaGroups.PatchProjectQuotaGroups(); err != nil {
			log.Fatal(err)
		}
	}
}
