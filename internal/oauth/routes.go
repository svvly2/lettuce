package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/svvly2/lettuce/internal/roblox"
	"github.com/svvly2/lettuce/internal/roblox/ide"
)

func (m *Manager) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/oauth/start", m.handleStart)
	mux.HandleFunc("/oauth/callback", m.handleCallback)
	mux.HandleFunc("/oauth/status", m.handleStatus)
	mux.HandleFunc("/oauth/upload-animation", m.handleUploadAnimation)
}

func (m *Manager) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, err := randomURLSafeString(24)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	verifier, challenge, err := NewPKCE()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.mu.Lock()
	m.pendings[state] = pendingRequest{State: state, Verifier: verifier, CreatedAt: time.Now()}
	m.mu.Unlock()

	clientID := getConfig("oauth_client_id", "ROBUX_OAUTH_CLIENT_ID")
	redirectURI := getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	scopes := getConfig("oauth_scopes", "ROBUX_OAUTH_SCOPES")
	if clientID == "" || redirectURI == "" {
		http.Error(w, ErrOAuthConfigMissing.Error(), http.StatusInternalServerError)
		return
	}
	if scopes == "" {
		scopes = "asset:write"
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", scopes)
	params.Set("state", state)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("prompt", "select_account consent")

	http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound)
}

func (m *Manager) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		http.Error(w, "missing oauth state", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	pending, ok := m.pendings[state]
	if ok {
		delete(m.pendings, state)
	}
	m.mu.Unlock()
	if !ok {
		http.Error(w, "invalid oauth session", http.StatusBadRequest)
		return
	}

	if errParam := strings.TrimSpace(r.URL.Query().Get("error")); errParam != "" {
		http.Error(w, fmt.Sprintf("oauth authorization failed: %s", errParam), http.StatusUnauthorized)
		return
	}

	if r.URL.Query().Get("state") != pending.State {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		http.Error(w, "missing oauth authorization code", http.StatusBadRequest)
		return
	}

	redirectURI := getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	session, err := ExchangeAuthorizationCode(code, pending.Verifier, redirectURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("oauth code exchange failed: %s", err.Error()), http.StatusInternalServerError)
		if m.onSessionUpdate != nil {
			m.onSessionUpdate(nil, err)
		}
		return
	}

	if err := m.SetCurrentSession(session); err != nil {
		http.Error(w, fmt.Sprintf("failed to store oauth session: %s", err.Error()), http.StatusInternalServerError)
		if m.onSessionUpdate != nil {
			m.onSessionUpdate(nil, err)
		}
		return
	}

	if m.onSessionUpdate != nil {
		m.onSessionUpdate(session, nil)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<html><body><h1>Roblox OAuth login complete</h1><p>You may now return to the app.</p></body></html>`)
}

func (m *Manager) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := m.CurrentSession()
	status := struct {
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id,omitempty"`
		ExpiresAt     string `json:"expires_at,omitempty"`
	}{
		Authenticated: session != nil,
	}
	if session != nil {
		status.UserID = session.UserID
		status.ExpiresAt = session.ExpiresAt.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (m *Manager) handleUploadAnimation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := m.RefreshSessionIfNeeded()
	if err != nil {
		http.Error(w, fmt.Sprintf("oauth session error: %s", err.Error()), http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, fmt.Sprintf("invalid upload request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	creatorValue := strings.TrimSpace(r.FormValue("creator"))
	if name == "" {
		http.Error(w, "animation name is required", http.StatusBadRequest)
		return
	}

	creatorID := session.UserID
	isGroup := false
	if creatorValue != "" && !strings.EqualFold(creatorValue, "user") {
		if strings.HasPrefix(strings.ToLower(creatorValue), "group:") {
			trimmed := strings.TrimSpace(creatorValue[len("group:"):])
			creatorID, err = strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				http.Error(w, "creator group id is invalid", http.StatusBadRequest)
				return
			}
			isGroup = true
		} else {
			creatorID, err = strconv.ParseInt(creatorValue, 10, 64)
			if err != nil {
				http.Error(w, "creator value must be 'user' or 'group:<id>'", http.StatusBadRequest)
				return
			}
			isGroup = true
		}
	} else if creatorID == 0 {
		http.Error(w, "creator is required because this OAuth token did not include a user id", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "animation file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var body bytes.Buffer
	if _, err := io.Copy(&body, file); err != nil {
		http.Error(w, fmt.Sprintf("failed to read animation file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	oauthClient := roblox.NewOAuthClient(session.AccessToken, session.RefreshToken, session.ExpiresAt, session.UserID)
	uploadFn, err := ide.NewUploadAnimationOAuthHandler(oauthClient, name, description, &body, creatorID, isGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("upload setup failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	assetID, err := uploadFn()
	if err != nil {
		status := http.StatusBadRequest
		var errText = err.Error()
		if strings.Contains(errText, "unauthorized") || strings.Contains(errText, "not authenticated") {
			status = http.StatusUnauthorized
		}
		if strings.Contains(strings.ToLower(errText), "permission") || strings.Contains(strings.ToLower(errText), "scope") {
			status = http.StatusForbidden
		}
		http.Error(w, errText, status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"assetId": assetID})
}
