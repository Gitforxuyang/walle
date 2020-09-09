package trace

import (
	"context"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/trace"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
)

func NewClientWrapper() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	conf := config.GetConfig().GetTraceConfig()
	tracer := trace.GetTracer()
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if conf.GRpcClient {
			ctx, span, err := tracer.StartGRpcClientSpanFromContext(ctx, method)
			if err != nil {
				logger.GetLogger().Error(ctx, "链路错误", logger.Fields{"err": utils.StructToMap(err)})
			}
			defer span.Finish()
			err = invoker(ctx, method, req, reply, cc, opts...)
			if conf.Log {
				span.LogFields(
					log.Object("req", utils.StructToJson(req)),
					log.Object("resp", utils.StructToJson(reply)),
				)
			}
			if err != nil {
				ext.Error.Set(span, true)
				span.LogFields(log.String("event", "error"))
				span.LogFields(
					log.Object("evaError", utils.StructToJson(err)),
				)
			}
			return err
		} else {
			ctx, err := tracer.InjectTraceToContext(ctx)
			if err != nil {
				logger.GetLogger().Error(ctx, "链路错误", logger.Fields{"err": utils.StructToMap(err)})
			}
			err = invoker(ctx, method, req, reply, cc, opts...)
			return err
		}
	}
}
