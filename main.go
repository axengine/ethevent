package main

import (
	"context"
	"github.com/axengine/ethevent/pkg/chainindex"
	"github.com/axengine/ethevent/pkg/database"
	"github.com/axengine/ethevent/pkg/http"
	"github.com/axengine/ethevent/pkg/svc"
	"github.com/axengine/utils/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

// @title eth-events API
// @version 0.1.0
// @description
// @host
// @BasePath /
func main() {
	dbo := database.New(filepath.Join(viper.GetString("datadir"), "events.db"), log.Logger)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	ci := chainindex.New(log.Logger, dbo)
	if err := ci.Init(); err != nil {
		log.Logger.Panic("Init", zap.Error(err))
	}
	wg.Add(1)
	go ci.Start(ctx, &wg)

	httpServer := http.New(svc.New(log.Logger, dbo))
	wg.Add(1)
	go httpServer.Start(ctx, &wg, true, viper.GetInt("http.port"))

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
			wg.Wait()
			log.Logger.Info("main exit")
			os.Exit(0)
		}
	}
}

func init() {
	pflag.Int("http.port", 8080, "server http port")
	pflag.String("datadir", ".", "db path")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}
