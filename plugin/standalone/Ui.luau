--!strict
local Ui = {}
Ui.__index = Ui

local C = { bg=Color3.fromRGB(9,9,9), card=Color3.fromRGB(20,20,20), text=Color3.fromRGB(247,247,247), muted=Color3.fromRGB(145,145,145), line=Color3.fromRGB(43,43,43), good=Color3.fromRGB(132,235,168), bad=Color3.fromRGB(255,116,116) }
local function font(weight) return Font.new("rbxasset://fonts/families/BuilderSans.json", weight, Enum.FontStyle.Normal) end
local function corner(parent, radius) local x=Instance.new("UICorner"); x.CornerRadius=UDim.new(0,radius); x.Parent=parent end
local function text(parent, value, size, color, weight)
	local x=Instance.new("TextLabel"); x.BackgroundTransparency=1; x.BorderSizePixel=0; x.FontFace=font(weight); x.Text=value; x.TextColor3=color; x.TextSize=size; x.TextXAlignment=Enum.TextXAlignment.Left; x.Parent=parent; return x
end
local function line(parent, pos, size, rotation)
	local x=Instance.new("Frame"); x.AnchorPoint=Vector2.new(.5,.5); x.BackgroundColor3=C.muted; x.BorderSizePixel=0; x.Position=pos; x.Size=size; x.Rotation=rotation or 0; x.Parent=parent; corner(x,2); return x
end
local function icon(parent, kind)
	local box=Instance.new("Frame"); box.BackgroundTransparency=1; box.Size=UDim2.fromOffset(22,22); box.Parent=parent
	if kind=="animation" then
		line(box,UDim2.fromScale(.5,.22),UDim2.fromOffset(15,3)); line(box,UDim2.fromScale(.5,.5),UDim2.fromOffset(15,3)); line(box,UDim2.fromScale(.5,.78),UDim2.fromOffset(15,3))
	else
		line(box,UDim2.fromScale(.22,.5),UDim2.fromOffset(3,8)); line(box,UDim2.fromScale(.43,.5),UDim2.fromOffset(3,15)); line(box,UDim2.fromScale(.64,.5),UDim2.fromOffset(3,11)); line(box,UDim2.fromScale(.85,.5),UDim2.fromOffset(3,6))
	end
	return box
end

function Ui.new(widget: DockWidgetPluginGui)
	local self=setmetatable({},Ui)
	local root=Instance.new("Frame"); root.BackgroundColor3=C.bg; root.BorderSizePixel=0; root.Size=UDim2.fromScale(1,1); root.Parent=widget; self.root=root
	local pad=Instance.new("UIPadding"); pad.PaddingTop=UDim.new(0,20); pad.PaddingBottom=UDim.new(0,16); pad.PaddingLeft=UDim.new(0,20); pad.PaddingRight=UDim.new(0,20); pad.Parent=root
	local title=text(root,"lettuce",27,C.text,Enum.FontWeight.Bold); title.Size=UDim2.new(1,-80,0,34)
	local version=text(root,"studio",12,C.muted,Enum.FontWeight.Medium); version.AnchorPoint=Vector2.new(1,0); version.Position=UDim2.new(1,0,0,8); version.Size=UDim2.fromOffset(70,18); version.TextXAlignment=Enum.TextXAlignment.Right
	local statusCard=Instance.new("Frame"); statusCard.BackgroundColor3=C.card; statusCard.BorderSizePixel=0; statusCard.Position=UDim2.new(0,0,0,48); statusCard.Size=UDim2.new(1,0,0,40); statusCard.Parent=root; corner(statusCard,15)
	local dot=Instance.new("Frame"); dot.AnchorPoint=Vector2.new(0,.5); dot.Position=UDim2.new(0,14,.5,0); dot.Size=UDim2.fromOffset(8,8); dot.BackgroundColor3=C.muted; dot.BorderSizePixel=0; dot.Parent=statusCard; corner(dot,8); self.statusDot=dot
	local status=text(statusCard,"checking lettuce…",13,C.muted,Enum.FontWeight.Medium); status.AnchorPoint=Vector2.new(0,.5); status.Position=UDim2.new(0,33,.5,0); status.Size=UDim2.new(1,-45,0,18); self.status=status
	local card=Instance.new("Frame"); card.BackgroundColor3=C.card; card.BorderSizePixel=0; card.Position=UDim2.new(0,0,0,100); card.Size=UDim2.new(1,0,1,-186); card.Parent=root; corner(card,20)
	local cp=Instance.new("UIPadding"); cp.PaddingTop=UDim.new(0,18); cp.PaddingBottom=UDim.new(0,18); cp.PaddingLeft=UDim.new(0,18); cp.PaddingRight=UDim.new(0,18); cp.Parent=card
	local kicker=text(card,"animations",12,C.muted,Enum.FontWeight.Medium); kicker.Size=UDim2.new(1,0,0,17); self.kicker=kicker
	local count=text(card,"scanning…",27,C.text,Enum.FontWeight.Bold); count.Position=UDim2.new(0,0,0,22); count.Size=UDim2.new(1,0,0,35); self.count=count
	local detail=text(card,"finding references",12,C.muted,Enum.FontWeight.Regular); detail.Position=UDim2.new(0,0,0,59); detail.Size=UDim2.new(1,0,0,18); self.detail=detail
	local primary=Instance.new("TextButton"); primary.AnchorPoint=Vector2.new(0,1); primary.AutoButtonColor=false; primary.BackgroundColor3=C.text; primary.BorderSizePixel=0; primary.FontFace=font(Enum.FontWeight.SemiBold); primary.Position=UDim2.new(0,0,1,0); primary.Size=UDim2.new(1,0,0,44); primary.Text="replace animations"; primary.TextColor3=C.bg; primary.TextSize=14; primary.Parent=card; corner(primary,15); self.primary=primary
	local nav=Instance.new("Frame"); nav.AnchorPoint=Vector2.new(.5,1); nav.Position=UDim2.new(.5,0,1,0); nav.Size=UDim2.new(0,210,0,58); nav.BackgroundColor3=C.card; nav.BorderSizePixel=0; nav.Parent=root; corner(nav,19); self.nav=nav
	local layout=Instance.new("UIListLayout"); layout.FillDirection=Enum.FillDirection.Horizontal; layout.HorizontalAlignment=Enum.HorizontalAlignment.Center; layout.VerticalAlignment=Enum.VerticalAlignment.Center; layout.Padding=UDim.new(0,6); layout.Parent=nav
	self.tabs={}
	for _, item in {{"Animation","animation"},{"Sound","sound"}} do
		local button=Instance.new("TextButton"); button.AutoButtonColor=false; button.BackgroundColor3=Color3.fromRGB(242,242,239); button.BackgroundTransparency=if item[1]=="Animation" then 0 else 1; button.BorderSizePixel=0; button.Size=UDim2.fromOffset(96,46); button.Text=""; button.Parent=nav; corner(button,14)
		local glyph=icon(button,item[2]); glyph.AnchorPoint=Vector2.new(.5,0); glyph.Position=UDim2.new(.5,0,0,4)
		local label=text(button,string.lower(item[1]),9,if item[1]=="Animation" then C.bg else C.muted,Enum.FontWeight.SemiBold); label.AnchorPoint=Vector2.new(.5,1); label.Position=UDim2.new(.5,0,1,-4); label.Size=UDim2.new(1,0,0,12); label.TextXAlignment=Enum.TextXAlignment.Center
		self.tabs[item[1]]={button=button,glyph=glyph,label=label}
	end
	local footer=text(root,"ready",11,C.muted,Enum.FontWeight.Medium); footer.AnchorPoint=Vector2.new(0,1); footer.Position=UDim2.new(0,0,1,-64); footer.Size=UDim2.new(1,0,0,16); footer.TextTruncate=Enum.TextTruncate.AtEnd; self.footer=footer
	local function responsive()
		local width=root.AbsoluteSize.X
		local compact=width<360
		pad.PaddingLeft=UDim.new(0,if compact then 14 else 20); pad.PaddingRight=UDim.new(0,if compact then 14 else 20)
		nav.Size=UDim2.new(0,math.min(210,math.max(180,width-40)),0,58)
		count.TextSize=if compact then 23 else 27
	end
	root:GetPropertyChangedSignal("AbsoluteSize"):Connect(responsive); task.defer(responsive)
	return self
end

function Ui:onRun(callback) self.primary.Activated:Connect(callback) end
function Ui:onMode(callback)
	for mode, tab in self.tabs do tab.button.Activated:Connect(function() callback(mode) end) end
end
function Ui:setMode(mode)
	self.mode=mode; self.kicker.Text=string.lower(mode).."s"
	for name,tab in self.tabs do
		local selected=name==mode; tab.button.BackgroundTransparency=if selected then 0 else 1; tab.label.TextColor3=if selected then C.bg else C.muted
		for _,part in tab.glyph:GetChildren() do if part:IsA("Frame") then part.BackgroundColor3=if selected then C.bg else C.muted end end
	end
end
function Ui:setConnection(ok,message) self.status.Text=message; self.status.TextColor3=if ok then C.text else C.muted; self.statusDot.BackgroundColor3=if ok then C.good else C.bad end
function Ui:setCount(count,unreadable)
	local noun=if self.mode=="Sound" then "sound" else "animation"; self.count.Text=`{count} {noun}{if count==1 then "" else "s"}`
	self.detail.Text=if count==0 then "nothing to replace yet" elseif unreadable and unreadable>0 then `ready · {unreadable} locked scripts skipped` else "ready to swap automatically"
	self.primary.Text=if count>0 then `replace {count}` else "scan again"
end
function Ui:setBusy(busy,message) self.primary.Active=not busy; self.primary.BackgroundColor3=if busy then C.line else C.text; self.primary.TextColor3=if busy then C.muted else C.bg; if message then self.primary.Text=message end end
function Ui:setFooter(message,isError) self.footer.Text=message; self.footer.TextColor3=if isError then C.bad else C.muted end
function Ui:setProgress(done,total) self.primary.Text=`replacing {done}/{total}`; self.footer.Text="Lettuce is uploading in the background" end
return Ui
