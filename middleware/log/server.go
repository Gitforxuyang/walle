package log

import (
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/gin-gonic/gin"
	"time"
)

func ServerLogger() gin.HandlerFunc {
	log := logger.GetLogger()
	config := config.GetConfig()
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		if config.GetLogConfig().Server {
			req, _ := ctx.Get("req")
			resp, _ := ctx.Get("resp")
			err, _ := ctx.Get("err")
			emap := map[string]interface{}{}
			if err != nil {
				emap, _ = utils.JsonToMap(utils.StructToJson(err))
			}
			log.Info(ctx, "收到的请求", logger.Fields{
				"req":     req,
				"resp":    resp,
				"method":  ctx.Request.Method,
				"url":     ctx.Request.URL.Path,
				"useTime": fmt.Sprintf("%s", time.Now().Sub(start).String()),
				"err":     emap,
			})
		}
	}
}
