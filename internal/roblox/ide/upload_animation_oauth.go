package ide

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/svvly2/lettuce/internal/roblox"
)

var UploadAnimationOAuthErrors = struct {
	ErrNotLoggedIn       error
	ErrTokenInvalid      error
	ErrInappropriateName error
}{
	ErrNotLoggedIn:       errors.New("not logged in"),
	ErrTokenInvalid:      errors.New("XSRF token validation failed"),
	ErrInappropriateName: errors.New("inappropriate name or description"),
}

var oauthErrTokenInvalid = errors.New("XSRF token validation failed")

func NewUploadAnimationOAuthHandler(
	c *roblox.OAuthClient,
	name,
	description string,
	data *bytes.Buffer,
	creatorID int64,
	isGroup bool,
) (func() (int64, error), error) {
	if c == nil {
		return nil, errors.New("oauth client is nil")
	}

	return func() (int64, error) {
		req, err := newCreateAssetRequest(
			"Animation",
			name,
			description,
			data,
			"model/x-rbxm",
			func() int64 {
				if isGroup {
					return creatorID
				}
				return c.UserID
			}(),
			isGroup,
		)
		if err != nil {
			return 0, err
		}

		id, err := executeCreateAssetOAuth(c, req, UploadAnimationOAuthErrors.ErrTokenInvalid, UploadAnimationOAuthErrors.ErrNotLoggedIn)
		if err == nil {
			return id, nil
		}

		if isInappropriateError(err.Error()) {
			return 0, UploadAnimationOAuthErrors.ErrInappropriateName
		}

		return 0, err
	}, nil
}

func executeCreateAssetOAuth(
	c *roblox.OAuthClient,
	req *http.Request,
	onTokenInvalid error,
	onNotLoggedIn error,
) (int64, error) {
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.DoRequest(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var operation operationResponse
		if err := json.Unmarshal(body, &operation); err != nil {
			return 0, err
		}
		if operation.Error != nil {
			return 0, errors.New(operation.Error.Message)
		}
		if operation.Done {
			return parseAssetID(&operation)
		}

		operationID := extractOperationID(operation.Path)
		if operationID == "" {
			return 0, errors.New("create asset operation id is empty")
		}

		var poll429Streak int
		for i := 0; i < maxPollAttempts; i++ {
			time.Sleep(pollInterval)
			polled, err := pollOperationOAuth(c, operationID)
			if err != nil {
				if errors.Is(err, onTokenInvalid) {
					return 0, onTokenInvalid
				}
				if errors.Is(err, ErrRateLimited) {
					poll429Streak++
					if poll429Streak > 40 {
						return 0, err
					}
					wait := 3 * time.Second
					var rle *RateLimitError
					if errors.As(err, &rle) && rle.RetryAfter > 0 {
						wait = rle.RetryAfter
					}
					wait += 1200 * time.Millisecond
					if poll429Streak >= 2 {
						wait += time.Duration(min(poll429Streak, 8)) * 350 * time.Millisecond
					}
					if wait > 45*time.Second {
						wait = 45 * time.Second
					}
					time.Sleep(wait)
					i--
					continue
				}
				return 0, err
			}
			poll429Streak = 0
			if !polled.Done {
				continue
			}
			if polled.Error != nil {
				return 0, errors.New(polled.Error.Message)
			}
			return parseAssetID(polled)
		}
		return 0, errors.New("asset operation timed out")
	case http.StatusUnauthorized:
		return 0, onNotLoggedIn
	case http.StatusTooManyRequests:
		return 0, newRateLimitError(resp.Header.Get("Retry-After"))
	case http.StatusForbidden:
		return 0, onTokenInvalid
	default:
		return 0, errors.New(decodeStatus(body, resp.Status))
	}
}

func pollOperationOAuth(c *roblox.OAuthClient, operationID string) (*operationResponse, error) {
	req, err := http.NewRequest("GET", operationBaseURL+operationID, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("User-Agent", "RobloxStudio/WinInet")

	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var operation operationResponse
		if err := json.Unmarshal(body, &operation); err != nil {
			return nil, err
		}
		return &operation, nil
	case http.StatusTooManyRequests:
		return nil, newRateLimitError(resp.Header.Get("Retry-After"))
	case http.StatusForbidden:
		return nil, oauthErrTokenInvalid
	default:
		return nil, errors.New(decodeStatus(body, resp.Status))
	}
}
