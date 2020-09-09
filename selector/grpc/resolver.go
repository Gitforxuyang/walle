package etcd

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/walle/config"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/naming"
)

const (
	//服务注册前缀
	ETCD_SERVICE_PREFIX = "/eva/service/"
)

type ServiceNode struct {
	Name     string `json:"name"`     //服务名
	Id       string `json:"id"`       //节点id 服务启动时随机生成的唯一id
	Endpoint string `json:"endpoint"` //服务的访问地址
}

type resolver struct {
	serviceName string // service name to resolve
}

// NewResolver return resolver with service name
func NewResolver(serviceName string) *resolver {
	return &resolver{serviceName: serviceName}
}

func (re *resolver) Resolve(target string) (naming.Watcher, error) {
	if re.serviceName == "" {
		panic("grpclb: no service name provided")
	}
	client := config.GetEtcdClient()
	return &watcher{re: re, client: client}, nil
}

type watcher struct {
	re            *resolver
	client        *clientv3.Client
	isInitialized bool
}

// Close do nothing
func (w *watcher) Close() {
}

// Next to return the updates
func (w *watcher) Next() ([]*naming.Update, error) {
	prefix := fmt.Sprintf("%s%s/", ETCD_SERVICE_PREFIX, w.re.serviceName)
	// check if is initialized
	if !w.isInitialized {
		resp, err := w.client.Get(context.Background(), prefix, clientv3.WithPrefix())
		w.isInitialized = true
		if err == nil {
			addrs := extractAddrs(resp)
			if l := len(addrs); l != 0 {
				updates := make([]*naming.Update, l)
				for i := range addrs {
					updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
				}
				return updates, nil
			}
		}
	}
	// generate etcd Watcher
	rch := w.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			node := ServiceNode{}
			err := utils.JsonToStruct(string(ev.Kv.Value), &node)
			utils.Must(err)
			switch ev.Type {
			case mvccpb.PUT:
				return []*naming.Update{{Op: naming.Add, Addr: node.Endpoint}}, nil
			case mvccpb.DELETE:
				return []*naming.Update{{Op: naming.Delete, Addr: node.Endpoint}}, nil
			}
		}
	}
	return nil, nil
}
func extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := []string{}
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			node := ServiceNode{}
			err := utils.JsonToStruct(string(v), &node)
			utils.Must(err)
			addrs = append(addrs, node.Endpoint)
		}
	}
	return addrs
}
