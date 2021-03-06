package config

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Gitforxuyang/walle/util/utils"
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"strings"
	"sync"
	"time"
)

//配置发送变更时的通知
type ChangeNotify func(config map[string]interface{})

type TraceConfig struct {
	Endpoint   string
	Ratio      float64
	GRpcClient bool //grpc client是否链路
	HttpClient bool //http client是否链路
	Log        bool //链路时是否输出更详细的log
}

type GRpcClientConfig struct {
	Mode     string //dns etcd
	Endpoint string
	Timeout  int64 //请求超时时间
}
type HttpClientConfig struct {
	Endpoint string
	Timeout  int64
	MaxConn  int //最大连接数
}
type LogConfig struct {
	Server     bool   //服务端日志是否打印
	GRpcClient bool   //grpc客户端日志
	HttpClient bool   //http客户端日志
	Level      string //日志打印级别
}
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
type EvaConfig struct {
	name              string
	port              int32
	env               string
	config            map[string]interface{}
	v                 *viper.Viper
	changeNotifyFuncs []ChangeNotify
	grpc              map[string]*GRpcClientConfig
	http              map[string]*HttpClientConfig
	log               *LogConfig
	trace             *TraceConfig
	etcd              []string
	sentryDSN         string
	redis             map[string]*RedisConfig
}

var (
	config     *EvaConfig
	configOnce sync.Once
	etcdClient *clientv3.Client
)

func Init() {
	configOnce.Do(func() {
		config = &EvaConfig{}
		config.config = make(map[string]interface{})
		config.changeNotifyFuncs = make([]ChangeNotify, 0)
		config.grpc = make(map[string]*GRpcClientConfig)
		config.redis = make(map[string]*RedisConfig)
		config.trace = &TraceConfig{}
		config.log = &LogConfig{Server: false, GRpcClient: false, HttpClient: false, Level: "INFO"}
		config.etcd = make([]string, 0, 3)
		v := viper.New()
		v.SetConfigName("config.default")
		v.AddConfigPath("./conf")
		v.SetConfigType("json")
		err := v.ReadInConfig()
		utils.Must(err)
		v.BindEnv("ENV")
		env := v.GetString("ENV")
		if env == "" {
			env = "local"
		}
		config.env = env
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		err = v.MergeInConfig()
		utils.Must(err)
		config.name = v.GetString("name")
		if config.name == "" {
			panic("配置文件中name不能为空")
		}
		err = v.UnmarshalKey("etcd", &config.etcd)
		utils.Must(err)
		if len(config.etcd) == 0 {
			panic("配置文件中etcd不能为空")
		}

		client, err := clientv3.New(clientv3.Config{
			Endpoints:   config.etcd,
			DialTimeout: time.Second * 3,
		})
		utils.Must(err)
		etcdClient = client
		resp, err := client.Get(context.TODO(), fmt.Sprintf("%s%s", ETCD_CONFIG_PREFIX, "global"))
		utils.Must(err)
		if len(resp.Kvs) == 0 {
			panic("配置中心global未找到")
		}
		v.MergeConfig(bytes.NewBuffer(resp.Kvs[0].Value))
		resp, err = client.Get(context.TODO(), fmt.Sprintf("%s%s", ETCD_CONFIG_PREFIX, config.name))
		utils.Must(err)
		if len(resp.Kvs) == 0 {
			panic(fmt.Sprintf("配置中心%s未找到", config.name))
		}
		v.MergeConfig(bytes.NewBuffer(resp.Kvs[0].Value))

		config.port = v.GetInt32("port")
		if config.port == 0 {
			panic("配置文件中port不能为空")
		}
		config.v = v
		err = v.UnmarshalKey("grpc", &config.grpc)
		utils.Must(err)
		err = v.UnmarshalKey("log", &config.log)
		utils.Must(err)
		err = v.UnmarshalKey("http", &config.http)
		utils.Must(err)
		err = v.UnmarshalKey("redis", &config.redis)
		utils.Must(err)
		if utils.IsNil(v.Get("trace")) {
			panic("trace设置不能为空")
		}
		err = v.UnmarshalKey("trace", &config.trace)
		utils.Must(err)
		config.sentryDSN = v.GetString("sentry")
		watch(client)
	})
}

func GetConfig() *EvaConfig {
	return config
}

func GetEtcdClient() *clientv3.Client {
	return etcdClient
}
func RegisterNotify(f ChangeNotify) {
	config.changeNotifyFuncs = append(config.changeNotifyFuncs, f)
}

func (m *EvaConfig) changeNotify(config map[string]interface{}) {
	for _, v := range m.changeNotifyFuncs {
		v(config)
	}
}
func (m *EvaConfig) GetName() string {
	return m.name
}
func (m *EvaConfig) GetPort() int32 {
	return m.port
}
func (m *EvaConfig) GetENV() string {
	return m.env
}

func (m *EvaConfig) GetTraceConfig() *TraceConfig {
	return m.trace
}

func (m *EvaConfig) GetGRpc(app string) *GRpcClientConfig {
	c := m.grpc[strings.ToLower(app)]
	if c == nil {
		//如果没有配置，则默认直接用appId来访问。
		return &GRpcClientConfig{Mode: "etcd", Endpoint: app, Timeout: 5}
	}
	return c
}
func (m *EvaConfig) GetHttp(http string) *HttpClientConfig {
	c := m.http[strings.ToLower(http)]
	if c == nil {
		panic(fmt.Sprintf("http：%s配置未找到", http))
	}
	return c
}

func (m *EvaConfig) GetLogConfig() *LogConfig {
	if m.log == nil {
		panic(fmt.Sprintf("log配置未找到"))
	}
	return m.log
}

func (m *EvaConfig) GetEtcd() []string {
	if m.etcd == nil {
		panic(fmt.Sprintf("etcd配置未找到"))
	}
	return m.etcd
}
func GetSentry() string {
	return config.sentryDSN
}

//func GetDynamic() map[string]interface{} {
//	return config.dynamic
//}
func (m *EvaConfig) GetRedis(name string) *RedisConfig {
	c := m.redis[strings.ToLower(name)]
	if c == nil {
		panic(fmt.Sprintf("redis：%s配置未找到", name))
	}
	return c
}
