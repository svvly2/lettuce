package oauth

import "time"

type Session struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
	Picture      string    `json:"picture,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

func (s *Session) ShouldRefresh() bool {
	if s == nil {
		return false
	}
	return time.Until(s.ExpiresAt) < 2*time.Minute
}

func (s *Session) IsValid() bool {
	if s == nil {
		return false
	}
	return s.AccessToken != "" && time.Now().Before(s.ExpiresAt)
}
