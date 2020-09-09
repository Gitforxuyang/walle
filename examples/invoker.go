package main

import (
	"context"
	"fmt"
	grpc2 "github.com/Gitforxuyang/walle/client/grpc"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	trace2 "github.com/Gitforxuyang/walle/util/trace"
	"github.com/gdong42/grpc-mate/metadata"
)

func main() {
	config.Init()
	conf := config.GetConfig()
	logger.Init(conf.GetName())
	trace2.Init(fmt.Sprintf("%s_%s", conf.GetName(), conf.GetENV()), conf.GetTraceConfig().Endpoint, conf.GetTraceConfig().Ratio)
	proxy := grpc2.NewTestProxy()
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(context.TODO(), "hello.SayHelloService", "Hello", []byte(`{"name":"dynamic","age":"abc"}`), &md)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(resp))
}
