package redisstore

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"example.com/akuanaktehat/auth-service/internal/application"
	"example.com/akuanaktehat/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed rotate.lua
var rotateScript string
var createScript = redis.NewScript(`
if redis.call('EXISTS',KEYS[1],KEYS[2])~=0 then return 0 end
local ttl=tonumber(ARGV[2])
if not ttl or ttl<1 then return 0 end
redis.call('HSET',KEYS[1],'record',ARGV[1],'used','0')
redis.call('PEXPIRE',KEYS[1],ttl)
redis.call('SET',KEYS[2],'active','PX',ttl)
return 1
`)

type Store struct {
	Client  *redis.Client
	Timeout time.Duration
}

func tokenKey(hash string) string { return "auth:token:" + hash }
func familyKey(id string) string  { return "auth:family:" + id }
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.Client.Ping(ctx).Err()
}
func (s *Store) Create(ctx context.Context, r domain.RefreshRecord) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	b, err := json.Marshal(r)
	if err != nil {
		return application.ErrUnavailable
	}
	n, err := createScript.Run(ctx, s.Client, []string{tokenKey(r.Hash), familyKey(r.Family)}, string(b), time.Until(r.ExpiresAt).Milliseconds()).Int()
	if err != nil || n != 1 {
		return application.ErrUnavailable
	}
	return nil
}
func (s *Store) Get(ctx context.Context, hash string) (domain.RefreshRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	var r domain.RefreshRecord
	b, err := s.Client.HGet(ctx, tokenKey(hash), "record").Result()
	if errors.Is(err, redis.Nil) {
		return r, application.ErrGrant
	}
	if err != nil {
		return r, application.ErrUnavailable
	}
	if json.Unmarshal([]byte(b), &r) != nil || r.Hash != hash || r.Family == "" {
		return r, application.ErrUnavailable
	}
	return r, nil
}
func (s *Store) Rotate(ctx context.Context, old, next domain.RefreshRecord) error {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	ob, err := json.Marshal(old)
	if err != nil {
		return application.ErrUnavailable
	}
	nb, err := json.Marshal(next)
	if err != nil {
		return application.ErrUnavailable
	}
	if old.Family != next.Family || old.ClientID != next.ClientID || old.Scope != next.Scope || !old.ExpiresAt.Equal(next.ExpiresAt) {
		return application.ErrUnavailable
	}
	n, err := redis.NewScript(rotateScript).Run(ctx, s.Client, []string{tokenKey(old.Hash), tokenKey(next.Hash), familyKey(old.Family)}, string(ob), string(nb)).Int()
	if err != nil || n == -2 {
		return application.ErrUnavailable
	}
	if n != 1 {
		return application.ErrGrant
	}
	return nil
}
