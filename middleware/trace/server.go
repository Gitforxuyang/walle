package trace

import (
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/trace"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

func ServerTrace() gin.HandlerFunc {
	tracer := trace.GetTracer()
	conf := config.GetConfig().GetTraceConfig()
	return func(ctx *gin.Context) {
		span, err := tracer.StartServerSpanFromContext(ctx,
			fmt.Sprintf("%s_%s", ctx.Request.URL.Path, ctx.Request.Method))
		if err != nil {
			logger.GetLogger().Error(ctx, "链路错误", logger.Fields{"err": utils.StructToMap(err)})
		}
		ctx.Set("_span", span)
		defer span.Finish()
		ctx.Next()
		status, _ := ctx.Get("status")
		s := uint16(status.(int))
		ext.HTTPStatusCode.Set(span, s)
		if conf.Log {
			req, _ := ctx.Get("req")
			resp, _ := ctx.Get("resp")
			span.LogFields(
				log.Object("req", utils.StructToJson(req)),
				log.Object("resp", utils.StructToJson(resp)),
			)
		}
		e, _ := ctx.Get("err")
		if e != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"))
			span.LogFields(
				log.Object("evaError", utils.StructToJson(e)),
			)
		}
	}
}
