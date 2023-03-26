package http

import (
	"context"
	"fmt"
	_ "github.com/axengine/ethevent/docs"
	"github.com/axengine/ethevent/pkg/http/validator"
	"github.com/axengine/ethevent/pkg/svc"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type HttpServer struct {
	svc *svc.Service
	ec  *echo.Echo
}

func New(svc *svc.Service) *HttpServer {
	return &HttpServer{svc: svc, ec: echo.New()}
}

func (hs *HttpServer) Start(ctx context.Context, dev bool, port int) {
	// set CORS
	e := hs.ec
	e.Validator = &validator.CustomValidator{}

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.DefaultCORSConfig))
	// set DOCS
	if dev {
		e.GET("/docs/*any", echoSwagger.WrapHandler)
	}

	v1 := e.Group("/v1")
	v1.GET("/task/list", hs.taskList)
	v1.POST("/task/add", hs.taskAdd)
	v1.POST("/event/list", hs.eventList)

	// Start the echo v4 server on port 8080
	fmt.Printf("Starting server on port %d\n", port)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func (hs *HttpServer) Stop(ctx context.Context) error {
	return hs.ec.Shutdown(ctx)
}
