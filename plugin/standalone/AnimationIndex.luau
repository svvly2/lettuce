--!strict
local ScriptEditorService = game:GetService("ScriptEditorService")
local AssetIndex = {}

export type AssetType = "Animation" | "Sound"
export type Mapping = { oldId: number, newId: number }
type Reference = { kind: string, instance: Instance, property: string?, attributeName: string? }
export type Index = { ids: { number }, references: { [number]: { Reference } }, unreadableScripts: number }

local function valid(value: string): boolean
	return #value >= 6 and #value <= 16 and tonumber(value) ~= nil and tonumber(value) ~= 0
end

local function firstId(value: string): number?
	local best: string? = nil
	for digits in string.gmatch(value, "%d+") do
		if valid(digits) and (best == nil or #digits > #(best :: string)) then best = digits end
	end
	return if best then tonumber(best) else nil
end

local function hinted(value: string, assetType: AssetType): boolean
	local lower = string.lower(value)
	return if assetType == "Animation" then string.find(lower, "anim", 1, true) ~= nil
		else string.find(lower, "sound", 1, true) ~= nil or string.find(lower, "audio", 1, true) ~= nil
end

local function add(refs: { [number]: { Reference } }, id: number?, reference: Reference)
	if id == nil then return end
	refs[id] = refs[id] or {}
	table.insert(refs[id], reference)
end

local function scanSource(source: string, assetType: AssetType, callback: (number) -> ())
	local seen: { [number]: boolean } = {}
	local cursor = 1
	while cursor <= #source do
		local startAt, endAt, digits = string.find(source, "(%d%d%d%d%d%d+)", cursor)
		if not startAt or not endAt or not digits then break end
		cursor = endAt + 1
		local context = string.sub(source, math.max(1, startAt - 96), math.min(#source, endAt + 24))
		local id = tonumber(digits)
		if valid(digits) and id and not seen[id] and hinted(context, assetType) then
			seen[id] = true
			callback(id)
		end
	end
end

function AssetIndex.scan(root: Instance, assetType: AssetType): Index
	local refs: { [number]: { Reference } } = {}
	local unreadable = 0
	for _, instance in root:GetDescendants() do
		if assetType == "Animation" and instance:IsA("Animation") then
			add(refs, firstId(instance.AnimationId), { kind = "property", instance = instance, property = "AnimationId" })
		elseif assetType == "Sound" and instance:IsA("Sound") then
			add(refs, firstId(instance.SoundId), { kind = "property", instance = instance, property = "SoundId" })
		elseif assetType == "Sound" and instance:IsA("AudioPlayer") then
			add(refs, firstId(instance.AssetId), { kind = "property", instance = instance, property = "AssetId" })
		elseif instance:IsA("LuaSourceContainer") then
			local ok, source = pcall(function() return instance.Source end)
			if not ok then unreadable += 1 else
				scanSource(source :: string, assetType, function(id)
					add(refs, id, { kind = "script", instance = instance })
				end)
			end
		elseif instance:IsA("StringValue") and hinted(instance.Name, assetType) then
			add(refs, firstId(instance.Value), { kind = "property", instance = instance, property = "Value" })
		elseif (instance:IsA("IntValue") or instance:IsA("NumberValue")) and hinted(instance.Name, assetType) then
			if instance.Value % 1 == 0 then add(refs, instance.Value, { kind = "number", instance = instance, property = "Value" }) end
		end
		for name, value in instance:GetAttributes() do
			if hinted(name, assetType) then
				if type(value) == "string" then add(refs, firstId(value), { kind = "attribute", instance = instance, attributeName = name })
				elseif type(value) == "number" and value % 1 == 0 then add(refs, value, { kind = "attribute", instance = instance, attributeName = name }) end
			end
		end
	end
	local ids = {}
	for id in refs do table.insert(ids, id) end
	table.sort(ids)
	return { ids = ids, references = refs, unreadableScripts = unreadable }
end

local function replaceToken(value: string, oldId: number, newId: number): string
	return string.gsub(value, `%f[%d]{oldId}%f[%D]`, tostring(newId))
end

function AssetIndex.apply(index: Index, mappings: { Mapping }): (number, { string })
	local changed, errors = 0, {}
	local scripts: { [LuaSourceContainer]: { Mapping } } = {}
	for _, mapping in mappings do
		for _, ref in index.references[mapping.oldId] or {} do
			if ref.kind == "script" then
				local scriptRef = ref.instance :: LuaSourceContainer
				scripts[scriptRef] = scripts[scriptRef] or {}
				table.insert(scripts[scriptRef], mapping)
				continue
			end
			local ok, err = pcall(function()
				if ref.kind == "attribute" and ref.attributeName then
					local current = ref.instance:GetAttribute(ref.attributeName)
					ref.instance:SetAttribute(ref.attributeName, if type(current) == "string" then replaceToken(current, mapping.oldId, mapping.newId) else mapping.newId)
				elseif ref.kind == "number" then (ref.instance :: any)[ref.property :: string] = mapping.newId
				elseif ref.property then
					local current = tostring((ref.instance :: any)[ref.property])
					(ref.instance :: any)[ref.property] = replaceToken(current, mapping.oldId, mapping.newId)
				end
			end)
			if ok then changed += 1 else table.insert(errors, tostring(err)) end
		end
	end
	for sourceContainer, mappingsForScript in scripts do
		local ok, err = pcall(function()
			ScriptEditorService:UpdateSourceAsync(sourceContainer, function(source)
				for _, mapping in mappingsForScript do source = replaceToken(source, mapping.oldId, mapping.newId) end
				return source
			end)
		end)
		if ok then changed += 1 else table.insert(errors, tostring(err)) end
	end
	return changed, errors
end

return AssetIndex
