package ide

import (
	"bytes"
	"errors"

	"github.com/svvly2/lettuce/internal/roblox"
)

var UploadMeshOAuthErrors = struct {
	ErrNotLoggedIn       error
	ErrTokenInvalid      error
	ErrInappropriateName error
}{
	ErrNotLoggedIn:       errors.New("not logged in"),
	ErrTokenInvalid:      errors.New("XSRF token validation failed"),
	ErrInappropriateName: errors.New("inappropriate name or description"),
}

func NewUploadMeshOAuthHandler(
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

	currentName := name
	return func() (int64, error) {
		req, err := newCreateAssetRequest(
			"Mesh",
			currentName,
			description,
			data,
			"model/x-file-mesh-data",
			func() int64 {
				if isGroup {
					return creatorID
				}
				if c.UserID > 0 {
					return c.UserID
				}
				return creatorID
			}(),
			isGroup,
		)
		if err != nil {
			return 0, err
		}

		id, err := executeCreateAssetOAuth(c, req, UploadMeshOAuthErrors.ErrTokenInvalid, UploadMeshOAuthErrors.ErrNotLoggedIn)
		if err == nil {
			return id, nil
		}

		if isInappropriateError(err.Error()) {
			currentName = "[Censored]"
			return 0, UploadMeshOAuthErrors.ErrInappropriateName
		}

		return 0, err
	}, nil
}
