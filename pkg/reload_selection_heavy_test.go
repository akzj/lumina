package v2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// TestHotReload_F5_SelectionHighlight_HeavyState reproduces the real app's
// pattern: ROOT has many useState hooks (20+), onClick handler calls multiple
// setState in one handler, deep nesting (Root → Page → Tree → ScrollView → Row).
func TestHotReload_F5_SelectionHighlight_HeavyState(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 20)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	script := `
		local ScrollView = require("lux.scrollview")

		local RowComp = lumina.defineComponent("HeavyRow", function(props)
			local bg = "#111111"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 40 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		local ListContainer = lumina.defineComponent("HeavyList", function(props)
			local rows = {}
			for i, item in ipairs(props.items) do
				rows[i] = lumina.createElement(RowComp, {
					key = "row-" .. item.id,
					label = item.label,
					isSelected = (props.selectedId == item.id),
					onSelect = function() props.onSelect(item.id) end,
				})
			end
			return lumina.createElement(ScrollView, {
				key = "heavy-scroll",
				style = { flex = 1, width = 40 },
			}, table.unpack(rows))
		end)

		local PageWrapper = lumina.defineComponent("HeavyPage", function(props)
			return lumina.createElement("vbox", {
				style = { width = 40, height = 20 },
			},
				lumina.createElement(ListContainer, {
					key = "list",
					items = props.items,
					selectedId = props.selectedId,
					onSelect = props.onSelect,
				}),
				lumina.createElement("text", {}, "sel=" .. (props.selectedId or "nil"))
			)
		end)

		lumina.app {
			id = "heavy-state-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "a")
				local s2, setS2 = lumina.useState("s2", "")
				local s3, setS3 = lumina.useState("s3", {})
				local s4, setS4 = lumina.useState("s4", {})
				local s5, setS5 = lumina.useState("s5", {})
				local s6, setS6 = lumina.useState("s6", {})
				local s7, setS7 = lumina.useState("s7", "")
				local s8, setS8 = lumina.useState("s8", {})
				local s9, setS9 = lumina.useState("s9", false)
				local s10, setS10 = lumina.useState("s10", false)
				local s11, setS11 = lumina.useState("s11", false)
				local s12, setS12 = lumina.useState("s12", false)
				local s13, setS13 = lumina.useState("s13", nil)
				local s14, setS14 = lumina.useState("s14", nil)
				local s15, setS15 = lumina.useState("s15", 0)

				local selRef = lumina.useRef("")
				selRef.current = sel
				local s3Ref = lumina.useRef({})
				s3Ref.current = s3
				local s4Ref = lumina.useRef({})
				s4Ref.current = s4

				local items = {
					{ id = "a", label = "Agent Alpha" },
					{ id = "b", label = "Agent Beta" },
					{ id = "c", label = "Agent Gamma" },
				}

				local function onSelectAgent(agentId)
					setSel(agentId)
					setS2("detail-" .. agentId)
					local newMsgs = {}
					for k, v in pairs(s3) do newMsgs[k] = v end
					newMsgs[agentId] = { "msg1", "msg2" }
					setS3(newMsgs)
					setS6({ llm = agentId })
					setS7("working memory for " .. agentId)
					setS8({ turns = 1 })
					setS5({ "key1", "key2" })
				end

				return lumina.createElement(PageWrapper, {
					key = "page",
					items = items,
					selectedId = sel,
					onSelect = onSelectAgent,
				})
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}
	flushScreen := func() {
		screen := app.engine.ToBuffer()
		_ = ta.WriteFull(screen)
	}
	click := func(x, y int) {
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y, Button: "left"})
		app.HandleEvent(&event.Event{Type: "mouseup", X: x, Y: y, Button: "left"})
		app.RenderDirty()
	}

	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("pre-reload initial: Row A bg = %q, want #00FF00", bg)
	}

	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Fatalf("pre-reload click B: bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Fatalf("pre-reload click B: expected 'sel=b'")
	}

	// F5 reload
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("post-reload initial: Row A bg = %q, want #00FF00", bg)
	}

	// Post-reload clicks
	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Errorf("post-reload click B: Row B bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 0); bg != "#111111" {
		t.Errorf("post-reload click B: Row A bg = %q, want #111111", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Errorf("post-reload click B: expected 'sel=b'")
	}

	click(5, 2)
	flushScreen()
	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Errorf("post-reload click C: Row C bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=c") {
		t.Errorf("post-reload click C: expected 'sel=c'")
	}

	click(5, 0)
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Errorf("post-reload click A: Row A bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=a") {
		t.Errorf("post-reload click A: expected 'sel=a'")
	}
}

// TestHotReload_F5_SelectionHighlight_DualDirty tests the case where a single
// click makes BOTH the root component AND a child component dirty simultaneously.
// This mimics the real app where handleSelectAgent calls:
// - onSelectAgent(agentId) → sets state on ROOT (selectedAgent, chatMessages, etc.)
// - openTab(...) → sets state on SessionPage (editorTabs, activeEditorTabId)
func TestHotReload_F5_SelectionHighlight_DualDirty(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 20)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	script := `
		local ScrollView = require("lux.scrollview")

		local RowComp = lumina.defineComponent("DDRow", function(props)
			local bg = "#111111"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 40 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		-- SessionPage equivalent: has its OWN state that gets set during onClick
		local SessionPage = lumina.defineComponent("DDSession", function(props)
			local selectedAgent = props.selectedAgent or ""
			local onSelectAgent = props.onSelectAgent

			-- SessionPage own state (like editorTabs, activeEditorTabId)
			local tabs, setTabs = lumina.useState("tabs", {})
			local activeTab, setActiveTab = lumina.useState("activeTab", "")
			local counter, setCounter = lumina.useState("counter", 0)

			local function handleSelectAgent(agentId)
				-- Call parent setter (sets state on ROOT)
				if onSelectAgent then onSelectAgent(agentId) end
				-- Also set OWN state (like openTab in real app)
				local newTabs = {}
				for _, t in ipairs(tabs) do newTabs[#newTabs + 1] = t end
				newTabs[#newTabs + 1] = agentId
				setTabs(newTabs)
				setActiveTab(agentId)
				setCounter(counter + 1)
			end

			local rows = {}
			local items = props.items or {}
			for i, item in ipairs(items) do
				rows[i] = lumina.createElement(RowComp, {
					key = "row-" .. item.id,
					label = item.label,
					isSelected = (selectedAgent == item.id),
					onSelect = function() handleSelectAgent(item.id) end,
				})
			end

			return lumina.createElement("vbox", {
				style = { width = 40, height = 20 },
			},
				lumina.createElement(ScrollView, {
					key = "dd-scroll",
					style = { flex = 1, width = 40 },
				}, table.unpack(rows)),
				lumina.createElement("text", {}, "sel=" .. selectedAgent .. " tab=" .. activeTab)
			)
		end)

		lumina.app {
			id = "dual-dirty-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "a")
				local chatMsgs, setChatMsgs = lumina.useState("msgs", {})
				local config, setConfig = lumina.useState("cfg", {})
				local memory, setMemory = lumina.useState("mem", "")
				local tasks, setTasks = lumina.useState("tasks", {})
				local ltmKeys, setLtmKeys = lumina.useState("ltm", {})

				local selRef = lumina.useRef("")
				selRef.current = sel
				local msgsRef = lumina.useRef({})
				msgsRef.current = chatMsgs

				local items = {
					{ id = "a", label = "Agent Alpha" },
					{ id = "b", label = "Agent Beta" },
					{ id = "c", label = "Agent Gamma" },
				}

				return lumina.createElement(SessionPage, {
					key = "session",
					items = items,
					selectedAgent = sel,
					onSelectAgent = function(agentId)
						setSel(agentId)
						local newMsgs = {}
						for k, v in pairs(chatMsgs) do newMsgs[k] = v end
						newMsgs[agentId] = { "hello" }
						setChatMsgs(newMsgs)
						setConfig({ agent = agentId })
						setMemory("wm:" .. agentId)
						setTasks({ t = agentId })
						setLtmKeys({ "k1" })
					end,
				})
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}
	flushScreen := func() {
		screen := app.engine.ToBuffer()
		_ = ta.WriteFull(screen)
	}
	click := func(x, y int) {
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y, Button: "left"})
		app.HandleEvent(&event.Event{Type: "mouseup", X: x, Y: y, Button: "left"})
		app.RenderDirty()
	}

	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("pre-reload initial: Row A bg = %q, want #00FF00", bg)
	}

	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Fatalf("pre-reload click B: bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Fatalf("pre-reload click B: expected 'sel=b'")
	}

	click(5, 2)
	flushScreen()
	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Fatalf("pre-reload click C: bg = %q, want #00FF00", bg)
	}

	// F5 reload
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("post-reload initial: Row A bg = %q, want #00FF00", bg)
	}

	// Post-reload clicks
	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Errorf("post-reload click B: Row B bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 0); bg != "#111111" {
		t.Errorf("post-reload click B: Row A bg = %q, want #111111", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Errorf("post-reload click B: expected 'sel=b'")
	}

	click(5, 2)
	flushScreen()
	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Errorf("post-reload click C: Row C bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 1); bg != "#111111" {
		t.Errorf("post-reload click C: Row B bg = %q, want #111111", bg)
	}
	if !screenHasString(ta, "sel=c") {
		t.Errorf("post-reload click C: expected 'sel=c'")
	}

	click(5, 0)
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Errorf("post-reload click A: Row A bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=a") {
		t.Errorf("post-reload click A: expected 'sel=a'")
	}
}

// TestHotReload_F5_SelectionHighlight_ManyRenderPasses tests that when a single
// click triggers many setState calls (causing multiple render passes in
// renderInOrder), all component onClick refs remain valid for subsequent clicks.
func TestHotReload_F5_SelectionHighlight_ManyRenderPasses(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	script := `
		local RowComp = lumina.defineComponent("MPRow", function(props)
			local bg = "#222222"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 40 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		lumina.app {
			id = "many-passes-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "a")
				local c1, setC1 = lumina.useState("c1", 0)
				local c2, setC2 = lumina.useState("c2", 0)
				local c3, setC3 = lumina.useState("c3", 0)
				local c4, setC4 = lumina.useState("c4", 0)
				local c5, setC5 = lumina.useState("c5", 0)
				local c6, setC6 = lumina.useState("c6", 0)
				local c7, setC7 = lumina.useState("c7", 0)
				local c8, setC8 = lumina.useState("c8", 0)
				local c9, setC9 = lumina.useState("c9", 0)
				local c10, setC10 = lumina.useState("c10", 0)

				return lumina.createElement("vbox", {
					style = { width = 40, height = 10 },
				},
					lumina.createElement(RowComp, {
						key = "row-a",
						label = "Row A",
						isSelected = (sel == "a"),
						onSelect = function()
							setSel("a")
							setC1(c1 + 1) setC2(c2 + 1) setC3(c3 + 1)
							setC4(c4 + 1) setC5(c5 + 1) setC6(c6 + 1)
							setC7(c7 + 1) setC8(c8 + 1) setC9(c9 + 1)
							setC10(c10 + 1)
						end,
					}),
					lumina.createElement(RowComp, {
						key = "row-b",
						label = "Row B",
						isSelected = (sel == "b"),
						onSelect = function()
							setSel("b")
							setC1(c1 + 1) setC2(c2 + 1) setC3(c3 + 1)
							setC4(c4 + 1) setC5(c5 + 1) setC6(c6 + 1)
							setC7(c7 + 1) setC8(c8 + 1) setC9(c9 + 1)
							setC10(c10 + 1)
						end,
					}),
					lumina.createElement(RowComp, {
						key = "row-c",
						label = "Row C",
						isSelected = (sel == "c"),
						onSelect = function()
							setSel("c")
							setC1(c1 + 1) setC2(c2 + 1) setC3(c3 + 1)
							setC4(c4 + 1) setC5(c5 + 1) setC6(c6 + 1)
							setC7(c7 + 1) setC8(c8 + 1) setC9(c9 + 1)
							setC10(c10 + 1)
						end,
					}),
					lumina.createElement("text", {}, "sel=" .. sel .. " clicks=" .. tostring(c1))
				)
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}
	flushScreen := func() {
		screen := app.engine.ToBuffer()
		_ = ta.WriteFull(screen)
	}
	click := func(x, y int) {
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y, Button: "left"})
		app.HandleEvent(&event.Event{Type: "mouseup", X: x, Y: y, Button: "left"})
		app.RenderDirty()
	}

	flushScreen()
	click(5, 1) // B
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Fatalf("pre-reload click B: bg = %q, want #00FF00", bg)
	}

	// F5 reload
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})
	flushScreen()

	// Post-reload clicks
	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Errorf("post-reload click B: bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Errorf("post-reload click B: expected 'sel=b'")
	}

	click(5, 2)
	flushScreen()
	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Errorf("post-reload click C: bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=c") {
		t.Errorf("post-reload click C: expected 'sel=c'")
	}

	click(5, 0)
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Errorf("post-reload click A: bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "sel=a") {
		t.Errorf("post-reload click A: expected 'sel=a'")
	}
}
