package log

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/utils"
	"google.golang.org/grpc"
	"time"
)

func NewClientWrapper() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log := logger.GetLogger()
	conf := config.GetConfig()
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		var err error
		defer func() {
			emap := map[string]interface{}{}
			if err != nil {
				emap, _ = utils.JsonToMap(utils.StructToJson(error2.DecodeStatus(err)))
			}
			r, _ := utils.JsonToMap(utils.StructToJson(req))
			res, _ := utils.JsonToMap(utils.StructToJson(reply))
			//errobject, _ := utils.JsonToMap(utils.StructToJson(emap))
			if conf.GetLogConfig().GRpcClient {
				log.Info(ctx, "发起的请求", logger.Fields{
					"req":     r,
					"resp":    res,
					"method":  method,
					"useTime": fmt.Sprintf("%s", time.Now().Sub(start).String()),
					"err":     emap,
				})
			}
		}()
		err = invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}
