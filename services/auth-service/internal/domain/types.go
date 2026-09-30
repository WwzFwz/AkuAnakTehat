package domain

import "time"

type Client struct {
	ID         string   `json:"client_id"`
	SecretHash string   `json:"secret_hash"`
	Scopes     []string `json:"scopes"`
}
type RefreshRecord struct {
	Hash      string    `json:"hash"`
	Family    string    `json:"family"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
}
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}
