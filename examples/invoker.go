package main

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/protoreflect/metadata"
	grpc2 "github.com/Gitforxuyang/walle/client/grpc"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	trace2 "github.com/Gitforxuyang/walle/util/trace"
)

func main() {
	config.Init()
	conf := config.GetConfig()
	logger.Init(conf.GetName())
	trace2.Init(fmt.Sprintf("%s_%s", conf.GetName(), conf.GetENV()), conf.GetTraceConfig().Endpoint, conf.GetTraceConfig().Ratio)
	proxy := grpc2.NewTestProxy()
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(context.TODO(), "hello.SayHelloService", "Hello", []byte(`{"name1":"dynamic","age":"123"}`), &md)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(resp))
}
