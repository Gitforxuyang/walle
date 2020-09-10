package server

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/walle/client/grpc"
	"github.com/Gitforxuyang/walle/client/redis"
	"github.com/Gitforxuyang/walle/config"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

//api控制器
type apiController struct {
	//访问地址对应的方法
	apis map[string]*Method
}

type Method struct {
	Service string   `json:"service"`
	Method  string   `json:"method"`
	Plugins []Plugin `json:"plugins"` //需要通过哪几个插件
}

type Plugin struct {
	Name  string   `json:"name"`
	Param []string `json:"param"`
}

var (
	apiOnce sync.Once
	api     *apiController
)

const (
	ETCD_API_KEY = "/eva/walle/api"
)

func InitApi() {
	apiOnce.Do(func() {
		api = &apiController{}
		api.apis = make(map[string]*Method)
		refreshApi()
		go watch()
		go task()
	})
}
func Api(ctx *gin.Context) (map[string]interface{}, error) {
	method := api.apis[fmt.Sprintf("%s_%s", ctx.Request.URL.Path, ctx.Request.Method)]
	if method == nil {
		return nil, error2.NotFoundError
	}
	proxy := grpc.GetProxy(method.Service)
	svc := grpc.GetService(method.Service)
	if proxy == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("proxy %s未找到", method.Service))
	}
	if svc == nil {
		return nil, error2.UnknowError.SetDetail(fmt.Sprintf("service %s未找到", method.Service))
	}
	r := ctx.Value("req")
	req := r.(map[string]interface{})
	req["realIp"] = ctx.Request.Header.Get("X-Real-IP")
	//对需要经过的插件依次执行
	for _, v := range method.Plugins {
		switch v.Name {
		case "checkSSO":
			uid, err := checkSSO(ctx, ctx.Request.Header.Get("token"))
			if err != nil {
				return nil, err
			}
			req["uid"] = uid
		}
	}
	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(ctx, fmt.Sprintf("%s.%s", svc.Package, svc.Name), method.Method, []byte(utils.MapToJson(req)), &md)
	if err != nil {
		return nil, err
	}
	res, err := utils.JsonToMap(string(resp))
	if err != nil {
		return nil, error2.UnknowError.SetDetail(err.Error())
	}
	return res, nil
}

//刷新api
func refreshApi() {
	client := redis.GetRedisClient()
	res := client.HGetAll(context.TODO(), "walle:api")
	if res.Err() != nil {
		logger.GetLogger().Error(context.TODO(), "获取redis报错", logger.Fields{"err": res.Err()})
		return
	}
	maps, err := res.Result()
	if err != nil {
		logger.GetLogger().Error(context.TODO(), "获取redis报错", logger.Fields{"err": err})
		return
	}
	for k, v := range maps {
		m := Method{}
		err := utils.JsonToStruct(v, &m)
		if err != nil {
			logger.GetLogger().Error(context.TODO(), "api转json报错", logger.Fields{"err": err})
			continue
		}
		api.apis[k] = &m
	}
	logger.GetLogger().Info(context.TODO(), "刷新api成功", logger.Fields{
		"api": api.apis,
	})
}
func watch() {
	etcd := config.GetEtcdClient()
	log := logger.GetLogger()
	for {
		chs := etcd.Watch(context.TODO(), ETCD_API_KEY)
		for v := range chs {
			for _, event := range v.Events {
				version := string(event.Kv.Value)
				refreshApi()
				log.Info(context.TODO(), "etcd更新api", logger.Fields{
					"version": version,
				})
			}
		}
	}
}

//定时刷新
func task() {
	for {
		<-time.Tick(time.Second * 30)
		refreshApi()
	}
}
