package components

// Rectangle represents two points that encapsulate area
type Rectangle struct {
	X0, Y0 int
	X1, Y1 int
}

// Size returns size of the rectangle [x,y]
func (c *Rectangle) Size() (int, int) {
	x := abs(c.X1 - c.X0)
	y := abs(c.Y1 - c.Y0)
	return x, y
}

// Limit limits the rectangle size by setting x1&y1 to maximum of limit
func (c *Rectangle) Limit(x, y int) {
	c.X1 = min(c.X1, c.X0+x)
	c.Y1 = min(c.Y1, c.Y0+x)

}

// Sanitize ensures rectangle is positive
func (c *Rectangle) Sanitize() {
	if c.X0 < 0 {
		c.X0 = 0
	}
	if c.Y0 < 0 {
		c.Y0 = 0
	}

	if c.X1 < c.X0 {
		c.X1 = c.X0
	}
	if c.Y1 < c.Y0 {
		c.Y1 = c.Y0
	}
}

// Set fixed value to both points. Rectangle area is thus 0.
func (c *Rectangle) Set(val int) {
	c.X0 = val
	c.X1 = val
	c.Y0 = val
	c.Y1 = val
}

// Point represents single x-y pair
type Point struct {
	X, Y int
}

// List returns point as list of ints
func (v *Point) List() (int, int) {
	return v.X, v.Y
}
