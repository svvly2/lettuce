package assets

import (
	"errors"

	"github.com/svvly2/lettuce/internal/app/assets/animation"
	"github.com/svvly2/lettuce/internal/app/assets/mesh"
	"github.com/svvly2/lettuce/internal/app/assets/shared/clientutils"
	"github.com/svvly2/lettuce/internal/app/assets/shared/permissions"
	"github.com/svvly2/lettuce/internal/app/assets/sound"
	"github.com/svvly2/lettuce/internal/app/context"
	"github.com/svvly2/lettuce/internal/app/request"
	"github.com/svvly2/lettuce/internal/app/response"
	"github.com/svvly2/lettuce/internal/console"
	"github.com/svvly2/lettuce/internal/roblox"
)

var assetModules = map[string]func(ctx *context.Context, r *request.Request){
	"Animation": animation.Reupload,
	"Mesh":      mesh.Reupload,
	"Sound":     sound.Reupload,
}

func NewReuploadHandlerWithType(assetType string, c *roblox.Client, r *request.RawRequest, resp *response.Response) (func() error, error) {
	reupload, exists := assetModules[assetType]
	if !exists {
		return func() error { return nil }, errors.New(assetType + " module does not exist")
	}

	return func() error {
		ctx := context.New(c, resp)

		console.ClearScreen()

		ctx.Logger.Info("Getting current place details...")
		req, err := request.FromRawRequest(c, r)
		console.ClearScreen()
		if err != nil {
			return err
		}

		if !c.IsOAuthOnly() {
			ctx.Logger.Info("Checking if account can edit universe...")
			err = permissions.CanEditUniverse(ctx, req)
			console.ClearScreen()
			if err != nil {
				clientutils.GetNewCookie(ctx, req, err.Error())
			}
		}

		reupload(ctx, req)
		return nil
	}, nil
}

func DoesModuleExist(m string) bool {
	_, exists := assetModules[m]
	return exists
}
