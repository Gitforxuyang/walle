package catch

import (
	"github.com/Gitforxuyang/walle/server/vo"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/logger"
	util "github.com/Gitforxuyang/walle/util/utils"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/opentracing/opentracing-go"
	log2 "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"io/ioutil"
)

func ServerCatch() gin.HandlerFunc {
	log := logger.GetLogger()
	return func(ctx *gin.Context) {
		defer catch(ctx, log)
		bytes, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Set("err", error2.UnknowError.SetDetail(err.Error()))
		}
		m, err := util.JsonToMap(string(bytes))
		if err != nil {
			ctx.Set("err", error2.UnknowError.SetDetail(err.Error()))
		}
		q := ctx.Request.URL.Query()
		for k, v := range q {
			if len(v) > 1 {
				m[k] = v
			} else {
				m[k] = v[0]
			}

		}
		ctx.Set("req", m)
		ctx.Next()
	}
}

func RpcServerCatch() gin.HandlerFunc {
	log := logger.GetLogger()
	return func(ctx *gin.Context) {
		defer catch(ctx, log)
		bytes, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Set("err", error2.UnknowError.SetDetail(err.Error()))
		}
		m, err := util.JsonToMap(string(bytes))
		if err != nil {
			ctx.Set("err", error2.UnknowError.SetDetail(err.Error()))
		}
		ctx.Set("req", m)
		ctx.Next()
	}
}

func catch(ctx *gin.Context, log logger.EvaLogger) {
	if e := recover(); e != nil {
		ctx.Set("err", error2.PanicError)
		log.Error(ctx, "发生panic", logger.Fields{"e": e})
		span, ok := ctx.Value("_span").(opentracing.Span)
		if ok {
			span.LogFields(log2.Object("stack", zap.Stack("stack")))
		}
		sentry.CaptureException(errors.New(e))
		ctx.Set("status", 500)
		resp := vo.Resp{Code: error2.PanicError.Code, Message: error2.PanicError.Message}
		ctx.Set("resp", resp)
		ctx.JSON(500, resp)
	} else {
		e, exists := ctx.Get("err")
		if exists {
			ee := error2.FromError(e.(error))
			ctx.JSON(500, vo.Resp{Code: ee.Code, Message: ee.Message})
		} else {
			resp, exists := ctx.Get("resp")
			if !exists {
				panic("不存在resp")
			}
			ctx.Set("status", 200)
			resp = vo.Resp{Code: 0, Message: "ok", Data: resp}
			ctx.Set("resp", resp)
			ctx.JSON(200, resp)
		}
	}
}
