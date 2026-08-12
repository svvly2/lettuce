package ide

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/svvly2/lettuce/internal/roblox"
)

func NewUploadAudioOAuthHandler(c *roblox.OAuthClient, name, description string, data *bytes.Buffer, creatorID int64, isGroup bool) (func() (int64, error), error) {
	if c == nil {
		return nil, errors.New("oauth client is nil")
	}
	contentType := http.DetectContentType(data.Bytes())
	switch contentType {
	case "audio/mpeg", "audio/ogg", "audio/wav", "audio/x-wav", "audio/flac", "audio/x-flac":
	default:
		return nil, errors.New("asset delivery did not return a supported audio file")
	}
	if contentType == "audio/x-wav" {
		contentType = "audio/wav"
	}
	if contentType == "audio/x-flac" {
		contentType = "audio/flac"
	}

	return func() (int64, error) {
		destination := c.UserID
		if isGroup {
			destination = creatorID
		}
		req, err := newCreateAssetRequest("Audio", name, description, data, contentType, destination, isGroup)
		if err != nil {
			return 0, err
		}
		return executeCreateAssetOAuth(c, req, UploadAnimationOAuthErrors.ErrTokenInvalid, UploadAnimationOAuthErrors.ErrNotLoggedIn)
	}, nil
}
