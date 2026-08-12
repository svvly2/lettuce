package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/svvly2/lettuce/internal/oauth"
	"github.com/svvly2/lettuce/internal/roblox"
)

// clientForStudioUpload refreshes the PKCE session before each Studio job.
// The plugin never receives this client or either token; it only receives the
// old-to-new asset ID mappings returned by the local bridge.
func (s *AppState) clientForStudioUpload() (*roblox.Client, error) {
	s.mu.Lock()
	client := s.Client
	manager := s.OAuthManager
	s.mu.Unlock()

	if client == nil {
		return nil, fmt.Errorf("sign in with Roblox OAuth first")
	}
	if !client.IsOAuthOnly() {
		return client, nil
	}
	if manager == nil {
		return nil, fmt.Errorf("oauth manager not initialized")
	}

	session, err := manager.RefreshSessionIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("refresh Roblox OAuth session: %w", err)
	}
	if session == nil || strings.TrimSpace(session.AccessToken) == "" {
		return nil, fmt.Errorf("Roblox OAuth session has no access token")
	}

	refreshed := roblox.NewOAuthAuthenticatedClient(session.AccessToken, roblox.UserInfo{
		ID:          session.UserID,
		Username:    session.Username,
		DisplayName: session.DisplayName,
		Picture:     session.Picture,
	})
	s.mu.Lock()
	s.Client = refreshed
	s.RefreshToken = session.RefreshToken
	s.IsLoggedIn = true
	s.mu.Unlock()
	return refreshed, nil
}

func (s *AppState) StartOAuthLogin() error {
	if s.OAuthManager == nil {
		return fmt.Errorf("oauth manager not initialized")
	}

	s.OAuthManager.SetOnSessionUpdate(s.applyOAuthSession)

	if err := s.OAuthManager.StartLogin(); err != nil {
		return err
	}

	s.AddLog("info", "Opening Roblox OAuth login in your default browser.")
	return nil
}

func (s *AppState) applyOAuthSession(session *oauth.Session, err error) {
	if err != nil || session == nil {
		return
	}
	s.mu.Lock()
	s.Client = roblox.NewOAuthAuthenticatedClient(session.AccessToken, roblox.UserInfo{ID: session.UserID, Username: session.Username, DisplayName: session.DisplayName, Picture: session.Picture})
	s.RefreshToken = session.RefreshToken
	s.IsLoggedIn = true
	cb := s.OnState
	s.mu.Unlock()
	if cb != nil {
		cb()
	}
	s.AddLog("success", "Successfully logged in with Roblox OAuth.")
	// Let the short-lived OAuth callback listener release the shared port first.
	go func() {
		time.Sleep(600 * time.Millisecond)
		if err := s.StartServer(); err != nil && !s.Snapshot().ServerRunning {
			s.AddLog("error", "Could not start Studio bridge: "+err.Error())
		}
	}()
}
