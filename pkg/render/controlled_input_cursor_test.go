package render

import (
	"testing"
)

// TestControlledInput_CursorNotResetOnFirstChar reproduces a bug where typing
// the first character in a controlled input inside a multi-level component
// hierarchy (Root → Wrapper → InputOwner) would reset the cursor to 0.
//
// Root cause: renderInOrder renders all dirty components sorted by depth in one
// pass. When Root renders, it marks Wrapper dirty (props changed). But
// InputOwner is already in the dirty batch and renders NEXT with stale props
// (value=""), clamping the cursor to 0. The fix: skip a dirty component if any
// ancestor is still dirty (it will be re-collected in the next iteration).
func TestControlledInput_CursorNotResetOnFirstChar(t *testing.T) {
	e, L := newTestEngine(t)

	// 3-level hierarchy:
	//   Root (depth 0) — holds state, passes value to Wrapper
	//   Wrapper (depth 1) — passes value through to InputOwner
	//   InputOwner (depth 2) — renders a controlled <input> with value=props.value
	err := L.DoString(`
		-- InputOwner: renders a controlled input
		InputOwner = lumina.defineComponent("InputOwner", function(props)
			return lumina.createElement("input", {
				id = "the-input",
				autoFocus = true,
				value = props.value or "",
				onChange = props.onChange,
			})
		end)

		-- Wrapper: just passes props through (the intermediate component)
		Wrapper = lumina.defineComponent("Wrapper", function(props)
			return lumina.createElement(InputOwner, {
				key = "inp",
				value = props.value,
				onChange = props.onChange,
			})
		end)

		-- Root: holds state
		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				local val, setVal = lumina.useState("val", "")
				-- Store setter globally so test can verify state
				_G.currentVal = val
				_G.setVal = setVal
				return lumina.createElement(Wrapper, {
					key = "wrap",
					value = val,
					onChange = function(newVal)
						setVal(newVal)
					end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Find the input node
	inputNode := findNodeByID(e.root.RootNode, "the-input")
	if inputNode == nil {
		t.Fatal("input node not found")
	}

	// Focus the input
	e.SetFocusedNode(inputNode)

	// Type 'a' — this triggers the bug scenario
	consumed := e.HandleInputKeyDown("a")
	if !consumed {
		t.Fatal("HandleInputKeyDown should consume 'a'")
	}

	// At this point:
	// - inputNode.Content = "a", CursorPos = 1 (set by HandleInputKeyDown)
	// - fireOnChange was called → Lua onChange → setVal("a") → Root marked dirty
	// - markOwnerDirty → InputOwner marked dirty
	// The bug: renderInOrder would render InputOwner with stale props (value="")
	// before Wrapper propagates the new value, clamping cursor to 0.

	e.RenderDirty()

	// After RenderDirty, we need to re-find the input node (tree may have been rebuilt)
	inputNode = findNodeByID(e.root.RootNode, "the-input")
	if inputNode == nil {
		t.Fatal("input node not found after RenderDirty")
	}

	// Also check via focused node
	focused := e.FocusedNode()
	if focused == nil {
		// Focus may have been lost during reconcile — use the found node
		focused = inputNode
	}

	if focused.Content != "a" {
		t.Errorf("content should be 'a', got %q", focused.Content)
	}
	if focused.CursorPos != 1 {
		t.Errorf("cursor should be 1 after typing 'a', got %d", focused.CursorPos)
	}

	// Type 'b' — should also work
	e.SetFocusedNode(focused)
	consumed = e.HandleInputKeyDown("b")
	if !consumed {
		t.Fatal("HandleInputKeyDown should consume 'b'")
	}

	e.RenderDirty()

	inputNode = findNodeByID(e.root.RootNode, "the-input")
	if inputNode == nil {
		t.Fatal("input node not found after second RenderDirty")
	}
	focused = e.FocusedNode()
	if focused == nil {
		focused = inputNode
	}

	if focused.Content != "ab" {
		t.Errorf("content should be 'ab', got %q", focused.Content)
	}
	if focused.CursorPos != 2 {
		t.Errorf("cursor should be 2 after typing 'b', got %d", focused.CursorPos)
	}
}

// findNodeByID already exists in input.go — reuse it.
