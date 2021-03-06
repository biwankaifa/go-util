package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang-module/carbon"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	tracinglog "github.com/opentracing/opentracing-go/log"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//    rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

// Nil reply returned by Redis when key does not exist.
const Nil = redis.Nil

type ConfigOfRedis struct {
	Database int
	Address  string
	Password string
	RunMode  string // 允许模式
}

// client Redis单例模式
var client map[int]*redis.Client
var mu sync.Mutex
var cfg *ConfigOfRedis

func (c *ConfigOfRedis) InitRedis() {
	client = make(map[int]*redis.Client)
	cfg = c
}

//Get 只执行一次
func Get(i ...int) *redis.Client {
	var db int
	if len(i) <= 0 {
		db = cfg.Database
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
				Addr:         cfg.Address,
				Password:     cfg.Password, // no password set
				DB:           db,           // use default DB
				MaxRetries:   3,
				PoolSize:     10,
				MinIdleConns: 5,
			})
			client[db].AddHook(&ClientHook{
				RunMode: cfg.RunMode,
			})
		}
	}

	return client[db]
}

type ClientHook struct {
	RunMode string
}

func (c ClientHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "processStartTime", time.Now())

	if opentracing.IsGlobalTracerRegistered() {
		_, ctx = opentracing.StartSpanFromContext(ctx, "redis")
	}
	return ctx, nil
}

func (c ClientHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	_, file, line, _ := runtime.Caller(4)

	span := opentracing.SpanFromContext(ctx)

	if span != nil && cmd.Name() != "ping" {
		defer span.Finish()
		ext.Component.Set(span, "redis")
		span.LogFields(tracinglog.Object("statement", fmt.Sprintf("%v", cmd.String())))
		span.LogFields(tracinglog.Object("file", fmt.Sprintf("%s:%s", file, strconv.FormatInt(int64(line), 10))))
		if err := cmd.Err(); err != nil && !errors.Is(err, Nil) {
			ext.Error.Set(span, true)
			span.LogFields(tracinglog.Object("err", err))
			return err
		}
	}

	if c.RunMode == "debug" {
		processStartTime := ctx.Value("processStartTime").(time.Time)
		elapsed := time.Since(processStartTime)
		fmt.Printf("\n%s %s\n\u001B[34m[Redis]\u001B[0m \u001B[33m[%.3fms]\u001B[0m %v\n", carbon.Now().ToDateTimeString(), file+":"+strconv.FormatInt(int64(line), 10), float64(elapsed.Nanoseconds())/1e6, cmd.String())
	}
	return nil
}

func (c ClientHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "processPipelineStartTime", time.Now())

	if opentracing.IsGlobalTracerRegistered() {
		_, ctx = opentracing.StartSpanFromContext(ctx, "redis")
	}

	return ctx, nil
}

func (c ClientHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	var s []interface{}
	errs := make([]error, 0)
	for _, cmd := range cmds {
		s = append(s, cmd.String())
		if cmd.Err() != nil && !errors.Is(cmd.Err(), Nil) {
			errs = append(errs, cmd.Err())
		}

	}
	_, file, line, _ := runtime.Caller(5)

	span := opentracing.SpanFromContext(ctx)

	if span != nil {
		defer span.Finish()
		ext.Component.Set(span, "redis")
		span.LogFields(tracinglog.Object("statement", fmt.Sprintf("%v", s)))
		span.LogFields(tracinglog.Object("file", fmt.Sprintf("%s:%s", file, strconv.FormatInt(int64(line), 10))))

		if len(errs) > 0 {
			ext.Error.Set(span, true)
			span.LogFields(tracinglog.Object("err:", errs))
		}
	}

	if c.RunMode == "debug" {
		processPipelineStartTime := ctx.Value("processPipelineStartTime").(time.Time)
		elapsed := time.Since(processPipelineStartTime)
		fmt.Printf("\n%s %s\n\u001B[34m[Redis]\u001B[0m \u001B[33m[%.3fms]\u001B[0m %v\n", carbon.Now().ToDateTimeString(), file+":"+strconv.FormatInt(int64(line), 10), float64(elapsed.Nanoseconds())/1e6, s)
	}
	return nil
}
