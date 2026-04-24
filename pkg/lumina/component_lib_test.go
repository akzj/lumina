package lumina

import (
	"strings"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func clRun(t *testing.T, code string) *lua.State {
	t.Helper()
	L := lua.NewState()
	Open(L)
	if err := L.DoString(code); err != nil {
		t.Fatalf("DoString failed: %v", err)
	}
	return L
}

func clStr(L *lua.State, field string) string {
	L.GetField(-1, field)
	s, _ := L.ToString(-1)
	L.Pop(1)
	return s
}

func clBool(L *lua.State, field string) bool {
	L.GetField(-1, field)
	b := L.ToBoolean(-1)
	L.Pop(1)
	return b
}

func clChildCount(L *lua.State) int {
	L.GetField(-1, "children")
	n := int(L.RawLen(-1))
	L.Pop(1)
	return n
}

func clPushChild(L *lua.State, i int) {
	L.GetField(-1, "children")
	L.RawGetI(-1, int64(i))
	L.Remove(-2)
}

func clAssertFactory(t *testing.T, L *lua.State, globalName, compName string) {
	t.Helper()
	L.GetGlobal(globalName)
	if L.Type(-1) != lua.TypeTable {
		t.Fatalf("%s: expected table, got %s", compName, L.TypeName(L.Type(-1)))
	}
	if n := clStr(L, "name"); n != compName {
		t.Errorf("%s: name = %q, want %q", compName, n, compName)
	}
	if !clBool(L, "isComponent") {
		t.Errorf("%s: isComponent should be true", compName)
	}
	L.Pop(1)
}

func TestListComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		List = lumina.defineComponent({
			name = "List",
			init = function(props) return { items = props.items or {}, selectedIndex = props.selectedIndex or 1 } end,
			render = function(instance)
				local items = instance.items or {}
				local selected = instance.selectedIndex or 1
				local children = {}
				for i, item in ipairs(items) do
					local isSel = (i == selected)
					children[#children + 1] = {
						type = "text",
						content = (isSel and "▸ " or "  ") .. tostring(item),
						foreground = isSel and "#00FF00" or "#FFFFFF",
					}
				end
				return { type = "vbox", children = children }
			end
		})
		result = List.render({ items = {"Apple", "Banana", "Cherry"}, selectedIndex = 2 })
	`)
	defer L.Close()
	clAssertFactory(t, L, "List", "List")
	L.GetGlobal("result")
	defer L.Pop(1)
	if clStr(L, "type") != "vbox" { t.Errorf("type = %q", clStr(L, "type")) }
	if n := clChildCount(L); n != 3 { t.Errorf("children = %d, want 3", n) }
	clPushChild(L, 2)
	if c := clStr(L, "content"); !strings.HasPrefix(c, "▸") { t.Errorf("selected = %q", c) }
	if clStr(L, "foreground") != "#00FF00" { t.Errorf("fg = %q", clStr(L, "foreground")) }
	L.Pop(1)
}

func TestListEmpty(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		List = lumina.defineComponent({
			name = "List",
			init = function(props) return { items = props.items or {} } end,
			render = function(instance)
				local ch = {}
				for _, item in ipairs(instance.items) do ch[#ch+1] = { type="text", content=item } end
				if #ch == 0 then ch[#ch+1] = { type="text", content="(empty)" } end
				return { type = "vbox", children = ch }
			end
		})
		result = List.render({ items = {} })
	`)
	defer L.Close()
	L.GetGlobal("result")
	defer L.Pop(1)
	clPushChild(L, 1)
	if clStr(L, "content") != "(empty)" { t.Errorf("empty list = %q", clStr(L, "content")) }
	L.Pop(1)
}

func TestTabsComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		Tabs = lumina.defineComponent({
			name = "Tabs",
			init = function(props) return { tabs = props.tabs or {}, activeTab = props.activeTab or 1 } end,
			render = function(instance)
				local tabs = instance.tabs or {}
				local active = instance.activeTab or 1
				local children = {}
				for i, tab in ipairs(tabs) do
					local isActive = (i == active)
					children[#children + 1] = {
						type = "text",
						content = isActive and ("[ " .. tab .. " ]") or ("  " .. tab .. "  "),
						foreground = isActive and "#00FFFF" or "#888888",
						bold = isActive,
					}
				end
				return { type = "hbox", children = children }
			end
		})
		result = Tabs.render({ tabs = {"General", "Advanced", "About"}, activeTab = 2 })
	`)
	defer L.Close()
	clAssertFactory(t, L, "Tabs", "Tabs")
	L.GetGlobal("result")
	defer L.Pop(1)
	if clStr(L, "type") != "hbox" { t.Errorf("type = %q", clStr(L, "type")) }
	if n := clChildCount(L); n != 3 { t.Errorf("children = %d", n) }
	clPushChild(L, 2)
	if c := clStr(L, "content"); !strings.Contains(c, "[ Advanced ]") { t.Errorf("active = %q", c) }
	if clStr(L, "foreground") != "#00FFFF" { t.Errorf("fg = %q", clStr(L, "foreground")) }
	L.Pop(1)
	clPushChild(L, 1)
	if clStr(L, "foreground") != "#888888" { t.Errorf("inactive fg = %q", clStr(L, "foreground")) }
	L.Pop(1)
}

func TestProgressComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		Progress = lumina.defineComponent({
			name = "Progress",
			init = function(props) return { value = props.value or 0, width = props.width or 20, showPercent = true } end,
			render = function(instance)
				local v = instance.value or 0
				if v < 0 then v = 0 end
				if v > 1 then v = 1 end
				local w = instance.width or 20
				local filled = math.floor(v * w + 0.5)
				local bar = string.rep("█", filled) .. string.rep("░", w - filled)
				return { type = "text", content = "[" .. bar .. "] " .. math.floor(v * 100 + 0.5) .. "%" }
			end
		})
	`)
	defer L.Close()
	clAssertFactory(t, L, "Progress", "Progress")

	for _, tt := range []struct{ name, code string; check func(string) bool }{
		{"0%", `result = Progress.render({ value = 0, width = 10 })`, func(s string) bool { return strings.Contains(s, "0%") && strings.Contains(s, "░░░░░░░░░░") }},
		{"100%", `result = Progress.render({ value = 1, width = 10 })`, func(s string) bool { return strings.Contains(s, "100%") && strings.Contains(s, "██████████") }},
		{"50%", `result = Progress.render({ value = 0.5, width = 10 })`, func(s string) bool { return strings.Contains(s, "50%") }},
		{"clamp-neg", `result = Progress.render({ value = -0.5, width = 10 })`, func(s string) bool { return strings.Contains(s, "0%") }},
		{"clamp-over", `result = Progress.render({ value = 1.5, width = 10 })`, func(s string) bool { return strings.Contains(s, "100%") }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.code); err != nil { t.Fatal(err) }
			L.GetGlobal("result")
			c := clStr(L, "content")
			L.Pop(1)
			if !tt.check(c) { t.Errorf("content = %q", c) }
		})
	}
}

func TestSpinnerComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		local frames = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		Spinner = lumina.defineComponent({
			name = "Spinner",
			init = function(props) return { frame = props.frame or 1, label = props.label or "" } end,
			render = function(instance)
				local f = ((instance.frame or 1) - 1) % #frames + 1
				local c = frames[f]
				if instance.label ~= "" then c = c .. " " .. instance.label end
				return { type = "text", content = c }
			end
		})
	`)
	defer L.Close()
	clAssertFactory(t, L, "Spinner", "Spinner")

	L.DoString(`result = Spinner.render({ frame = 1, label = "Loading..." })`)
	L.GetGlobal("result")
	c := clStr(L, "content")
	L.Pop(1)
	if !strings.Contains(c, "⠋") || !strings.Contains(c, "Loading...") { t.Errorf("frame1 = %q", c) }

	L.DoString(`result = Spinner.render({ frame = 2, label = "" })`)
	L.GetGlobal("result")
	if clStr(L, "content") != "⠙" { t.Errorf("frame2 = %q", clStr(L, "content")) }
	L.Pop(1)

	L.DoString(`result = Spinner.render({ frame = 11, label = "" })`)
	L.GetGlobal("result")
	if clStr(L, "content") != "⠋" { t.Errorf("frame11 wrap = %q", clStr(L, "content")) }
	L.Pop(1)
}

func TestTableComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		local function fit(s, w) s = tostring(s or ""); if #s > w then return s:sub(1,w-1).."…" end; return s .. string.rep(" ", w-#s) end
		TableC = lumina.defineComponent({
			name = "Table",
			init = function(props) return { columns = props.columns or {}, data = props.data or {}, selectedRow = props.selectedRow or 0 } end,
			render = function(inst)
				local cols, data, sel = inst.columns or {}, inst.data or {}, inst.selectedRow or 0
				local ch = {}
				local h = ""
				for i, col in ipairs(cols) do if i > 1 then h = h .. "│" end; h = h .. fit(col.label, col.width) end
				ch[#ch+1] = { type = "text", content = h, bold = true }
				local sep = ""
				for i, col in ipairs(cols) do if i > 1 then sep = sep .. "┼" end; sep = sep .. string.rep("─", col.width) end
				ch[#ch+1] = { type = "text", content = sep }
				for ri, row in ipairs(data) do
					local rc = ""
					for i, col in ipairs(cols) do if i > 1 then rc = rc .. "│" end; rc = rc .. fit(row[col.key], col.width) end
					ch[#ch+1] = { type = "text", content = rc, foreground = (ri == sel) and "#00FF00" or "#FFFFFF" }
				end
				return { type = "vbox", children = ch }
			end
		})
		result = TableC.render({
			columns = { { key="name", label="Name", width=10 }, { key="age", label="Age", width=5 } },
			data = { { name="Alice", age="30" }, { name="Bob", age="25" } },
			selectedRow = 1,
		})
	`)
	defer L.Close()
	L.GetGlobal("TableC")
	if clStr(L, "name") != "Table" { t.Errorf("name = %q", clStr(L, "name")) }
	L.Pop(1)

	L.GetGlobal("result")
	defer L.Pop(1)
	if n := clChildCount(L); n != 4 { t.Errorf("children = %d, want 4", n) }
	clPushChild(L, 1)
	if h := clStr(L, "content"); !strings.Contains(h, "Name") || !strings.Contains(h, "Age") { t.Errorf("header = %q", h) }
	L.Pop(1)
	clPushChild(L, 2)
	if s := clStr(L, "content"); !strings.Contains(s, "─") || !strings.Contains(s, "┼") { t.Errorf("sep = %q", s) }
	L.Pop(1)
	clPushChild(L, 3)
	if !strings.Contains(clStr(L, "content"), "Alice") { t.Errorf("row1 missing Alice") }
	if clStr(L, "foreground") != "#00FF00" { t.Errorf("sel fg = %q", clStr(L, "foreground")) }
	L.Pop(1)
}

func TestModalComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		Modal = lumina.defineComponent({
			name = "Modal",
			init = function(props) return { visible = props.visible ~= false, title = props.title or "Dialog", children = props.children or {}, buttons = props.buttons or {} } end,
			render = function(inst)
				if not inst.visible then return { type = "box" } end
				local body = {}
				body[#body+1] = { type = "text", content = " " .. inst.title .. " ", bold = true }
				for _, ch in ipairs(inst.children or {}) do body[#body+1] = ch end
				local btns = {}
				for _, b in ipairs(inst.buttons or {}) do btns[#btns+1] = { type = "text", content = "[ " .. b.label .. " ]" } end
				if #btns > 0 then body[#body+1] = { type = "hbox", children = btns } end
				return { type = "vbox", style = { border = "rounded" }, children = body }
			end
		})
		result = Modal.render({ visible = true, title = "Confirm", children = { { type = "text", content = "Sure?" } }, buttons = { { label = "Cancel" }, { label = "OK" } } })
	`)
	defer L.Close()
	clAssertFactory(t, L, "Modal", "Modal")
	L.GetGlobal("result")
	if clStr(L, "type") != "vbox" { t.Errorf("type = %q", clStr(L, "type")) }
	if n := clChildCount(L); n != 3 { t.Errorf("children = %d, want 3", n) }
	clPushChild(L, 1)
	if !strings.Contains(clStr(L, "content"), "Confirm") { t.Errorf("title = %q", clStr(L, "content")) }
	L.Pop(1)
	L.Pop(1)

	L.DoString(`result = Modal.render({ visible = false })`)
	L.GetGlobal("result")
	if clStr(L, "type") != "box" { t.Errorf("hidden type = %q", clStr(L, "type")) }
	L.Pop(1)
}

func TestTreeComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		Tree = lumina.defineComponent({
			name = "Tree",
			init = function(props) return { data = props.data or {}, expanded = props.expanded or {}, indent = 2 } end,
			render = function(inst)
				local ch = {}
				local exp = inst.expanded or {}
				local function add(nodes, depth)
					for _, n in ipairs(nodes) do
						local pfx = string.rep(" ", depth * 2)
						local has = n.children and #n.children > 0
						local icon = has and (exp[n.label] and "▾ " or "▸ ") or "  "
						ch[#ch+1] = { type = "text", content = pfx .. icon .. (n.label or ""), foreground = has and "#00FFFF" or "#FFFFFF" }
						if has and exp[n.label] then add(n.children, depth + 1) end
					end
				end
				add(inst.data, 0)
				if #ch == 0 then ch[#ch+1] = { type = "text", content = "(empty)" } end
				return { type = "vbox", children = ch }
			end
		})
		result = Tree.render({ data = { { label = "src", children = { { label = "main.lua" }, { label = "utils.lua" } } }, { label = "README.md" } }, expanded = { src = true } })
	`)
	defer L.Close()
	clAssertFactory(t, L, "Tree", "Tree")
	L.GetGlobal("result")
	if n := clChildCount(L); n != 4 { t.Errorf("expanded children = %d, want 4", n) }
	clPushChild(L, 1)
	if c := clStr(L, "content"); !strings.Contains(c, "▾") || !strings.Contains(c, "src") { t.Errorf("root = %q", c) }
	L.Pop(1)
	clPushChild(L, 2)
	if c := clStr(L, "content"); !strings.Contains(c, "main.lua") || !strings.HasPrefix(c, "  ") { t.Errorf("child = %q", c) }
	L.Pop(1)
	L.Pop(1)

	L.DoString(`result = Tree.render({ data = { { label = "src", children = { { label = "main.lua" } } }, { label = "README.md" } }, expanded = {} })`)
	L.GetGlobal("result")
	if n := clChildCount(L); n != 2 { t.Errorf("collapsed children = %d, want 2", n) }
	clPushChild(L, 1)
	if c := clStr(L, "content"); !strings.Contains(c, "▸") { t.Errorf("collapsed = %q", c) }
	L.Pop(1)
	L.Pop(1)
}

func TestStatusBarComponent(t *testing.T) {
	L := clRun(t, `
		local lumina = require("lumina")
		StatusBar = lumina.defineComponent({
			name = "StatusBar",
			init = function(props) return { left = props.left or "", center = props.center or "", right = props.right or "" } end,
			render = function(inst)
				local ch = {}
				ch[#ch+1] = { type = "text", content = " " .. inst.left }
				if inst.center ~= "" then ch[#ch+1] = { type = "text", content = inst.center, bold = true } end
				ch[#ch+1] = { type = "text", content = inst.right .. " " }
				return { type = "hbox", style = { height = 1 }, children = ch }
			end
		})
		result = StatusBar.render({ left = "Lumina v0.3", center = "main.lua", right = "Ln 42" })
	`)
	defer L.Close()
	clAssertFactory(t, L, "StatusBar", "StatusBar")
	L.GetGlobal("result")
	if clStr(L, "type") != "hbox" { t.Errorf("type = %q", clStr(L, "type")) }
	if n := clChildCount(L); n != 3 { t.Errorf("children = %d, want 3", n) }
	clPushChild(L, 1)
	if !strings.Contains(clStr(L, "content"), "Lumina v0.3") { t.Errorf("left = %q", clStr(L, "content")) }
	L.Pop(1)
	clPushChild(L, 2)
	if clStr(L, "content") != "main.lua" { t.Errorf("center = %q", clStr(L, "content")) }
	L.Pop(1)
	clPushChild(L, 3)
	if !strings.Contains(clStr(L, "content"), "Ln 42") { t.Errorf("right = %q", clStr(L, "content")) }
	L.Pop(1)
	L.Pop(1)

	L.DoString(`result = StatusBar.render({ left = "Left", center = "", right = "Right" })`)
	L.GetGlobal("result")
	if n := clChildCount(L); n != 2 { t.Errorf("no-center children = %d, want 2", n) }
	L.Pop(1)
}
