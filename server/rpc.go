package server

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/walle/client/grpc"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/utils"
)

func Rpc(ctx context.Context, service string, method string, req interface{}) (map[string]interface{}, error) {
	proxy := grpc.GetProxy(service)
	svc := grpc.GetService(service)
	if proxy == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("proxy %s未找到", service))
	}
	if svc == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("service %s未找到", service))
	}
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(ctx, fmt.Sprintf("%s.%s", svc.Package, svc.Name), method, []byte(utils.MapToJson(req)), &md)
	if err != nil {
		return nil, err
	}
	r, err := utils.JsonToMap(string(resp))
	if err != nil {
		return nil, error2.UnknowError.SetDetail(err.Error())
	}
	return r, nil
}
