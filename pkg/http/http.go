package http

import (
	"context"
	"fmt"
	_ "github.com/axengine/ethevent/docs"
	"github.com/axengine/ethevent/pkg/http/validator"
	"github.com/axengine/ethevent/pkg/svc"
	echopprof "github.com/hiko1129/echo-pprof"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"sync"
)

type HttpServer struct {
	svc *svc.Service
	ec  *echo.Echo
}

func New(svc *svc.Service) *HttpServer {
	return &HttpServer{svc: svc, ec: echo.New()}
}

func (hs *HttpServer) Start(ctx context.Context, wg *sync.WaitGroup, dev bool, port int) {
	defer wg.Done()
	// set CORS
	e := hs.ec
	e.Validator = &validator.CustomValidator{}

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.DefaultCORSConfig))

	if dev {
		// set DOCS
		e.GET("/docs/*any", echoSwagger.WrapHandler)
		// pprof
		echopprof.Wrap(e)
	}

	v1 := e.Group("/v1")
	v1.GET("/task/list", hs.taskList)
	v1.POST("/task/add", hs.taskAdd)
	v1.POST("/task/pause", hs.taskPause)
	v1.POST("/task/delete", hs.taskDelete)
	v1.POST("/task/update", hs.taskUpdate)

	v1.POST("/event/list", hs.eventList)

	// Start the echo v4 server on port 8080
	fmt.Printf("Starting server on port %d\n", port)

	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		e.Logger.Error(err)
	}
}

func (hs *HttpServer) Stop(ctx context.Context) error {
	return hs.ec.Shutdown(ctx)
}
