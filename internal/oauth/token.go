package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	appconfig "github.com/svvly2/lettuce/internal/app/config"
)

const (
	authURL     = "https://apis.roblox.com/oauth/v1/authorize"
	tokenURL    = "https://apis.roblox.com/oauth/v1/token"
	userInfoURL = "https://apis.roblox.com/oauth/v1/userinfo"
)

var ErrOAuthConfigMissing = errors.New("oauth_client_id and oauth_redirect_uri must be configured")

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	UserID       int64  `json:"user_id,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

type userInfoResponse struct {
	Subject           string `json:"sub"`
	UserID            int64  `json:"user_id,omitempty"`
	Name              string `json:"name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

func getConfig(key, envKey string) string {
	if value := strings.TrimSpace(appconfig.Get(key)); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(strings.ReplaceAll(envKey, "ROBUX_", "ROBLOX_")))
}

func oauthConfig() (clientID, redirectURI string, err error) {
	clientID = getConfig("oauth_client_id", "ROBUX_OAUTH_CLIENT_ID")
	redirectURI = getConfig("oauth_redirect_uri", "ROBUX_OAUTH_REDIRECT_URI")
	if clientID == "" || redirectURI == "" {
		return "", "", ErrOAuthConfigMissing
	}
	return clientID, redirectURI, nil
}

func ExchangeAuthorizationCode(code, codeVerifier, redirectURI string) (*Session, error) {
	clientID, _, err := oauthConfig()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", clientID)
	values.Set("code", code)
	values.Set("redirect_uri", redirectURI)
	values.Set("code_verifier", codeVerifier)

	token, err := requestToken(values)
	if err != nil {
		return nil, err
	}

	userID := int64(0)
	var info userInfoResponse
	if token.UserID != 0 {
		userID = token.UserID
	}
	info, err = fetchUserInfo(token.AccessToken)
	if err == nil && userID == 0 {
		userID = info.UserID
		if userID == 0 {
			userID, _ = parseSubject(info.Subject)
		}
	}

	return &Session{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		UserID:       userID,
		Username:     info.PreferredUsername,
		DisplayName:  firstNonEmpty(info.Nickname, info.Name, info.PreferredUsername),
		Picture:      info.Picture,
		IDToken:      token.IDToken,
		Scope:        token.Scope,
	}, nil
}

func RefreshToken(refreshToken, redirectURI string) (*Session, error) {
	clientID, _, err := oauthConfig()
	if err != nil {
		return nil, err
	}
	if refreshToken == "" {
		return nil, errors.New("refresh token is not available")
	}

	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", clientID)
	values.Set("refresh_token", refreshToken)
	values.Set("redirect_uri", redirectURI)

	token, err := requestToken(values)
	if err != nil {
		return nil, err
	}

	userID := int64(0)
	var info userInfoResponse
	if token.UserID != 0 {
		userID = token.UserID
	}
	info, err = fetchUserInfo(token.AccessToken)
	if err == nil && userID == 0 {
		userID = info.UserID
		if userID == 0 {
			userID, _ = parseSubject(info.Subject)
		}
	}

	return &Session{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		UserID:       userID,
		Username:     info.PreferredUsername,
		DisplayName:  firstNonEmpty(info.Nickname, info.Name, info.PreferredUsername),
		Picture:      info.Picture,
		IDToken:      token.IDToken,
		Scope:        token.Scope,
	}, nil
}

func requestToken(values url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token tokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("token response decode failed: %w", err)
	}
	if token.Error != "" {
		detail := strings.TrimSpace(token.ErrorDesc)
		if detail != "" {
			return nil, fmt.Errorf("oauth token error: %s: %s", token.Error, detail)
		}
		return nil, fmt.Errorf("oauth token error: %s", token.Error)
	}
	if token.AccessToken == "" {
		return nil, errors.New("oauth token response missing access_token")
	}

	return &token, nil
}

func fetchUserInfo(accessToken string) (userInfoResponse, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, userInfoURL, http.NoBody)
	if err != nil {
		return userInfoResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return userInfoResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return userInfoResponse{}, fmt.Errorf("userinfo request failed: %s", strings.TrimSpace(string(body)))
	}

	var info userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return userInfoResponse{}, err
	}
	return info, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseSubject(subject string) (int64, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return 0, errors.New("oauth userinfo subject is empty")
	}

	if id, err := strconv.ParseInt(subject, 10, 64); err == nil {
		return id, nil
	}

	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, subject)
	if digits == "" {
		return 0, fmt.Errorf("invalid oauth subject: %q", subject)
	}

	return strconv.ParseInt(digits, 10, 64)
}
