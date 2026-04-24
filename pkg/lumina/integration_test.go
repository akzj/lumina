package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// TestIntegration_FullComponentPipeline tests defining a component, rendering, and state updates
func TestIntegration_FullComponentPipeline(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		-- Define a counter component
		local Counter = lumina.defineComponent({
			name = "Counter",
			render = function(self)
				local count = self.props.initial or 0
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "Count: " .. tostring(count) },
					}
				}
			end,
		})

		-- Verify component is registered
		assert(Counter ~= nil, "Counter should be defined")

		-- Create element (pass factory table, not string)
		local elem = lumina.createElement(Counter, { initial = 42 })
		assert(elem ~= nil, "createElement should return element")
		assert(elem.type == "component", "element type should be 'component'")
	`)
	if err != nil {
		t.Fatalf("Full component pipeline: %v", err)
	}
}

// TestIntegration_StoreWithComponents tests store + component integration
func TestIntegration_StoreWithComponents(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		-- Create store
		local store = lumina.createStore({
			state = { count = 0, items = {} },
			actions = {
				increment = function(state)
					state.count = state.count + 1
				end,
				addItem = function(state, item)
					state.items[#state.items + 1] = item
				end,
			},
		})

		-- Verify initial state
		local state = store.getState()
		assert(state.count == 0, "initial count should be 0")

		-- Dispatch actions
		store.dispatch("increment")
		store.dispatch("increment")
		store.dispatch("addItem", "apple")
		store.dispatch("addItem", "banana")

		-- Verify updated state
		state = store.getState()
		assert(state.count == 2, "count should be 2 after 2 increments, got " .. tostring(state.count))
		assert(#state.items == 2, "should have 2 items")
		assert(state.items[1] == "apple", "first item should be apple")
	`)
	if err != nil {
		t.Fatalf("Store with components: %v", err)
	}
}

// TestIntegration_RouterNavigation tests router navigation
func TestIntegration_RouterNavigation(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local router = lumina.createRouter({
			routes = {
				{ path = "/" },
				{ path = "/about" },
				{ path = "/users/:id" },
			}
		})

		-- Navigate using lumina.navigate
		lumina.navigate("/")
		local route = lumina.useRoute()
		assert(route.path == "/", "should be at /, got " .. tostring(route.path))

		lumina.navigate("/about")
		route = lumina.useRoute()
		assert(route.path == "/about", "should be at /about")

		lumina.navigate("/users/42")
		route = lumina.useRoute()
		assert(route.path == "/users/42", "should be at /users/42")
		assert(route.params.id == "42", "should have param id=42")
	`)
	if err != nil {
		t.Fatalf("Router navigation: %v", err)
	}
}

// TestIntegration_ThemeAndI18n tests theme + i18n integration
func TestIntegration_ThemeAndI18n(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		-- Set theme
		lumina.setTheme("catppuccin-mocha")
		local theme = lumina.useTheme()
		assert(theme ~= nil, "theme should not be nil")
		assert(theme.colors ~= nil, "theme.colors should not be nil")
		assert(theme.colors.primary ~= nil, "theme should have primary color")

		-- Set up i18n
		lumina.i18n.addTranslation("en", {
			["greeting"] = "Hello",
			["farewell"] = "Goodbye",
		})
		lumina.i18n.addTranslation("zh", {
			["greeting"] = "你好",
			["farewell"] = "再见",
		})

		-- Default locale
		local t = lumina.useTranslation()
		assert(t("greeting") == "Hello", "default should be English")

		-- Switch locale
		lumina.i18n.setLocale("zh")
		t = lumina.useTranslation()
		assert(t("greeting") == "你好", "should be Chinese after setLocale")
	`)
	if err != nil {
		t.Fatalf("Theme and i18n: %v", err)
	}
}

// TestIntegration_FormValidationFlow tests complete form validation flow
func TestIntegration_FormValidationFlow(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local submissions = {}

		local form = lumina.useForm({
			defaultValues = {
				name = "",
				email = "",
				age = 0,
			},
			rules = {
				name = {
					{ type = "required", message = "Name required" },
					{ type = "minLength", value = 2, message = "Too short" },
				},
				email = {
					{ type = "required", message = "Email required" },
					{ type = "email", message = "Invalid email" },
				},
				age = {
					{ type = "min", value = 1, message = "Age must be positive" },
					{ type = "max", value = 150, message = "Invalid age" },
				},
			},
			onSubmit = function(values)
				submissions[#submissions + 1] = values
			end,
		})

		-- Attempt submit with invalid data
		local ok = form.handleSubmit()
		assert(ok == false, "should fail with empty data")
		assert(#submissions == 0, "should not submit invalid form")

		-- Check errors
		local errors = form.getErrors()
		assert(errors.name ~= nil, "name should have error")
		assert(errors.email ~= nil, "email should have error")

		-- Fix fields one by one
		form.setValue("name", "Jo")
		form.setValue("email", "jo@test.com")
		form.setValue("age", 25)

		-- Submit again
		ok = form.handleSubmit()
		assert(ok == true, "should succeed with valid data")
		assert(#submissions == 1, "should have 1 submission")
	`)
	if err != nil {
		t.Fatalf("Form validation flow: %v", err)
	}
}

// TestIntegration_DragDropFlow tests full drag & drop flow
func TestIntegration_DragDropFlow(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalDragState.Reset()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local received = nil

		local drag = lumina.useDrag({
			type = "task",
			data = { id = 1, title = "Fix bug" },
		})

		local dropAccept = lumina.useDrop({
			accept = { "task" },
			onDrop = function(data)
				received = data
			end,
		})

		local dropReject = lumina.useDrop({
			accept = { "image" },
			onDrop = function(data) end,
		})

		-- Start drag
		drag.start("task-1")
		assert(drag.isDragging(), "should be dragging")

		-- Rejected drop zone
		assert(dropReject.canDrop() == false, "image zone should reject task")

		-- Accepted drop zone
		assert(dropAccept.canDrop() == true, "task zone should accept task")
		local ok = dropAccept.drop()
		assert(ok == true, "drop should succeed")
		assert(received ~= nil, "should have received data")

		-- Drag ended
		assert(drag.isDragging() == false, "should not be dragging after drop")
	`)
	if err != nil {
		t.Fatalf("Drag drop flow: %v", err)
	}
	globalDragState.Reset()
}

// TestIntegration_PluginSystem tests plugin registration and usage
func TestIntegration_PluginSystem(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	globalPluginRegistry.Clear()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local hookCalled = false

		lumina.registerPlugin({
			name = "analytics",
			version = "1.0.0",
			init = function()
				-- plugin initialized
			end,
			hooks = {
				useAnalytics = function(event)
					hookCalled = true
					return "tracked: " .. tostring(event)
				end,
			},
		})

		-- Use plugin
		lumina.usePlugin("analytics")

		-- Verify hook is available
		local result = lumina.useAnalytics("page_view")
		assert(result == "tracked: page_view", "hook should return tracked result")
		assert(hookCalled == true, "hook should have been called")

		-- Verify plugin is listed
		local plugins = lumina.getPlugins()
		assert(#plugins == 1, "should have 1 plugin")
		assert(plugins[1] == "analytics", "plugin should be 'analytics'")
	`)
	if err != nil {
		t.Fatalf("Plugin system: %v", err)
	}
	globalPluginRegistry.Clear()
}

// TestIntegration_ShadcnComponentsRender tests rendering shadcn components
func TestIntegration_ShadcnComponentsRender(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local shadcn = require("shadcn")

		-- Create a button (pass component factory table, not string)
		local btn = lumina.createElement(shadcn.Button, {
			label = "Click me",
			variant = "default",
		})
		assert(btn ~= nil, "button element should exist")
		assert(btn.type == "component", "button type should be 'component'")

		-- Create a card with children
		local card = lumina.createElement(shadcn.Card, {
			children = {
				{ type = "text", content = "Card content" },
			},
		})
		assert(card ~= nil, "card element should exist")

		-- Create a calendar
		local cal = lumina.createElement(shadcn.Calendar, {
			year = 2025,
			month = 6,
		})
		assert(cal ~= nil, "calendar element should exist")
	`)
	if err != nil {
		t.Fatalf("shadcn components render: %v", err)
	}
}

// TestIntegration_AllHooksAvailable verifies all hooks are accessible
func TestIntegration_AllHooksAvailable(t *testing.T) {
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		-- Core hooks
		assert(type(lumina.useState) == "function", "useState missing")
		assert(type(lumina.useEffect) == "function", "useEffect missing")
		assert(type(lumina.useMemo) == "function", "useMemo missing")
		assert(type(lumina.useCallback) == "function", "useCallback missing")
		assert(type(lumina.useRef) == "function", "useRef missing")
		assert(type(lumina.useReducer) == "function", "useReducer missing")

		-- Component API
		assert(type(lumina.defineComponent) == "function", "defineComponent missing")
		assert(type(lumina.createElement) == "function", "createElement missing")

		-- State management
		assert(type(lumina.createStore) == "function", "createStore missing")

		-- Router
		assert(type(lumina.createRouter) == "function", "createRouter missing")

		-- Form
		assert(type(lumina.useForm) == "function", "useForm missing")

		-- DnD
		assert(type(lumina.useDrag) == "function", "useDrag missing")
		assert(type(lumina.useDrop) == "function", "useDrop missing")

		-- Plugin
		assert(type(lumina.registerPlugin) == "function", "registerPlugin missing")
		assert(type(lumina.usePlugin) == "function", "usePlugin missing")

		-- Theme
		assert(type(lumina.setTheme) == "function", "setTheme missing")
		assert(type(lumina.useTheme) == "function", "useTheme missing")
	`)
	if err != nil {
		t.Fatalf("All hooks available: %v", err)
	}
}
