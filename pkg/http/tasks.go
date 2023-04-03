package http

import (
	"context"
	"github.com/axengine/ethevent/pkg/errorx"
	"github.com/axengine/ethevent/pkg/http/bean"
	"github.com/labstack/echo/v4"
	"net/http"
)

// taskList
// @Summary 查询任务列表
// @Description 查询任务列表
// @Tags TASK
// @Accept json
// @Produce json
// @Param Request query bean.PageRo true "request param"
// @Success 200 {array} model.Task "success"
// @Router /v1/task/list [GET]
func (hs *HttpServer) taskList(c echo.Context) error {
	var req bean.PageRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailMsg("invalid parameter"))
	}
	data, err := hs.svc.TaskList(req.Cursor, req.Limit, "ASC")
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(data))
}

// taskAdd
// @Summary 添加任务
// @Description 添加任务
// @Tags TASK
// @Accept json
// @Produce json
// @Param Request query bean.TaskAddRo true "request param"
// @Success 200 {object} bean.Resp "success"
// @Router /v1/task/add [POST]
func (hs *HttpServer) taskAdd(c echo.Context) error {
	var req bean.TaskAddRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, errorx.ErrParamInvalid.MultiErr(err)))
	}
	err := hs.svc.TaskAdd(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(nil))
}

// taskPause
// @Summary 任务暂停与运行
// @Description 任务暂停与运行
// @Tags TASK
// @Accept json
// @Produce json
// @Param Request query bean.TaskPauseRo true "request param"
// @Success 200 {object} bean.Resp "success"
// @Router /v1/task/pause [POST]
func (hs *HttpServer) taskPause(c echo.Context) error {
	var req bean.TaskPauseRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, errorx.ErrParamInvalid.MultiErr(err)))
	}
	err := hs.svc.TaskPause(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(nil))
}

// taskDelete
// @Summary 删除任务
// @Description 删除任务
// @Tags TASK
// @Accept json
// @Produce json
// @Param Request query bean.TaskDeleteRo true "request param"
// @Success 200 {object} bean.Resp "success"
// @Router /v1/task/delete [POST]
func (hs *HttpServer) taskDelete(c echo.Context) error {
	var req bean.TaskDeleteRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, errorx.ErrParamInvalid.MultiErr(err)))
	}
	err := hs.svc.TaskDelete(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(nil))
}

// taskUpdate
// @Summary 更新任务
// @Description 更新任务
// @Tags TASK
// @Accept json
// @Produce json
// @Param Request query bean.TaskUpdateRo true "request param"
// @Success 200 {object} bean.Resp "success"
// @Router /v1/task/update [POST]
func (hs *HttpServer) taskUpdate(c echo.Context) error {
	var req bean.TaskUpdateRo
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, errorx.ErrParamInvalid.MultiErr(err)))
	}
	err := hs.svc.TaskUpdate(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, new(bean.Resp).FailErr(c, err))
	}
	return c.JSON(http.StatusOK, new(bean.Resp).Success(nil))
}
