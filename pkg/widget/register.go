package widget

// All returns all built-in widgets.
func All() []*Widget {
	return []*Widget{
		Checkbox,
		Switch,
		Radio,
		Label,
		Select,
		Tooltip,
		Table,
		Menu,
		Dropdown,
		Spacer,
		Window,
		ScrollView,
	}
}
