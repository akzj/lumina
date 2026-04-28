package widget

// All returns all built-in widgets.
func All() []*Widget {
	return []*Widget{
		Button,
		Checkbox,
		Switch,
		Radio,
		Label,
		Select,
	}
}
