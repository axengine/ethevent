package http

import (
	"github.com/axengine/ethevent/pkg/http/bean"
	"github.com/labstack/echo/v4"
	"net/http"
)

// eventList
// @Summary 查询事件
// @Description 查询事件
// @Tags Event
// @Accept json
// @Produce json
// @Param Request body bean.EventListRo true "request param"
// @Success 200 {array} bean.Event "success"
// @Router /v1/event/list [POST]
func (hs *HttpServer) eventList(c echo.Context) error {
	var req bean.EventListRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailMsg("invalid parameter"))
	}
	data, err := hs.svc.EventList(&req)
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(data))
}
