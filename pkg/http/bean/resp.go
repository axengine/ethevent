package bean

import (
	"github.com/axengine/ethevent/pkg/errorx"
	"github.com/axengine/utils/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

// Resp http resp
type Resp struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	TraceId string      `json:"traceId,omitempty"`
}

// ResultPage result with page info
type ResultPage struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// Success display successful signal
func (r *Resp) Success(result interface{}) *Resp {
	r.Code = 0
	r.Msg = "ok"
	r.Data = result
	return r
}

// SuccessPage display successful result with page info
func (r *Resp) SuccessPage(content interface{}, total int64) *Resp {
	r.Code = 0
	r.Msg = "ok"
	r.Data = ResultPage{content, total}
	return r
}

// Fail display fail signal
func (r *Resp) Fail(code int, desc string, result interface{}) *Resp {
	r.Code = code
	r.Msg = desc
	r.Data = result
	return r
}

// FailMsg display fail msg
func (r *Resp) FailMsg(desc string) *Resp {
	return r.Fail(errorx.CodeFailed, desc, nil)
}

func (r *Resp) FailErr(c echo.Context, err error) *Resp {
	langParam := c.FormValue("lang")
	accept := c.Request().Header.Get("Accept-Language")
	tags := parseTags(langParam, accept)

	var lang = "en"
	if len(tags) > 0 {
		lang = tags[0].String()
	}

	code := errorx.CodeSuccess
	msg := errorx.GetMessage(code, lang)

	if err != nil {
		r.TraceId = uuid.New().String()
		log.Logger.Warn("FailErr", zap.String("traceId", r.TraceId), zap.Error(err))
		if e, ok := err.(errorx.Error); ok {
			code = e.Code
			msg = errorx.GetMessage(code, lang)
		} else {
			code = errorx.CodeSystem
			msg = errorx.GetMessage(code, lang)
		}
	}

	r.Code = code
	r.Msg = msg
	return r
}

func parseTags(langs ...string) []language.Tag {
	tags := []language.Tag{}
	for _, lang := range langs {
		t, _, err := language.ParseAcceptLanguage(lang)
		if err != nil {
			continue
		}
		tags = append(tags, t...)
	}
	return tags
}
