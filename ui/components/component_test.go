package components

import "testing"

// Test that component automatic sizing works as expected
func TestComponentSizing(t *testing.T) {
	c := component{}
	c.SizeMin = Point{10, 10}
	c.SizeCurrent = Point{20, 20}
	c.SizeMax = Point{30, 30}

	// Min scaling
	c.Scaling = scalingMin
	max := c.outerCorner(0, 0, 15, 15)
	if max.X != 10 || max.Y != 10 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}

	max = c.outerCorner(0, 0, 40, 40)
	if max.X != 10 || max.Y != 10 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}

	// Max scaling
	c.Scaling = scalingMax
	max = c.outerCorner(0, 0, 40, 40)
	if max.X != 30 || max.Y != 30 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}

	max = c.outerCorner(0, 0, 20, 20)
	if max.X != 20 || max.Y != 20 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}

	// Disable scaling
	c.Scaling = scalingDisabled
	max = c.outerCorner(0, 0, 40, 40)
	if max.X != 20 || max.Y != 20 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}

	max = c.outerCorner(0, 0, 15, 15)
	if max.X != 15 || max.Y != 15 {
		t.Error("Invalid max size, got: ", max.X, max.Y)
	}
}
