package v2

import (
	"fmt"
	"testing"
	"time"

	"github.com/akzj/lumina/pkg/render"
)

// TestScrollJumpToLatest verifies that "jump to latest" scroll behavior works
// correctly, including:
//   - Initial render: scroll container at top (scrollY=0)
//   - After scrolling up: scrollY < maxScroll
//   - After clicking "jump to latest" (which uses setTimeout → scrollTo): scrollY == maxScroll
//   - Second click of "jump to latest" still scrolls to bottom
func TestScrollJumpToLatest(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	// Create a component that mimics the chat panel pattern:
	// - A scroll container (vbox overflow=scroll) with many text children
	// - A "Jump to latest" button that uses setTimeout(fn, 0) to scroll after layout
	// - useState to control followingTail
	err := app.RunString(`
		local scrollId = "test-scroll"

		lumina.createComponent({
			id = "scroll-jump-test",
			name = "ScrollJumpTest",
			render = function(props)
				local followingTail, setFollowingTail = lumina.useState("followTail", true)

				local function scrollToBottomNextFrame()
					lumina.setTimeout(function()
						lumina.scrollTo(scrollId, 999999)
					end, 0)
				end

				local function jumpToLatest()
					setFollowingTail(true)
					scrollToBottomNextFrame()
				end

				local function handleScroll(ev)
					-- If user scrolls up, unfollow tail
					local info = lumina.getScrollInfo(scrollId)
					if info then
						local atBottom = (info.scrollY >= info.maxScroll - 1)
						if not atBottom then
							setFollowingTail(false)
						end
					end
				end

				-- Build 40 text children (each height=1, total=40, visible=20 → maxScroll=20)
				local msgChildren = {}
				for i = 1, 40 do
					msgChildren[i] = lumina.createElement("text", {
						key = "msg-" .. i,
						style = { height = 1 },
					}, "Message " .. i)
				end

				local children = {
					lumina.createElement("vbox", {
						id = scrollId,
						style = { overflow = "scroll", height = 20, width = 80 },
						onScroll = handleScroll,
					}, table.unpack(msgChildren)),
				}

				-- "Jump to latest" button (always visible for testing)
				children[#children + 1] = lumina.createElement("text", {
					id = "jump-btn",
					style = { height = 1, width = 80 },
					onClick = jumpToLatest,
				}, "Jump to latest")

				-- Initial load: scroll to bottom
				if followingTail then
					scrollToBottomNextFrame()
				end

				return lumina.createElement("vbox", {
					style = { height = 24, width = 80 },
				}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}

	app.RenderAll()

	// Fire timers to execute the initial scrollToBottomNextFrame
	time.Sleep(5 * time.Millisecond)
	app.FireTimers()
	app.RenderDirty()

	// Find the scroll container
	engine := app.Engine()
	scrollNode := engine.FindNodeByID("test-scroll")
	if scrollNode == nil {
		t.Fatal("scroll container 'test-scroll' not found")
	}
	if scrollNode.Style.Overflow != "scroll" {
		t.Fatalf("expected overflow=scroll, got %q", scrollNode.Style.Overflow)
	}

	maxScroll := computeMaxScrollYPublic(scrollNode)
	t.Logf("Initial state: ScrollY=%d, ScrollHeight=%d, H=%d, maxScroll=%d",
		scrollNode.ScrollY, scrollNode.ScrollHeight, scrollNode.H, maxScroll)

	if maxScroll <= 0 {
		t.Fatalf("maxScroll should be > 0 (content should overflow), got %d", maxScroll)
	}

	// STEP 1: After initial render + timer fire, should be at bottom
	if scrollNode.ScrollY != maxScroll {
		t.Errorf("STEP 1 - Initial scroll: ScrollY=%d, want maxScroll=%d", scrollNode.ScrollY, maxScroll)
	} else {
		t.Logf("STEP 1 PASS: Initial scroll at bottom (ScrollY=%d == maxScroll=%d)", scrollNode.ScrollY, maxScroll)
	}

	// STEP 2: Scroll up (simulate user scrolling up via mouse wheel)
	for i := 0; i < 20; i++ {
		engine.HandleScroll(10, 10, -1) // delta=-1 = scroll up
	}
	// Animate momentum scroll to completion
	for i := 0; i < 200; i++ {
		if !engine.TickSmoothScroll() {
			break
		}
	}

	if scrollNode.ScrollY >= maxScroll {
		t.Errorf("STEP 2 - After scroll up: ScrollY=%d should be < maxScroll=%d", scrollNode.ScrollY, maxScroll)
	} else {
		t.Logf("STEP 2 PASS: Scrolled up (ScrollY=%d < maxScroll=%d)", scrollNode.ScrollY, maxScroll)
	}
	savedScrollY := scrollNode.ScrollY

	// Render to propagate the scroll change → followingTail should become false
	app.RenderDirty()

	// STEP 3: Click "Jump to latest" button
	jumpBtn := engine.FindNodeByID("jump-btn")
	if jumpBtn == nil {
		t.Fatal("jump button 'jump-btn' not found")
	}
	t.Logf("Jump button at X=%d Y=%d W=%d H=%d", jumpBtn.X, jumpBtn.Y, jumpBtn.W, jumpBtn.H)

	// Click the button
	clickX := jumpBtn.X + jumpBtn.W/2
	clickY := jumpBtn.Y
	engine.HandleClick(clickX, clickY, "left")
	app.RenderDirty()

	// Before timer fires, scroll should still be at the saved position
	// (setTimeout hasn't fired yet)
	t.Logf("After click, before timer: ScrollY=%d", scrollNode.ScrollY)

	// Fire timers — this should execute the setTimeout(fn, 0) callback
	time.Sleep(5 * time.Millisecond)
	app.FireTimers()
	app.RenderDirty()

	// Recalculate maxScroll in case layout changed after re-render
	maxScroll = computeMaxScrollYPublic(scrollNode)

	if scrollNode.ScrollY != maxScroll {
		t.Errorf("STEP 3 - After jump to latest: ScrollY=%d, want maxScroll=%d (was %d before jump)",
			scrollNode.ScrollY, maxScroll, savedScrollY)
	} else {
		t.Logf("STEP 3 PASS: Jump to latest works (ScrollY=%d == maxScroll=%d)", scrollNode.ScrollY, maxScroll)
	}

	// STEP 4: Scroll up again, then click jump a SECOND time
	for i := 0; i < 15; i++ {
		engine.HandleScroll(10, 10, -1)
	}
	for i := 0; i < 200; i++ {
		if !engine.TickSmoothScroll() {
			break
		}
	}
	app.RenderDirty()

	if scrollNode.ScrollY >= maxScroll {
		t.Logf("WARNING: Could not scroll up for step 4 (ScrollY=%d, maxScroll=%d)", scrollNode.ScrollY, maxScroll)
	}

	// Second click on jump button
	engine.HandleClick(clickX, clickY, "left")
	app.RenderDirty()
	time.Sleep(5 * time.Millisecond)
	app.FireTimers()
	app.RenderDirty()

	maxScroll = computeMaxScrollYPublic(scrollNode)
	if scrollNode.ScrollY != maxScroll {
		t.Errorf("STEP 4 - Second jump to latest: ScrollY=%d, want maxScroll=%d", scrollNode.ScrollY, maxScroll)
	} else {
		t.Logf("STEP 4 PASS: Second jump to latest works (ScrollY=%d == maxScroll=%d)", scrollNode.ScrollY, maxScroll)
	}
}

// computeMaxScrollYPublic replicates the engine's computeMaxScrollY for use in tests
// outside the render package.
func computeMaxScrollYPublic(node *render.Node) int {
	bw := 0
	if node.Style.Border != "" {
		bw = 1
	}
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom
	if contentH <= 0 {
		return 0
	}
	maxScroll := node.ScrollHeight - contentH
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// TestScrollJumpToLatest_DynamicContent verifies that jump-to-latest works
// when content changes (new messages added) between scrolls.
func TestScrollJumpToLatest_DynamicContent(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunString(`
		_msgCount = 30
		_scrollId = "dynamic-scroll"

		lumina.createComponent({
			id = "dynamic-scroll-test",
			name = "DynamicScrollTest",
			render = function(props)
				local count, setCount = lumina.useState("count", _msgCount)

				local function scrollToBottomNextFrame()
					lumina.setTimeout(function()
						lumina.scrollTo(_scrollId, 999999)
					end, 0)
				end

				local function addMessages()
					setCount(count + 10)
					scrollToBottomNextFrame()
				end

				local function jumpToLatest()
					scrollToBottomNextFrame()
				end

				local msgChildren = {}
				for i = 1, count do
					msgChildren[i] = lumina.createElement("text", {
						key = "msg-" .. i,
						style = { height = 1 },
					}, "Message " .. i)
				end

				return lumina.createElement("vbox", {
					style = { height = 24, width = 80 },
				},
					lumina.createElement("vbox", {
						id = _scrollId,
						style = { overflow = "scroll", height = 20, width = 80 },
					}, table.unpack(msgChildren)),
					lumina.createElement("text", {
						id = "add-btn",
						style = { height = 1, width = 40 },
						onClick = addMessages,
					}, "Add messages"),
					lumina.createElement("text", {
						id = "jump-btn-2",
						style = { height = 1, width = 40 },
						onClick = jumpToLatest,
					}, "Jump to latest")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString: %v", err)
	}

	app.RenderAll()

	engine := app.Engine()
	scrollNode := engine.FindNodeByID("dynamic-scroll")
	if scrollNode == nil {
		t.Fatal("scroll container not found")
	}

	maxScroll := computeMaxScrollYPublic(scrollNode)
	t.Logf("Initial: ScrollY=%d, maxScroll=%d, ScrollHeight=%d, H=%d",
		scrollNode.ScrollY, maxScroll, scrollNode.ScrollHeight, scrollNode.H)

	// Scroll up
	for i := 0; i < 10; i++ {
		engine.HandleScroll(10, 10, -1)
	}
	for i := 0; i < 200; i++ {
		if !engine.TickSmoothScroll() {
			break
		}
	}
	savedY := scrollNode.ScrollY
	t.Logf("After scroll up: ScrollY=%d", savedY)

	// Click "Add messages" to change content
	addBtn := engine.FindNodeByID("add-btn")
	if addBtn == nil {
		t.Fatal("add button not found")
	}
	engine.HandleClick(addBtn.X+5, addBtn.Y, "left")
	app.RenderDirty()
	time.Sleep(5 * time.Millisecond)
	app.FireTimers()
	app.RenderDirty()

	newMaxScroll := computeMaxScrollYPublic(scrollNode)
	t.Logf("After adding messages: ScrollY=%d, newMaxScroll=%d (was %d)",
		scrollNode.ScrollY, newMaxScroll, maxScroll)

	if newMaxScroll <= maxScroll {
		t.Errorf("maxScroll should increase after adding messages: was %d, now %d", maxScroll, newMaxScroll)
	}

	if scrollNode.ScrollY != newMaxScroll {
		t.Errorf("After add+scroll: ScrollY=%d, want newMaxScroll=%d", scrollNode.ScrollY, newMaxScroll)
	} else {
		t.Logf("PASS: After adding messages, scrolled to new bottom (ScrollY=%d == maxScroll=%d)",
			scrollNode.ScrollY, newMaxScroll)
	}

	_ = fmt.Sprintf("test complete") // suppress unused import
}
