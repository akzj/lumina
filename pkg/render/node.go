package render

import "reflect"

// LuaRef is a Lua registry reference. 0 means nil/unset.
type LuaRef = int64

// Style holds layout and visual properties.
type Style struct {
	// Sizing
	Width, Height        int // 0 = auto
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	Flex                 int // flex grow factor

	// Spacing
	Padding                    int
	PaddingTop, PaddingBottom  int
	PaddingLeft, PaddingRight  int
	Margin                     int
	MarginTop, MarginBottom    int
	MarginLeft, MarginRight    int
	Gap                        int

	// Alignment
	Justify string // "start", "center", "end", "space-between", "space-around"
	Align   string // "stretch", "start", "center", "end"

	// Visual
	Border     string // "none", "single", "double", "rounded"
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Underline  bool

	// Overflow
	Overflow string // "hidden", "scroll"

	// Positioning
	Position      string // "", "relative", "absolute", "fixed"
	Top, Left     int
	Right, Bottom int // -1 = unset
	ZIndex        int
}

// Node is a persistent UI node. Created once, updated in-place.
// Never garbage collected during normal operation.
type Node struct {
	// Identity
	Type string // "box", "vbox", "hbox", "text", "input", "textarea", "component"
	ID   string // from props.id
	Key  string // from props.key (for reconciliation)

	// Tree (persistent, not rebuilt)
	Parent   *Node
	Children []*Node

	// Layout (cached, computed by layout engine)
	X, Y, W, H int  // absolute position + size
	LayoutDirty bool // true → recompute this subtree's layout

	// Paint
	Content     string // text content
	Placeholder string // placeholder text (input/textarea)
	Style       Style  // visual style
	PaintDirty  bool   // true → repaint this node

	// Input/textarea state
	CursorPos int  // cursor position within Content
	AutoFocus bool // auto-focus on mount

	// Events (persistent references, not re-registered per frame)
	OnClick      LuaRef
	OnMouseEnter LuaRef
	OnMouseLeave LuaRef
	OnKeyDown    LuaRef

	// Lifecycle
	Removed bool // true when orphaned from tree (stale pointer guard)
	OnChange     LuaRef
	OnScroll     LuaRef

	// Component (if this is a component root node)
	Component     *Component // nil for plain elements
	ComponentType string     // factory name for type="component" nodes

	// Scroll state
	ScrollY      int
	ScrollHeight int // total content height (set by layout for overflow=scroll containers)
}

// Component is a stateful UI unit.
// hookKind identifies the type of a positional hook slot.
type hookKind int

const (
	hookEffect   hookKind = iota
	hookMemo              // also used for useCallback
	hookRef
)

// effectSlot stores state for a single useEffect call.
type effectSlot struct {
	deps        []any  // dependency values from last render
	callbackRef int    // Lua registry ref to effect function
	cleanupRef  int    // Lua registry ref to cleanup function (0 = none)
	pending     bool   // needs to fire this cycle
}

// memoSlot stores state for a single useMemo/useCallback call.
type memoSlot struct {
	deps []any // dependency values from last render
	ref  int   // Lua registry ref to cached value
}

// refSlot stores state for a single useRef call.
type refSlot struct {
	tableRef int // Lua registry ref to the {current=...} table
}

// hookSlot is a tagged union for positional hook tracking.
type hookSlot struct {
	kind   hookKind
	effect *effectSlot
	memo   *memoSlot
	ref    *refSlot
}

type Component struct {
	ID   string
	Type string // factory name (for reconciliation matching)
	Name string

	// State
	Props map[string]any
	State map[string]any
	Dirty bool // needs re-render

	// Lua
	RenderFn LuaRef // Lua registry ref to render function

	// Tree
	Parent   *Component
	Children []*Component
	ChildMap map[string]*Component // "type:key" → child

	// Render output
	RootNode *Node // the RenderNode subtree this component owns

	// Lifecycle
	IsRoot    bool
	Mounted   bool
	LastError string // last render error (empty = no error)

	// Hooks (React-style positional)
	hookIdx   int         // current hook call index (reset each render)
	hookSlots []*hookSlot // ordered hook slots
}

// NewNode creates a new Node with the given type.
func NewNode(nodeType string) *Node {
	return &Node{
		Type: nodeType,
		Style: Style{
			Right:  -1,
			Bottom: -1,
		},
	}
}

// NewComponent creates a new Component.
func NewComponent(id, typeName, name string) *Component {
	return &Component{
		ID:       id,
		Type:     typeName,
		Name:     name,
		Props:    make(map[string]any),
		State:    make(map[string]any),
		ChildMap: make(map[string]*Component),
		Dirty:    true, // needs initial render
	}
}

// AddChild appends a child node and sets its parent.
func (n *Node) AddChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// RemoveChild removes a child node.
func (n *Node) RemoveChild(child *Node) {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}

// MarkPaintDirty marks this node for repaint.
func (n *Node) MarkPaintDirty() {
	n.PaintDirty = true
}

// MarkLayoutDirty marks this node's subtree for re-layout.
// Propagates up to the nearest fixed-size ancestor.
func (n *Node) MarkLayoutDirty() {
	n.LayoutDirty = true
	// Propagate up: parent needs re-layout if child size might change
	p := n.Parent
	for p != nil {
		if p.LayoutDirty {
			break // already dirty above
		}
		// If parent has fixed size, it acts as layout boundary
		if p.Style.Width > 0 && p.Style.Height > 0 {
			p.LayoutDirty = true
			break
		}
		p.LayoutDirty = true
		p = p.Parent
	}
}

// SetState sets a state value and marks the component dirty.
func (c *Component) SetState(key string, value any) {
	if reflect.DeepEqual(c.State[key], value) {
		return // no change
	}
	c.State[key] = value
	c.Dirty = true
}

// FindChild finds a child component by type+key.
func (c *Component) FindChild(typeName, key string) *Component {
	mapKey := typeName
	if key != "" {
		mapKey = typeName + ":" + key
	}
	return c.ChildMap[mapKey]
}

// AddChild adds a child component with an explicit lookup key for the ChildMap.
// The lookupKey should match what FindChild will use (typically node.ID or node.Key).
func (c *Component) AddChild(child *Component, lookupKey ...string) {
	child.Parent = c
	c.Children = append(c.Children, child)
	var mapKey string
	if len(lookupKey) > 0 && lookupKey[0] != "" {
		mapKey = child.Type + ":" + lookupKey[0]
	} else {
		mapKey = child.Type
	}
	c.ChildMap[mapKey] = child
}
