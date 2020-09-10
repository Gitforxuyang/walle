package redis

import (
	"github.com/Gitforxuyang/walle/config"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

var (
	rdb       *redis.Client
	redisOnce sync.Once
)

func Init() {
	redisOnce.Do(func() {
		conf := config.GetConfig()
		c := conf.GetRedis("walle")
		rdb = redis.NewClient(&redis.Options{
			Addr:         c.Addr,
			Password:     c.Password, // no password set
			DB:           c.DB,       // use default DB
			PoolSize:     c.PoolSize,
			MinIdleConns: c.MinIdleConns,
			DialTimeout:  time.Second * time.Duration(c.DialTimeout),
			ReadTimeout:  time.Second * time.Duration(c.ReadTimeout),
			WriteTimeout: time.Second * time.Duration(c.WriteTimeout),
		})
	})
}

func GetRedisClient() *redis.Client {
	return rdb
}
