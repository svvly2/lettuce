package develop

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/svvly2/lettuce/internal/roblox"
)

type AssetInfo struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	TypeID      int32  `json:"typeId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Creator     struct {
		Type     string `json:"type"`
		TypeID   int32  `json:"typeId"`
		TargetID int64  `json:"targetId"`
	} `json:"creator"`
	Genres                []string  `json:"genres"`
	Created               time.Time `json:"created"`
	Updated               time.Time `json:"updated"`
	EnableComments        bool      `json:"enableComments"`
	IsCopyingAllowed      bool      `json:"isCopyingAllowed"`
	IsPublicDomainEnabled bool      `json:"isPublicDomainEnabled"`
	IsModerated           bool      `json:"isModerated"`
	ReviewStatus          string    `json:"reviewStatus"`
	IsVersioningEnabled   bool      `json:"isVersioningEnabled"`
	IsArchivable          bool      `json:"isArchivable"`
	CanHaveThumbnail      bool      `json:"canHaveThumbnail"`
}

var GetAssetsInfoErrors = struct {
	ErrUnauthorized error
}{
	ErrUnauthorized: errors.New("unauthorized"),
}

type GetAssetsInfoResponse struct {
	Data   []*AssetInfo `json:"data,omitempty"`
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type int64String int64

func (n *int64String) UnmarshalJSON(data []byte) error {
	var asNumber int64
	if err := json.Unmarshal(data, &asNumber); err == nil {
		*n = int64String(asNumber)
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err != nil {
		return err
	}
	parsed, err := strconv.ParseInt(asString, 10, 64)
	if err != nil {
		return err
	}
	*n = int64String(parsed)
	return nil
}

type openCloudAsset struct {
	AssetID         int64String `json:"assetId"`
	AssetType       string      `json:"assetType"`
	DisplayName     string      `json:"displayName"`
	Description     string      `json:"description"`
	CreationContext struct {
		Creator struct {
			UserID  int64String `json:"userId"`
			GroupID int64String `json:"groupId"`
		} `json:"creator"`
	} `json:"creationContext"`
}

func newAssetInfoBulkURL(assetIDs []int64) string {
	strIDs := make([]string, len(assetIDs))
	for i, id := range assetIDs {
		strIDs[i] = strconv.FormatInt(id, 10)
	}

	return fmt.Sprintf("https://develop.roblox.com/v1/assets?assetIds=%s", strings.Join(strIDs, ","))
}

func newAssetsInfoRequest(assetIDs []int64) (*http.Request, error) {
	url := newAssetInfoBulkURL(assetIDs)
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func NewAssetsInfoHandler(c *roblox.Client, assetIDs []int64) (func() (GetAssetsInfoResponse, error), error) {
	if c != nil && c.IsOAuthOnly() {
		return NewAssetsInfoOAuthHandler(c, assetIDs)
	}

	req, err := newAssetsInfoRequest(assetIDs)
	if err != nil {
		return func() (GetAssetsInfoResponse, error) { return GetAssetsInfoResponse{}, nil }, err
	}

	return func() (GetAssetsInfoResponse, error) {
		req.AddCookie(&http.Cookie{
			Name:  ".ROBLOSECURITY",
			Value: c.Cookie,
		})

		resp, err := c.DoRequest(req)
		if err != nil {
			return GetAssetsInfoResponse{}, err
		}
		defer resp.Body.Close()

		var bulkResponse GetAssetsInfoResponse
		json.NewDecoder(resp.Body).Decode(&bulkResponse)

		switch resp.StatusCode {
		case http.StatusOK:
			return bulkResponse, nil
		case http.StatusUnauthorized:
			return bulkResponse, GetAssetsInfoErrors.ErrUnauthorized
		default:
			if bulkResponse.Errors != nil {
				if message := bulkResponse.Errors[0].Message; message != "" {
					return bulkResponse, errors.New(bulkResponse.Errors[0].Message)
				}
			}

			return bulkResponse, errors.New(resp.Status)
		}
	}, nil
}

func NewAssetsInfoOAuthHandler(c *roblox.Client, assetIDs []int64) (func() (GetAssetsInfoResponse, error), error) {
	if c == nil {
		return nil, errors.New("roblox client is nil")
	}

	return func() (GetAssetsInfoResponse, error) {
		out := GetAssetsInfoResponse{Data: make([]*AssetInfo, 0, len(assetIDs))}
		for _, assetID := range assetIDs {
			info, err := getOpenCloudAssetInfo(c, assetID)
			if err != nil {
				return out, err
			}
			out.Data = append(out.Data, info)
		}
		return out, nil
	}, nil
}

func getOpenCloudAssetInfo(c *roblox.Client, assetID int64) (*AssetInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://apis.roblox.com/assets/v1/assets/%d", assetID), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

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
		var asset openCloudAsset
		if err := json.Unmarshal(body, &asset); err != nil {
			return nil, err
		}
		return openCloudAssetToDevelopInfo(asset, assetID), nil
	case http.StatusUnauthorized:
		return nil, GetAssetsInfoErrors.ErrUnauthorized
	default:
		return nil, errors.New(decodeOpenCloudStatus(body, resp.Status))
	}
}

func openCloudAssetToDevelopInfo(asset openCloudAsset, fallbackID int64) *AssetInfo {
	id := int64(asset.AssetID)
	if id == 0 {
		id = fallbackID
	}

	info := &AssetInfo{
		ID:          id,
		Type:        asset.AssetType,
		TypeID:      openCloudAssetTypeID(asset.AssetType),
		Name:        asset.DisplayName,
		Description: asset.Description,
	}
	if info.Name == "" {
		info.Name = fmt.Sprintf("Asset %d", id)
	}

	if groupID := int64(asset.CreationContext.Creator.GroupID); groupID > 0 {
		info.Creator.Type = "Group"
		info.Creator.TargetID = groupID
		return info
	}

	info.Creator.Type = "User"
	info.Creator.TargetID = int64(asset.CreationContext.Creator.UserID)
	return info
}

func openCloudAssetTypeID(assetType string) int32 {
	switch strings.TrimPrefix(strings.ToUpper(assetType), "ASSET_TYPE_") {
	case "ANIMATION":
		return 24
	case "MESH":
		return 4
	case "AUDIO":
		return 3
	default:
		return 0
	}
}

func decodeOpenCloudStatus(body []byte, fallback string) string {
	var status struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &status); err == nil && status.Message != "" {
		return status.Message
	}
	return fallback
}
