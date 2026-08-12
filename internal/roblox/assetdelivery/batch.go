package assetdelivery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/svvly2/lettuce/internal/roblox"
)

var ErrBodyTooLarge = errors.New("batch request body is too large")

func NewOpenCloudAssetURL(assetID int64) string {
	return fmt.Sprintf("https://apis.roblox.com/asset-delivery-api/v1/assetId/%d", assetID)
}

func GetPublicAssetLocation(assetID int64) (string, error) {
	url := fmt.Sprintf("https://assetdelivery.roblox.com/v2/assetId/%d", assetID)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "RobloxStudio/WinInet")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("public asset delivery: %s", resp.Status)
	}
	var delivery AssetLocation
	if err := json.NewDecoder(resp.Body).Decode(&delivery); err != nil {
		return "", err
	}
	if len(delivery.Locations) == 0 {
		message := "no public download location"
		if len(delivery.Errors) > 0 && strings.TrimSpace(delivery.Errors[0].Message) != "" {
			message = delivery.Errors[0].Message
		}
		return "", errors.New(message)
	}
	return delivery.Locations[0].Location, nil
}

type AssetRequestItem struct {
	AssetName                             string `json:"assetName"`
	AssetType                             string `json:"assetType"`
	ClientInsert                          bool   `json:"clientInsert"`
	PlaceID                               int64  `json:"placeId"`
	RequestID                             string `json:"requestId"`
	ScriptInsert                          bool   `json:"scriptInsert"`
	ServerPlaceID                         int64  `json:"serverPlaceId,omitempty"` // omit so you don't get "403 Asset is not trusted for this place"
	UniverseID                            int64  `json:"universeId"`
	Accept                                string `json:"accept"`
	Encoding                              string `json:"encoding"`
	Hash                                  string `json:"hash"`
	UserAssetID                           int64  `json:"userAssetId"`
	AssetID                               int64  `json:"assetId"`
	Version                               int32  `json:"version"`
	AssetVersionID                        int64  `json:"assetVersionId"`
	ModulePlaceID                         int64  `json:"modulePlaceId"`
	AssetFormat                           string `json:"assetFormat"`
	RobloxAssetFormat                     string `json:"roblox-assetFormat"`
	ContentRepresentationPriorityList     string `json:"contentRepresentationPriorityList"`
	DoNotFallbackToBaselineRepresentation bool   `json:"doNotFallbackToBaselineRepresentation"`
}

type AssetLocation struct {
	Locations []struct {
		AssetFormat    string `json:"assetFormat"`
		Location       string `json:"location"`
		AssetMetadatas []struct {
			MetadataType int32  `json:"metadataType"`
			Value        string `json:"value"`
		} `json:"assetMetadatas"`
	} `json:"locations,omitempty"`
	Errors []struct {
		Code            int32  `json:"Code"`
		Message         string `json:"Message"`
		CustomErrorCode int32  `json:"CustomErrorCode"`
	} `json:"errors,omitempty"`
	RequestID                      string `json:"requestId"`
	IsHashDynamic                  bool   `json:"IsHashDynamic"`
	IsCopyrightProtected           bool   `json:"IsCopyrightProtected"`
	IsArchived                     bool   `json:"isArchived"`
	AssetTypeID                    int32  `json:"assetTypeId"`
	ContentRepresentationSpecifier struct {
		Format       string `json:"format"`
		MajorVersion string `json:"majorVersion"`
		Fidelity     string `json:"fidelity"`
	} `json:"contentRepresentationSpecifier"`
}

func newBatchRequest(body []*AssetRequestItem, placeID int64) (*http.Request, error) {
	if len(body) > 50 {
		return nil, ErrBodyTooLarge
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://assetdelivery.roblox.com/v2/assets/batch", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RobloxStudio/WinInet")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Roblox-Place-Id", strconv.FormatInt(placeID, 10))

	return req, nil
}

func NewBatchHandler(c *roblox.Client, body []*AssetRequestItem, placeID ...int64) (func() ([]*AssetLocation, error), error) {
	var placeIDValue int64
	if len(placeID) > 0 {
		placeIDValue = placeID[0]
	}

	req, err := newBatchRequest(body, placeIDValue)
	if err != nil {
		return func() ([]*AssetLocation, error) { return nil, nil }, err
	}

	return func() ([]*AssetLocation, error) {
		req.AddCookie(&http.Cookie{
			Name:  ".ROBLOSECURITY",
			Value: c.Cookie,
		})

		resp, err := c.DoRequest(req)
		if err != nil {
			return make([]*AssetLocation, 0), err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status)
		}

		var locations []*AssetLocation
		json.NewDecoder(resp.Body).Decode(&locations)
		return locations, nil
	}, nil
}
