package grpc

import (
	"context"
	"github.com/Gitforxuyang/protoreflect/dynamic/grpcdynamic"
	"github.com/Gitforxuyang/protoreflect/grpcreflect"
	"github.com/Gitforxuyang/protoreflect/grpcurl"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/protoreflect/proxy/reflection"
	"github.com/Gitforxuyang/protoreflect/proxy/stub"
	"github.com/Gitforxuyang/walle/middleware/catch"
	"github.com/Gitforxuyang/walle/middleware/log"
	"github.com/Gitforxuyang/walle/middleware/trace"
	etcd2 "github.com/Gitforxuyang/walle/selector/grpc"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"time"
)

type Proxy struct {
	cc         *grpc.ClientConn
	reflector  reflection.Reflector
	stub       stub.Stub
	descSource grpcurl.DescriptorSource
}

func NewTestProxy() *Proxy {
	conn, err := grpc.Dial(":50001",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		//grpc.WithBalancer(grpc.RoundRobin(etcd.NewResolver(""))),
		//grpc.WithKeepaliveParams(
		//	keepalive.ClientParameters{
		//		Time:                time.Second * 10,
		//		Timeout:             time.Second * 1,
		//		PermitWithoutStream: true,
		//	}),
		grpc.WithChainUnaryInterceptor(
			trace.NewClientWrapper(),
			log.NewClientWrapper(),
			catch.NewClientWrapper(5),
		),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.TODO()
	rc := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn))

	proxy := &Proxy{
		cc:         conn,
		reflector:  reflection.NewReflector(rc),
		stub:       stub.NewStub(grpcdynamic.NewStub(conn)),
		descSource: grpcurl.DescriptorSourceFromServer(ctx, rc),
	}
	return proxy
}

func (p *Proxy) Invoke(ctx context.Context,
	serviceName, methodName string,
	message []byte,
	md *metadata.Metadata,
) ([]byte, error) {
	invocation, err := p.reflector.CreateInvocation(serviceName, methodName, message)
	if err != nil {
		return nil, error2.GRpcError.SetDetail(err.Error())
	}

	outputMsg, err := p.stub.InvokeRPC(ctx, invocation, md)
	if err != nil {
		return nil, error2.DecodeStatus(err)
	}
	m, err := outputMsg.MarshalJSON()
	if err != nil {
		return nil, error2.GRpcError.SetDetail(err.Error())
	}
	return m, err
}

func NewProxy(service string) *Proxy {
	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		//grpc.WithBlock(),
		grpc.WithBalancer(grpc.RoundRobin(etcd2.NewResolver(service))),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                time.Second * 10,
				Timeout:             time.Second * 1,
				PermitWithoutStream: true,
			}),
		grpc.WithChainUnaryInterceptor(
			trace.NewClientWrapper(),
			log.NewClientWrapper(),
			catch.NewClientWrapper(5),
		),
	)
	ctx := context.TODO()
	if err != nil {
		logger.GetLogger().Error(ctx, "newProxy失败", logger.Fields{
			"err": err,
		})
	}
	rc := grpcreflect.NewClient(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn))
	proxy := &Proxy{
		cc:         conn,
		reflector:  reflection.NewReflector(rc),
		stub:       stub.NewStub(grpcdynamic.NewStub(conn)),
		descSource: grpcurl.DescriptorSourceFromServer(ctx, rc),
	}
	return proxy
}
