package assetutils

import (
	"github.com/svvly2/lettuce/internal/app/context"
	"github.com/svvly2/lettuce/internal/app/request"
	"github.com/svvly2/lettuce/internal/roblox/develop"
)

func NewFilter(ctx *context.Context, r *request.Request, assetTypeID int32) func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
	creatorID := r.CreatorID
	userID := ctx.Client.UserInfo.ID
	checkUserID := !r.IsGroup

	return func(assetsInfo develop.GetAssetsInfoResponse) []*develop.AssetInfo {
		filteredAssetsInfo := assetsInfo.Data[:0]
		for _, info := range assetsInfo.Data {
			if info.TypeID != assetTypeID {
				continue
			}

			assetCreatorID := info.Creator.TargetID
			// Skip only assets already owned by the destination. Roblox-owned
			// animations still need copying when Studio cannot use them directly.
			if assetCreatorID == creatorID || (checkUserID && assetCreatorID == userID) {
				continue
			}

			filteredAssetsInfo = append(filteredAssetsInfo, info)
		}
		return filteredAssetsInfo
	}
}
