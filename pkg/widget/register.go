package widget

// All returns all built-in widgets.
func All() []*Widget {
	return []*Widget{
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
