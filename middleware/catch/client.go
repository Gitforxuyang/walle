package catch

import (
	"context"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"google.golang.org/grpc"
	"time"
)

//用来将其它服务的返回错误转换为eva定义的错规范
func NewClientWrapper(timeout int64) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		deadline, _ := ctx.Deadline()
		//如果超时5s在deadline之后，则重置deadline为5s后
		if time.Now().Add(time.Second * time.Duration(timeout)).After(deadline) {
			ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			e := error2.DecodeStatus(err)
			err = e
		}
		return err
	}
}
