package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func shadcnAdvState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

// Phase 22: Form component require tests

func TestShadcn_Select(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Select = require("shadcn.select")
		assert(Select ~= nil)
		assert(Select.isComponent == true)
		assert(Select.name == "ShadcnSelect")
	`)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
}

func TestShadcn_Checkbox(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Checkbox = require("shadcn.checkbox")
		assert(Checkbox ~= nil)
		assert(Checkbox.isComponent == true)
		assert(Checkbox.name == "ShadcnCheckbox")
	`)
	if err != nil {
		t.Fatalf("Checkbox: %v", err)
	}
}

func TestShadcn_RadioGroup(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local RadioGroup = require("shadcn.radio_group")
		assert(RadioGroup ~= nil)
		assert(RadioGroup.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("RadioGroup: %v", err)
	}
}

func TestShadcn_Slider(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Slider = require("shadcn.slider")
		assert(Slider ~= nil)
		assert(Slider.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Slider: %v", err)
	}
}

func TestShadcn_Textarea(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Textarea = require("shadcn.textarea")
		assert(Textarea ~= nil)
		assert(Textarea.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Textarea: %v", err)
	}
}

func TestShadcn_Field(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Field = require("shadcn.field")
		assert(Field ~= nil)
		assert(Field.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Field: %v", err)
	}
}

func TestShadcn_InputGroup(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local InputGroup = require("shadcn.input_group")
		assert(InputGroup ~= nil)
		assert(InputGroup.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("InputGroup: %v", err)
	}
}

func TestShadcn_InputOTP(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local InputOTP = require("shadcn.input_otp")
		assert(InputOTP ~= nil)
		assert(InputOTP.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("InputOTP: %v", err)
	}
}

func TestShadcn_Combobox(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Combobox = require("shadcn.combobox")
		assert(Combobox ~= nil)
		assert(Combobox.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Combobox: %v", err)
	}
}

func TestShadcn_NativeSelect(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local NativeSelect = require("shadcn.native_select")
		assert(NativeSelect ~= nil)
		assert(NativeSelect.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("NativeSelect: %v", err)
	}
}

func TestShadcn_Form(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Form = require("shadcn.form")
		assert(Form ~= nil)
		assert(Form.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Form: %v", err)
	}
}

// Render tests for key form components

func TestShadcn_SelectRender(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Select = require("shadcn.select")
		local tree = lumina.render(Select, {
			options = { {value="a", label="Apple"}, {value="b", label="Banana"} },
			value = "a",
		})
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("SelectRender: %v", err)
	}
}

func TestShadcn_CheckboxRender(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Checkbox = require("shadcn.checkbox")
		local tree = lumina.render(Checkbox, { checked = true, label = "Accept terms" })
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("CheckboxRender: %v", err)
	}
}

func TestShadcn_SliderRender(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Slider = require("shadcn.slider")
		local tree = lumina.render(Slider, { value = 30, min = 0, max = 100, showValue = true })
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("SliderRender: %v", err)
	}
}

func TestShadcn_InitModule_WithForms(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local shadcn = require("shadcn")
		assert(shadcn.Select ~= nil, "Select should exist")
		assert(shadcn.Checkbox ~= nil, "Checkbox should exist")
		assert(shadcn.RadioGroup ~= nil, "RadioGroup should exist")
		assert(shadcn.Slider ~= nil, "Slider should exist")
		assert(shadcn.Textarea ~= nil, "Textarea should exist")
		assert(shadcn.Field ~= nil, "Field should exist")
		assert(shadcn.InputGroup ~= nil, "InputGroup should exist")
		assert(shadcn.InputOTP ~= nil, "InputOTP should exist")
		assert(shadcn.Combobox ~= nil, "Combobox should exist")
		assert(shadcn.NativeSelect ~= nil, "NativeSelect should exist")
		assert(shadcn.Form ~= nil, "Form should exist")
	`)
	if err != nil {
		t.Fatalf("InitModule with forms: %v", err)
	}
}

// Phase 23: Overlay component require tests

func TestShadcn_Dialog(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Dialog = require("shadcn.dialog")
		assert(Dialog ~= nil)
		assert(Dialog.isComponent == true)
		assert(Dialog.name == "ShadcnDialog")
	`)
	if err != nil {
		t.Fatalf("Dialog: %v", err)
	}
}

func TestShadcn_AlertDialog(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local AlertDialog = require("shadcn.alert_dialog")
		assert(AlertDialog ~= nil)
		assert(AlertDialog.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("AlertDialog: %v", err)
	}
}

func TestShadcn_Sheet(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Sheet = require("shadcn.sheet")
		assert(Sheet ~= nil)
		assert(Sheet.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Sheet: %v", err)
	}
}

func TestShadcn_Drawer(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Drawer = require("shadcn.drawer")
		assert(Drawer ~= nil)
		assert(Drawer.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Drawer: %v", err)
	}
}

func TestShadcn_DropdownMenu(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local DropdownMenu = require("shadcn.dropdown_menu")
		assert(DropdownMenu ~= nil)
		assert(DropdownMenu.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("DropdownMenu: %v", err)
	}
}

func TestShadcn_ContextMenu(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local ContextMenu = require("shadcn.context_menu")
		assert(ContextMenu ~= nil)
		assert(ContextMenu.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("ContextMenu: %v", err)
	}
}

func TestShadcn_Popover(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Popover = require("shadcn.popover")
		assert(Popover ~= nil)
		assert(Popover.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Popover: %v", err)
	}
}

func TestShadcn_Tooltip(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Tooltip = require("shadcn.tooltip")
		assert(Tooltip ~= nil)
		assert(Tooltip.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("Tooltip: %v", err)
	}
}

func TestShadcn_HoverCard(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local HoverCard = require("shadcn.hover_card")
		assert(HoverCard ~= nil)
		assert(HoverCard.isComponent == true)
	`)
	if err != nil {
		t.Fatalf("HoverCard: %v", err)
	}
}

// Render tests for overlay components

func TestShadcn_DialogRender(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local Dialog = require("shadcn.dialog")
		local tree = lumina.render(Dialog, {
			open = true,
			title = "Edit Profile",
			description = "Make changes to your profile here.",
		})
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("DialogRender: %v", err)
	}
}

func TestShadcn_DropdownMenuRender(t *testing.T) {
	L := shadcnAdvState(t)
	defer L.Close()
	err := L.DoString(`
		local DropdownMenu = require("shadcn.dropdown_menu")
		local tree = lumina.render(DropdownMenu, {
			open = true,
			trigger = "Actions",
			items = {
				{ label = "Edit", shortcut = "Ctrl+E" },
				{ separator = true },
				{ label = "Delete", shortcut = "Ctrl+D" },
			},
		})
		assert(tree ~= nil)
	`)
	if err != nil {
		t.Fatalf("DropdownMenuRender: %v", err)
	}
}
