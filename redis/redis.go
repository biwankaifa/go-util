package redis

import (
	"context"
	"fmt"
	"github.com/biwankaifa/go-utilconfig"
	"github.com/go-redis/redis/v8"
	"github.com/golang-module/carbon"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//
//    rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

const Nil = redis.Nil

// client Redis单例模式
var client map[int]*redis.Client
var mu sync.Mutex

func init() {
	client = make(map[int]*redis.Client)
}

//Get 只执行一次
func Get(i ...int) *redis.Client {
	c := config.Get().Redis
	var db int
	if len(i) <= 0 {
		db = c.Db
	} else {
		db = i[0]
	}

	if db > 15 {
		db = 15
	}
	if db < 0 {
		db = 0
	}

	if client[db] == nil {
		mu.Lock()
		defer mu.Unlock()
		if client[db] == nil {
			client[db] = redis.NewClient(&redis.Options{
				Addr:         c.Addr,
				Password:     c.Pass, // no password set
				DB:           db,     // use default DB
				MaxRetries:   3,
				PoolSize:     10,
				MinIdleConns: 5,
			})
			client[db].AddHook(&ClientHook{})
		}
	}

	return client[db]
}

type ClientHook struct{}

var processStartTime time.Time

func (c ClientHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	processStartTime = time.Now()
	return ctx, nil
}

func (c ClientHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	_, file, line, _ := runtime.Caller(4)
	elapsed := time.Since(processStartTime)
	fmt.Printf("\n%s %s\n\u001B[34m[Redis]\u001B[0m \u001B[33m[%.3fms]\u001B[0m %v\n", carbon.Now().ToDateTimeString(), file+":"+strconv.FormatInt(int64(line), 10), float64(elapsed.Nanoseconds())/1e6, cmd.String())
	return nil
}

var processPipelineStartTime time.Time

func (c ClientHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	processPipelineStartTime = time.Now()
	return ctx, nil
}

func (c ClientHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	var s []interface{}
	for _, cmd := range cmds {
		s = append(s, cmd.String())
	}
	_, file, line, _ := runtime.Caller(5)
	elapsed := time.Since(processPipelineStartTime)
	fmt.Printf("\n%s %s\n\u001B[34m[Redis]\u001B[0m \u001B[33m[%.3fms]\u001B[0m %v\n", carbon.Now().ToDateTimeString(), file+":"+strconv.FormatInt(int64(line), 10), float64(elapsed.Nanoseconds())/1e6, s)
	return nil
}
