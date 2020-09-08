package server

import (
	"fmt"
	"github.com/Gitforxuyang/walle/middleware/catch"
	"github.com/gin-gonic/gin"
)

func InitServer() {
	r := gin.New()
	//内部转发
	r.Group("/rpc",
		func(ctx *gin.Context) {
		}).
		Use(catch.RpcServerCatch()).
		POST("/:Service/:Method", func(ctx *gin.Context) {
			fmt.Println(ctx.Params.Get("Service"))
			fmt.Println(ctx.Params.Get("Method"))
			m := ctx.GetStringMap("req")
			fmt.Println(m)
			ctx.JSON(200, map[string]string{"code": "123"})
		})

	r.Use(catch.ServerCatch())
	r.Run(":8080")
}
