package request

import (
	"strings"

	"github.com/svvly2/lettuce/internal/roblox"
	"github.com/svvly2/lettuce/internal/roblox/games"
)

type RawRequest struct {
	PlaceID         int64   `json:"placeId"`
	CreatorID       int64   `json:"creatorId"`
	IDs             []int64 `json:"ids"`
	DefaultPlaceIDs []int64 `json:"defaultPlaceIds"`
	PluginVersion   string  `json:"pluginVersion"`
	AssetType       string  `json:"assetType"`
	ExportJSON      bool    `json:"exportJSON"`
	IsGroup         bool    `json:"isGroup"`
}

type Request struct {
	UniverseID      int64
	PlaceID         int64
	CreatorID       int64
	IDs             []int64
	DefaultPlaceIDs []int64
	AssetType       string
	IsGroup         bool
}

func FromRawRequest(c *roblox.Client, req *RawRequest) (*Request, error) {
	placeID := req.PlaceID

	placesInfo, err := games.MultiGetPlaceDetails(c, []int64{placeID})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unauthorized") || (c != nil && c.IsOAuthOnly()) {
			return fromRawWithoutPlaceDetails(req), nil
		}
		return nil, err
	}
	if len(placesInfo) == 0 {
		return fromRawWithoutPlaceDetails(req), nil
	}

	return &Request{
		UniverseID:      placesInfo[0].UniverseID,
		PlaceID:         placeID,
		CreatorID:       req.CreatorID,
		IDs:             req.IDs,
		DefaultPlaceIDs: req.DefaultPlaceIDs,
		AssetType:       req.AssetType,
		IsGroup:         req.IsGroup,
	}, nil
}

func fromRawWithoutPlaceDetails(req *RawRequest) *Request {
	return &Request{
		UniverseID:      0,
		PlaceID:         req.PlaceID,
		CreatorID:       req.CreatorID,
		IDs:             req.IDs,
		DefaultPlaceIDs: req.DefaultPlaceIDs,
		AssetType:       req.AssetType,
		IsGroup:         req.IsGroup,
	}
}
