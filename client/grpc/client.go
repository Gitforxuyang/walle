package grpc

import (
	"context"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/logger"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

var (
	invokers map[string]*Proxy   = make(map[string]*Proxy)
	services map[string]*Service = make(map[string]*Service)
	initOnce sync.Once
	etcd     *clientv3.Client
)

type Service struct {
	Package string            `json:"package"`
	Name    string            `json:"name"`
	AppId   string            `json:"appId"`
	Methods map[string]Method `json:"methods"`
}
type Method struct {
	Req  map[string]string `json:"req"`
	Resp map[string]string `json:"resp"`
}

const (
	ETCD_WALLE_SERVICE_PREFIX = "/eva/walle/service/"
)

func Init() {
	initOnce.Do(func() {
		etcd := config.GetEtcdClient()
		res, err := etcd.Get(context.TODO(), ETCD_WALLE_SERVICE_PREFIX, clientv3.WithPrefix())
		utils.Must(err)
		for _, v := range res.Kvs {
			service := Service{}
			err := utils.JsonToStruct(string(v.Value), &service)
			utils.Must(err)
			proxy := NewProxy(service.AppId)
			invokers[service.Name] = proxy
			services[service.Name] = &service
		}
		go watch()
	})
}
func GetProxy(service string) *Proxy {
	return invokers[service]
}

func GetService(service string) *Service {
	return services[service]
}

func watch() {
	for {
		etcd := config.GetEtcdClient()
		chs := etcd.Watch(context.TODO(), ETCD_WALLE_SERVICE_PREFIX, clientv3.WithPrefix())
		for ch := range chs {
			for _, event := range ch.Events {
				switch event.Type {
				case mvccpb.PUT:
					service := Service{}
					err := utils.JsonToStruct(string(event.Kv.Value), &service)
					if err != nil {
						logger.GetLogger().Error(context.TODO(), "监听服务描述信息变化报错",
							logger.Fields{"err": err, "value": string(event.Kv.Value)})
						continue
					}
					if invokers[service.Name] == nil {
						proxy := NewProxy(service.Name)
						invokers[service.Name] = proxy
					} else {
						invokers[service.Name].reflector.RefreshDesc(service.Name)
					}
					services[service.Name] = &service
				case mvccpb.DELETE:

				}
			}
		}

	}
}
