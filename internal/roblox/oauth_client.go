package roblox

import (
	"net/http"
	"sync"
	"time"
)

type OAuthClient struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       int64

	httpClient *http.Client
	tokenMu    sync.RWMutex
}

func NewOAuthClient(accessToken, refreshToken string, expiresAt time.Time, userID int64) *OAuthClient {
	return &OAuthClient{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserID:       userID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OAuthClient) DoRequest(req *http.Request) (*http.Response, error) {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	return c.httpClient.Do(req)
}

func (c *OAuthClient) SetAccessToken(token string, expiresAt time.Time) {
	c.tokenMu.Lock()
	c.AccessToken = token
	c.ExpiresAt = expiresAt
	c.tokenMu.Unlock()
}
