--!strict

local HttpService = game:GetService("HttpService")
local Config = require(script.Parent.Config)

local Bridge = {}

export type Mapping = {
	oldId: number,
	newId: number,
}

export type UploadRequest = {
	pluginVersion: string,
	creatorId: number,
	isGroup: boolean,
	assetType: string,
	ids: { number },
	placeId: number,
	defaultPlaceIds: { number },
	exportJSON: boolean,
}

export type PollCallbacks = {
	onBatch: ({ Mapping }) -> (),
	onDone: () -> (),
	onError: (string) -> (),
}

type Session = {
	cancelled: boolean,
	Cancel: (self: Session) -> (),
}

local BASE_URL = `http://{Config.HOST}:{Config.PORT}`

local function request(method: string, path: string, body: string?): (boolean, string, number)
	local options = {
		Url = BASE_URL .. path,
		Method = method,
		Headers = {
			["Accept"] = "application/json",
		},
	}

	if body ~= nil then
		options.Headers["Content-Type"] = "application/json"
		options.Body = body
	end

	local called, response = pcall(function()
		return HttpService:RequestAsync(options)
	end)
	if not called then
		return false, tostring(response), 0
	end

	local result = response :: any
	if not result.Success then
		local message = result.Body
		if message == nil or message == "" then
			message = `HTTP {result.StatusCode} {result.StatusMessage}`
		end
		return false, tostring(message), result.StatusCode
	end

	return true, tostring(result.Body or ""), result.StatusCode
end

local function connectionHelp(detail: string): string
	local lower = string.lower(detail)
	if string.find(lower, "http requests are not enabled", 1, true) then
		return "turn on Allow HTTP Requests in Game Settings → Security"
	end
	return "open Lettuce, sign in, and leave it running"
end

function Bridge.ping(): (boolean, string?)
	local ok, detail = request("GET", "/", nil)
	if not ok then
		return false, connectionHelp(detail)
	end
	return true, nil
end

function Bridge.start(payload: UploadRequest, callbacks: PollCallbacks): (Session?, string?)
	local encodedOk, body = pcall(HttpService.JSONEncode, HttpService, payload)
	if not encodedOk then
		return nil, "couldn't prepare the animation list"
	end

	local sent, detail, status = request("POST", "/reupload", body :: string)
	if not sent then
		if status == 401 then
			return nil, "sign in to Roblox in Lettuce first"
		elseif status == 503 then
			return nil, "Lettuce is already handling another swap"
		elseif status == 422 then
			return nil, "this Lettuce build can't upload animations with OAuth"
		end
		return nil, connectionHelp(detail)
	end

	local session: Session = {
		cancelled = false,
		Cancel = function(self: Session)
			self.cancelled = true
		end,
	}

	task.spawn(function()
		local failures = 0
		while not session.cancelled do
			task.wait(Config.POLL_INTERVAL_SECONDS)
			local ok, responseBody = request("GET", "/", nil)
			if not ok then
				failures += 1
				if failures >= Config.MAX_CONSECUTIVE_POLL_FAILURES then
					session.cancelled = true
					callbacks.onError("lost connection to Lettuce")
				end
				continue
			end

			failures = 0
			local trimmed = string.gsub(responseBody, "^%s*(.-)%s*$", "%1")
			if trimmed == "" or trimmed == "null" or trimmed == "[]" then
				continue
			elseif trimmed == "done" then
				session.cancelled = true
				callbacks.onDone()
				break
			end

			local decoded, value = pcall(HttpService.JSONDecode, HttpService, trimmed)
			if decoded and type(value) == "table" then
				local mappings = {} :: { Mapping }
				for _, item in value :: any do
					if type(item) == "table" and type(item.oldId) == "number" and type(item.newId) == "number" then
						table.insert(mappings, {
							oldId = item.oldId,
							newId = item.newId,
						})
					end
				end
				if #mappings > 0 then
					callbacks.onBatch(mappings)
				end
			end
		end
	end)

	return session, nil
end

return Bridge
