package server

import (
	"fmt"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/walle/client/grpc"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/gin-gonic/gin"
	"sync"
)

//api控制器
type apiController struct {
	//访问地址对应的方法
	apis map[string]*method
}

type method struct {
	service string   `json:"service"`
	method  string   `json:"method"`
	plugins []plugin `json:"plugins"` //需要通过哪几个插件
}

type plugin struct {
	name  string `json:"name"`
	param string `json:"param"`
}

var (
	apiOnce sync.Once
	api     *apiController
)

func InitApi() {
	apiOnce.Do(func() {
		api = &apiController{}
		api.apis = make(map[string]*method)
	})
}
func Api(ctx *gin.Context) (map[string]interface{}, error) {
	method := api.apis[fmt.Sprintf("%s_%s", ctx.Request.URL.Path, ctx.Request.Method)]
	if method == nil {
		return nil, error2.NotFoundError
	}
	proxy := grpc.GetProxy(method.service)
	svc := grpc.GetService(method.service)
	if proxy == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("proxy %s未找到", method.service))
	}
	if svc == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("service %s未找到", method.service))
	}
	r := ctx.Value("req")
	req := r.(map[string]interface{})
	req["realIp"] = ctx.Request.Header.Get("X-Real-IP")
	//对需要经过的插件依次执行
	for _, v := range method.plugins {
		switch v.name {
		case "checkSSO":
			uid, err := checkSSO(ctx, ctx.Request.Header.Get("token"))
			if err != nil {
				return nil, err
			}
			req["uid"] = uid
		}
	}
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(ctx, fmt.Sprintf("%s.%s", svc.Package, svc.Name), method.method, []byte(utils.MapToJson(req)), &md)
	if err != nil {
		return nil, err
	}
	res, err := utils.JsonToMap(string(resp))
	if err != nil {
		return nil, error2.UnknowError.SetDetail(err.Error())
	}
	return res, nil
}
