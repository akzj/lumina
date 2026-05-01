package widget

import (
	"testing"
)

func TestAllWidgetsEmpty(t *testing.T) {
	all := All()
	if len(all) != 0 {
		t.Fatalf("All() should return 0 widgets (all removed), got %d", len(all))
	}
}
