package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type Signer struct {
	Key              ed25519.PrivateKey
	Issuer, Audience string
	Now              func() time.Time
}

func LoadKey(path string) (ed25519.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseEdPrivateKeyFromPEM(b)
	if err != nil {
		return nil, err
	}
	k, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("expected Ed25519 key")
	}
	return k, nil
}
func Random() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func Hash(raw string) string { h := sha256.Sum256([]byte(raw)); return hex.EncodeToString(h[:]) }
func (s Signer) Sign(client, scope string, ttl time.Duration) (string, error) {
	now := s.Now()
	id, err := Random()
	if err != nil {
		return "", err
	}
	claims := struct {
		Scope string `json:"scope"`
		jwt.RegisteredClaims
	}{
		Scope: scope, RegisteredClaims: jwt.RegisteredClaims{Issuer: s.Issuer, Audience: jwt.ClaimStrings{s.Audience}, Subject: client, ID: id, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl))}}
	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(s.Key)
}
