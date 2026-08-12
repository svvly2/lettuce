package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zalando/go-keyring"
)

const defaultRedirectURI = "http://localhost:38073/oauth/callback"
const keyringService = "com.lettuce.desktop"
const keyringAccount = "roblox-oauth-session"

type pendingRequest struct {
	State     string
	Verifier  string
	CreatedAt time.Time
}

type storedSession struct {
	RefreshToken string `json:"r"`
	UserID       int64  `json:"u"`
	Username     string `json:"n,omitempty"`
	DisplayName  string `json:"d,omitempty"`
	Picture      string `json:"p,omitempty"`
	Scope        string `json:"s,omitempty"`
}

type Manager struct {
	mu              sync.Mutex
	storeFile       string
	loadedSession   *Session
	pendings        map[string]pendingRequest
	onSessionUpdate func(*Session, error)
}

func NewManager(storeFile string) *Manager {
	m := &Manager{
		storeFile: storeFile,
		pendings:  make(map[string]pendingRequest),
	}
	_ = m.loadCurrentSession()
	return m
}

func (m *Manager) CurrentSession() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadedSession == nil {
		return nil
	}
	copy := *m.loadedSession
	return &copy
}

func (m *Manager) SetOnSessionUpdate(cb func(*Session, error)) {
	m.mu.Lock()
	m.onSessionUpdate = cb
	m.mu.Unlock()
}

func (m *Manager) ClearSession() error {
	m.mu.Lock()
	m.loadedSession = nil
	m.mu.Unlock()
	err := keyring.Delete(keyringService, keyringAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (m *Manager) loadCurrentSession() error {
	encoded, err := keyring.Get(keyringService, keyringAccount)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(encoded) == "" {
		return nil
	}

	var stored storedSession
	if err := json.Unmarshal([]byte(encoded), &stored); err != nil {
		return err
	}
	session := Session{RefreshToken: stored.RefreshToken, UserID: stored.UserID, Username: stored.Username, DisplayName: stored.DisplayName, Picture: stored.Picture, Scope: stored.Scope}
	m.mu.Lock()
	m.loadedSession = &session
	m.mu.Unlock()
	return nil
}

func (m *Manager) saveCurrentSession() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadedSession == nil {
		err := keyring.Delete(keyringService, keyringAccount)
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	stored := storedSession{RefreshToken: m.loadedSession.RefreshToken, UserID: m.loadedSession.UserID, Username: m.loadedSession.Username, DisplayName: m.loadedSession.DisplayName, Picture: m.loadedSession.Picture, Scope: m.loadedSession.Scope}
	data, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, keyringAccount, string(data))
}

func (m *Manager) SetCurrentSession(session *Session) error {
	m.mu.Lock()
	m.loadedSession = session
	m.mu.Unlock()
	return m.saveCurrentSession()
}

func (m *Manager) RefreshSessionIfNeeded() (*Session, error) {
	m.mu.Lock()
	session := m.loadedSession
	m.mu.Unlock()
	if session == nil {
		return nil, errors.New("oauth session not found")
	}
	if !session.ShouldRefresh() {
		return session, nil
	}

	redirectURI := getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = defaultRedirectURI
	}
	updated, err := RefreshToken(session.RefreshToken, redirectURI)
	if err != nil {
		return nil, err
	}
	if err := m.SetCurrentSession(updated); err != nil {
		return nil, err
	}
	if m.onSessionUpdate != nil {
		m.onSessionUpdate(updated, nil)
	}
	return updated, nil
}

func (m *Manager) PrepareLogin() (string, error) {
	verifier, challenge, err := NewPKCE()
	if err != nil {
		return "", err
	}

	redirectURI := getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = defaultRedirectURI
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URI: %w", err)
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return "", fmt.Errorf("oauth callback listener on port %s: %w", port, err)
	}

	state, err := randomURLSafeString(24)
	if err != nil {
		_ = listener.Close()
		return "", err
	}

	m.mu.Lock()
	m.pendings[state] = pendingRequest{State: state, Verifier: verifier, CreatedAt: time.Now()}
	m.mu.Unlock()

	srv := &http.Server{ReadHeaderTimeout: 10 * time.Second}
	mux := http.NewServeMux()
	srv.Handler = mux
	done := make(chan error, 1)

	mux.HandleFunc("/oauth/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<html><body style="font-family:system-ui;background:#050505;color:#fff;padding:32px"><h1>OAuth listener is running</h1><p>If you can read this, the local callback server is alive.</p></body></html>`)
	})

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		pending, ok := m.pendings[r.URL.Query().Get("state")]
		if ok {
			delete(m.pendings, r.URL.Query().Get("state"))
		}
		m.mu.Unlock()
		if !ok {
			http.Error(w, "invalid oauth session", http.StatusBadRequest)
			done <- errors.New("invalid oauth session")
			return
		}

		if errParam := strings.TrimSpace(r.URL.Query().Get("error")); errParam != "" {
			http.Error(w, fmt.Sprintf("oauth authorization failed: %s", errParam), http.StatusUnauthorized)
			done <- errors.New(errParam)
			return
		}
		if r.URL.Query().Get("state") != pending.State {
			http.Error(w, "invalid oauth state", http.StatusBadRequest)
			done <- errors.New("invalid oauth state")
			return
		}
		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			http.Error(w, "missing oauth authorization code", http.StatusBadRequest)
			done <- errors.New("missing oauth authorization code")
			return
		}

		session, err := ExchangeAuthorizationCode(code, pending.Verifier, redirectURI)
		if err != nil {
			http.Error(w, fmt.Sprintf("oauth code exchange failed: %s", err.Error()), http.StatusInternalServerError)
			done <- err
			if m.onSessionUpdate != nil {
				m.onSessionUpdate(nil, err)
			}
			return
		}

		if err := m.SetCurrentSession(session); err != nil {
			http.Error(w, fmt.Sprintf("failed to store oauth session: %s", err.Error()), http.StatusInternalServerError)
			done <- err
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
		done <- nil
	})

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			done <- err
		}
	}()

	clientID, _, err := oauthConfig()
	if err != nil {
		_ = srv.Close()
		return "", err
	}

	scopes := getConfig("oauth_scopes", "ROBUX_OAUTH_SCOPES")
	if scopes == "" {
		scopes = "openid profile asset:read asset:write"
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", scopes)
	params.Set("state", state)
	params.Set("nonce", state)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("prompt", "select_account consent")

	auth := authURL + "?" + params.Encode()

	go func() {
		select {
		case err := <-done:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				if m.onSessionUpdate != nil {
					m.onSessionUpdate(nil, err)
				}
			}
		case <-time.After(3 * time.Minute):
			_ = srv.Close()
			if m.onSessionUpdate != nil {
				m.onSessionUpdate(nil, errors.New("oauth login timed out"))
			}
		}
	}()

	return auth, nil
}

func (m *Manager) StartLogin() error {
	auth, err := m.PrepareLogin()
	if err != nil {
		return err
	}
	if err := openBrowser(auth); err != nil {
		return err
	}
	redirectURI := getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = defaultRedirectURI
	}
	u, _ := url.Parse(redirectURI)
	port := "38073"
	if u != nil && u.Port() != "" {
		port = u.Port()
	}
	go func() {
		time.Sleep(900 * time.Millisecond)
		_ = openBrowser("http://localhost:" + port + "/oauth/debug")
	}()
	return nil
}

func openBrowser(target string) error {
	switch runtime.GOOS {
	case "windows":
		if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start(); err == nil {
			return nil
		}
		if err := exec.Command("cmd", "/c", "start", "", fmt.Sprintf("\"%s\"", target)).Start(); err == nil {
			return nil
		}
		if err := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Start-Process '%s'", target)).Start(); err == nil {
			return nil
		}
		return exec.Command("explorer.exe", target).Start()
	case "darwin":
		return exec.Command("open", target).Start()
	default:
		return exec.Command("xdg-open", target).Start()
	}
}
