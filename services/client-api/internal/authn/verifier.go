package authn

import (
	"crypto/ed25519"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
)

type Claims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}
type Verifier struct {
	Key              ed25519.PublicKey
	Issuer, Audience string
}

func LoadKey(path string) (ed25519.PublicKey, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	key, e := jwt.ParseEdPublicKeyFromPEM(b)
	if e != nil {
		return nil, e
	}
	k, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("expected Ed25519 public key")
	}
	return k, nil
}
func (v Verifier) Verify(raw string) (Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(raw, &c, func(t *jwt.Token) (any, error) { return v.Key, nil }, jwt.WithValidMethods([]string{"EdDSA"}), jwt.WithIssuer(v.Issuer), jwt.WithAudience(v.Audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return Claims{}, err
	}
	if c.Subject == "" || c.ID == "" || c.Scope == "" || c.IssuedAt == nil {
		return Claims{}, errors.New("missing claims")
	}
	return c, nil
}
