package server

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/middleware/catch"
	"github.com/Gitforxuyang/walle/middleware/log"
	"github.com/Gitforxuyang/walle/middleware/trace"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//注册关闭服务时的回调
type RegisterShutdown func()

var (
	shutdownFunc []RegisterShutdown = make([]RegisterShutdown, 0)
)

func InitServer() {
	conf := config.GetConfig()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	//内部转发
	r.Group("/rpc",
		func(ctx *gin.Context) {
		}).
		Use(trace.ServerTrace()).
		Use(log.ServerLogger()).
		Use(catch.RpcServerCatch()).
		POST("/:Service/:Method", func(ctx *gin.Context) {
			//m := ctx.GetStringMap("req")
			ctx.Set("resp", map[string]string{"code": "123"})
		})

	r.
		Use(trace.ServerTrace()).
		Use(log.ServerLogger()).
		Use(catch.ServerCatch())
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.GetPort()),
		Handler: r,
	}
	go func() {
		err := srv.ListenAndServe()
		time.Sleep(time.Millisecond * 500)
		utils.Must(err)
	}()
	time.Sleep(time.Millisecond * 200)
	logger.GetLogger().Info(context.TODO(), "server started", logger.Fields{
		"port":   config.GetConfig().GetPort(),
		"server": config.GetConfig().GetName(),
		"env":    config.GetConfig().GetENV(),
	})
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	logger.GetLogger().Info(context.TODO(), "signal", logger.Fields{
		"signal": s.String(),
	})
	srv.Shutdown(context.TODO())
	//做一些资源关闭动作
	for _, v := range shutdownFunc {
		v()
	}
	logger.GetLogger().Info(context.TODO(), "server stop", logger.Fields{})
}
