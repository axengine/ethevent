package main

import (
	"context"
	"github.com/axengine/ethevent/pkg/chainindex"
	"github.com/axengine/ethevent/pkg/dbo"
	"github.com/axengine/utils/log"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dbo := dbo.New("./test.db", log.Logger)
	ci := chainindex.New(log.Logger, dbo)
	if err := ci.Init(); err != nil {
		log.Logger.Panic("Init", zap.Error(err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := ci.Start(ctx); err != nil {
		log.Logger.Panic("Start", zap.Error(err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-exit:
		cancel()
	}
	log.Logger.Info("main exit")
}
