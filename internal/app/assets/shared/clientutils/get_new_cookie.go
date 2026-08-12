package clientutils

import (
	"errors"
	"strings"

	"github.com/svvly2/lettuce/internal/app/assets/shared/permissions"
	"github.com/svvly2/lettuce/internal/app/config"
	"github.com/svvly2/lettuce/internal/app/context"
	"github.com/svvly2/lettuce/internal/app/request"
	"github.com/svvly2/lettuce/internal/console"
	"github.com/svvly2/lettuce/internal/files"
)

var cookieFile = config.Get("cookie_file")

func GetNewCookie(ctx *context.Context, r *request.Request, m string) {
	pauseController := ctx.PauseController

	if !pauseController.Pause() {
		pauseController.WaitIfPaused()
		return
	}

	console.ClearScreen()

	client := ctx.Client
	if client != nil && strings.TrimSpace(client.Cookie) == "" && strings.TrimSpace(client.AccessToken) != "" {
		ctx.Logger.Error(m + ": OAuth is connected, but this Roblox endpoint rejected the request. Skipping instead of waiting for a cookie.")
		pauseController.Unpause()
		return
	}

	inputErr := errors.New(m)
	for {
		ctx.Logger.Error(inputErr)

		i, err := console.LongInput("ROBLOSECURITY: ")
		console.ClearScreen()
		if err != nil {
			inputErr = err
			continue
		}

		ctx.Logger.Info("Authenticating cookie...")
		err = client.SetCookie(i)
		console.ClearScreen()
		if err != nil {
			inputErr = err
			continue
		}

		ctx.Logger.Info("Checking if account can edit universe...")
		err = permissions.CanEditUniverse(ctx, r)
		console.ClearScreen()
		if err != nil {
			inputErr = err
			continue
		}

		break
	}

	if err := files.Write(cookieFile, client.Cookie); err != nil {
		ctx.Logger.Error("Failed to save cookie: ", err)
	}

	pauseController.Unpause()
}
