package observability

import (
	"context"
	"github.com/redis/go-redis/v9"
	"net"
	"time"
)

type RedisHook struct{}

func (RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)
		redisLog(ctx, "connect", start, err)
		return conn, err
	}
}
func (RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		redisLog(ctx, cmd.Name(), start, err)
		return err
	}
}
func (RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		redisLog(ctx, "pipeline", start, err)
		return err
	}
}
func redisLog(ctx context.Context, op string, start time.Time, err error) {
	id := ID(ctx)
	if id == "" {
		id = "redis-background"
	}
	Logger().Info("outbound_call", "dependency", "redis", "operation", op, "correlation_id", id, "latency_ms", float64(time.Since(start).Microseconds())/1000, "ok", err == nil)
}
