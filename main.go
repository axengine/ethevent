package main

import (
	"context"
	"github.com/axengine/ethevent/pkg/chainindex"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/http"
	"github.com/axengine/ethevent/pkg/svc"
	"github.com/axengine/utils/log"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title eth-events API
// @version 0.1.0
// @description
// @host
// @BasePath /
func main() {
	dbo := database.New("./test.db", log.Logger)
	ci := chainindex.New(log.Logger, dbo)
	if err := ci.Init(); err != nil {
		log.Logger.Panic("Init", zap.Error(err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := ci.Start(ctx); err != nil {
		log.Logger.Panic("Start", zap.Error(err))
	}

	httpServer := http.New(svc.New(log.Logger, dbo))
	go httpServer.Start(ctx, true, 8080)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-exit:
			cancel()
			if err := httpServer.Stop(ctx); err != nil {
				log.Logger.Warn("http.Stop", zap.Error(err))
			}
		case <-ctx.Done():
			log.Logger.Info("main exit")
			time.Sleep(time.Second * 3)
			os.Exit(0)
		}
	}
}
