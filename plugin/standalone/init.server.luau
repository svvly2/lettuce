--!strict

local RunService = game:GetService("RunService")

if not RunService:IsEdit() then
	return
end

-- Rojo packages these modules as children of the entry Script. Nothing here
-- reaches into another plugin or downloads executable code.
local AnimationIndex = require(script.AnimationIndex)
local Bridge = require(script.Bridge)
local Config = require(script.Config)
local Ui = require(script.Ui)

local toolbar = plugin:CreateToolbar("Lettuce")
local toggle = toolbar:CreateButton("Lettuce", "open lettuce", "rbxassetid://83178708216253")
local widget = plugin:CreateDockWidgetPluginGui(
	"LettuceStandaloneV2",
	DockWidgetPluginGuiInfo.new(Enum.InitialDockState.Float, false, true, 380, 390, 330, 330)
)
widget.Title = "lettuce"
widget.ZIndexBehavior = Enum.ZIndexBehavior.Sibling

local ui = Ui.new(widget)
local assetType: AnimationIndex.AssetType = "Animation"
local currentIndex = AnimationIndex.scan(game, assetType)
local currentSession: any = nil
local busy = false

local function rescan()
	currentIndex = AnimationIndex.scan(game, assetType)
	ui:setCount(#currentIndex.ids, currentIndex.unreadableScripts)
	ui:setFooter("ready")
end

local function refreshConnection()
	local connected, message = Bridge.ping()
	ui:setConnection(connected, if connected then "lettuce is connected" else (message or "lettuce is offline"))
end

local function placeIdentity(): (number, boolean, number, string?)
	local placeId = game.PlaceId
	local creatorId = game.CreatorId
	if placeId == 0 then
		return 0, false, 0, "publish this place first"
	elseif creatorId == 0 then
		return 0, false, placeId, "couldn't read the place owner"
	end
	return creatorId, game.CreatorType == Enum.CreatorType.Group, placeId, nil
end

local function run()
	if busy then
		return
	end

	rescan()
	if #currentIndex.ids == 0 then
		refreshConnection()
		return
	end

	local creatorId, isGroup, placeId, identityError = placeIdentity()
	if identityError ~= nil then
		ui:setFooter(identityError, true)
		return
	end

	local connected, connectionError = Bridge.ping()
	if not connected then
		ui:setConnection(false, connectionError or "lettuce is offline")
		ui:setFooter(connectionError or "open Lettuce first", true)
		return
	end

	busy = true
	ui:setConnection(true, "lettuce is connected")
	ui:setBusy(true, "sending…")
	ui:setFooter("starting the OAuth upload")

	local completed = 0
	local total = #currentIndex.ids
	local session, startError = Bridge.start({
		pluginVersion = Config.VERSION,
		creatorId = creatorId,
		isGroup = isGroup,
		assetType = assetType,
		ids = currentIndex.ids,
		placeId = placeId,
		defaultPlaceIds = {},
		exportJSON = false,
	}, {
		onBatch = function(mappings)
			local _, replacementErrors = AnimationIndex.apply(currentIndex, mappings)
			completed += #mappings
			ui:setProgress(completed, total)
			if #replacementErrors > 0 then
				ui:setFooter(`{#replacementErrors} references couldn't be changed`, true)
			end
		end,
		onDone = function()
			busy = false
			currentSession = nil
			ui:setBusy(false)
			rescan()
			ui:setFooter(if completed > 0 then `done · {completed} replaced` else "done · no uploads succeeded", completed == 0)
		end,
		onError = function(message)
			busy = false
			currentSession = nil
			ui:setBusy(false, "try again")
			ui:setConnection(false, message)
			ui:setFooter(message, true)
		end,
	})

	if session == nil then
		busy = false
		ui:setBusy(false, "try again")
		ui:setFooter(startError or "couldn't start", true)
		return
	end
	currentSession = session
end

ui:onRun(run)
ui:onMode(function(nextMode: AnimationIndex.AssetType)
	if busy or nextMode == assetType then return end
	assetType = nextMode
	ui:setMode(assetType)
	rescan()
end)
ui:setMode(assetType)

toggle.Click:Connect(function()
	widget.Enabled = not widget.Enabled
end)

widget:GetPropertyChangedSignal("Enabled"):Connect(function()
	toggle:SetActive(widget.Enabled)
	if widget.Enabled and not busy then
		rescan()
		refreshConnection()
	end
end)

plugin.Unloading:Connect(function()
	if currentSession ~= nil then
		currentSession:Cancel()
	end
end)

rescan()
task.spawn(refreshConnection)
