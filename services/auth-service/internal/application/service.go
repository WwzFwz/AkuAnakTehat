package application

import (
	"context"
	"crypto/subtle"
	"errors"
	"example.com/akuanaktehat/auth-service/internal/domain"
	"example.com/akuanaktehat/auth-service/internal/token"
	"strings"
	"time"
)

var ErrCredentials = errors.New("invalid_client")
var ErrGrant = errors.New("invalid_grant")
var ErrUnavailable = errors.New("temporarily_unavailable")

type Store interface {
	Create(context.Context, domain.RefreshRecord) error
	Get(context.Context, string) (domain.RefreshRecord, error)
	Rotate(context.Context, domain.RefreshRecord, domain.RefreshRecord) error
	Ping(context.Context) error
}
type Signer interface {
	Sign(string, string, time.Duration) (string, error)
}
type Service struct {
	Clients               map[string]domain.Client
	Store                 Store
	Signer                Signer
	AccessTTL, RefreshTTL time.Duration
	Now                   func() time.Time
}

func (s *Service) Issue(ctx context.Context, id, secret string) (domain.TokenPair, error) {
	c, ok := s.Clients[id]
	if !ok || subtle.ConstantTimeCompare([]byte(c.SecretHash), []byte(token.Hash(secret))) != 1 {
		return domain.TokenPair{}, ErrCredentials
	}
	family, err := token.Random()
	if err != nil {
		return domain.TokenPair{}, ErrUnavailable
	}
	pair, record, err := s.prepare(id, strings.Join(c.Scopes, " "), family, s.Now().Add(s.RefreshTTL))
	if err != nil {
		return domain.TokenPair{}, err
	}
	if err = s.Store.Create(ctx, record); err != nil {
		return domain.TokenPair{}, ErrUnavailable
	}
	return pair, nil
}
func (s *Service) Refresh(ctx context.Context, raw string) (domain.TokenPair, error) {
	if len(raw) != 43 {
		return domain.TokenPair{}, ErrGrant
	}
	old, err := s.Store.Get(ctx, token.Hash(raw))
	if err != nil {
		return domain.TokenPair{}, err
	}
	if !s.Now().Before(old.ExpiresAt) {
		return domain.TokenPair{}, ErrGrant
	}
	pair, next, err := s.prepare(old.ClientID, old.Scope, old.Family, old.ExpiresAt)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if err = s.Store.Rotate(ctx, old, next); err != nil {
		return domain.TokenPair{}, err
	}
	return pair, nil
}
func (s *Service) prepare(id, scope, family string, expires time.Time) (domain.TokenPair, domain.RefreshRecord, error) {
	raw, err := token.Random()
	if err != nil {
		return domain.TokenPair{}, domain.RefreshRecord{}, ErrUnavailable
	}
	access, err := s.Signer.Sign(id, scope, s.AccessTTL)
	if err != nil {
		return domain.TokenPair{}, domain.RefreshRecord{}, ErrUnavailable
	}
	return domain.TokenPair{AccessToken: access, TokenType: "Bearer", ExpiresIn: int64(s.AccessTTL.Seconds()), RefreshToken: raw, Scope: scope},
		domain.RefreshRecord{Hash: token.Hash(raw), Family: family, ClientID: id, Scope: scope, ExpiresAt: expires}, nil
}
