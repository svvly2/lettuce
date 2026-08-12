package clientutils

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/svvly2/lettuce/internal/retry"
	"github.com/svvly2/lettuce/internal/roblox"
)

func GetRequest(c *roblox.Client, url string) (*bytes.Buffer, error) {
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	if c != nil && strings.TrimSpace(c.AccessToken) != "" && req.URL.Host == "apis.roblox.com" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	body, err := retry.Do(
		retry.NewOptions(retry.Tries(3)),
		func(_ int) (*bytes.Buffer, error) {
			resp, err := c.DoRequest(req)
			if err != nil {
				return nil, &retry.ContinueRetry{Err: err}
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, &retry.ExitRetry{Err: errors.New(resp.Status)}
			}

			var buffer bytes.Buffer
			io.Copy(&buffer, resp.Body)
			return &buffer, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return body, nil
}
