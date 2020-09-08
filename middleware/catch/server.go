package catch

import (
	"fmt"
	"github.com/Gitforxuyang/walle/util"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func ServerCatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Println(ctx.Request.URL.Path)
		q := ctx.Request.URL.Query()
		bytes, err := ioutil.ReadAll(ctx.Request.Body)
		util.Must(err)
		m, err := util.JsonToMap(string(bytes))
		util.Must(err)
		for k, v := range q {
			if len(v) > 1 {
				m[k] = v
			} else {
				m[k] = v[0]
			}

		}
		ctx.Set("req", m)
	}
}

func RpcServerCatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bytes, err := ioutil.ReadAll(ctx.Request.Body)
		util.Must(err)
		m, err := util.JsonToMap(string(bytes))
		util.Must(err)
		ctx.Set("req", m)
	}
}
