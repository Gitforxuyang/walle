package grpc

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/middleware/catch"
	"github.com/Gitforxuyang/walle/middleware/log"
	"github.com/Gitforxuyang/walle/middleware/trace"
	"github.com/Gitforxuyang/walle/util/logger"
	trace2 "github.com/Gitforxuyang/walle/util/trace"
	"github.com/fullstorydev/grpcurl"
	"github.com/gdong42/grpc-mate/metadata"
	"github.com/gdong42/grpc-mate/proxy/reflection"
	"github.com/gdong42/grpc-mate/proxy/stub"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	config.Init()
	conf := config.GetConfig()
	logger.Init(conf.GetName())
	trace2.Init(fmt.Sprintf("%s_%s", conf.GetName(), conf.GetENV()), conf.GetTraceConfig().Endpoint, conf.GetTraceConfig().Ratio)
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
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(ctx, "hello.SayHelloService", "Hello", []byte(`{"name":"dynamic","age":123}`), &md)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(resp))
	time.Sleep(time.Second * 3)
}