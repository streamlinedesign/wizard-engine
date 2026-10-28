package render

import "math"

// Camera is a world-space window the size of the canvas. Positions are kept
// as floats so following is smooth, but Offset snaps to whole pixels so tiles
// never shimmer at sub-pixel positions.
type Camera struct {
	// X, Y is the world position of the top-left corner of the view.
	X, Y float64
	// Width, Height is the view size in world pixels (normally the canvas size).
	Width, Height int

	// Bounds, when set, clamps the view inside a rectangle (usually the map).
	Bounds *Rect
	// Lerp is how much of the remaining distance to the target the camera
	// closes per tick. 1 means snap instantly.
	Lerp float64
}

// Rect is an axis-aligned rectangle in world pixels.
type Rect struct {
	X, Y, W, H float64
}

// NewCamera makes a camera with a comfortable default follow speed.
func NewCamera(width, height int) *Camera {
	return &Camera{Width: width, Height: height, Lerp: 0.15}
}

// Follow moves the camera toward centring on (tx, ty).
func (c *Camera) Follow(tx, ty float64) {
	wantX := tx - float64(c.Width)/2
	wantY := ty - float64(c.Height)/2
	c.X += (wantX - c.X) * c.Lerp
	c.Y += (wantY - c.Y) * c.Lerp
	c.clamp()
}

// Center snaps the camera so (tx, ty) is in the middle of the view.
func (c *Camera) Center(tx, ty float64) {
	c.X = tx - float64(c.Width)/2
	c.Y = ty - float64(c.Height)/2
	c.clamp()
}

func (c *Camera) clamp() {
	if c.Bounds == nil {
		return
	}
	maxX := c.Bounds.X + c.Bounds.W - float64(c.Width)
	maxY := c.Bounds.Y + c.Bounds.H - float64(c.Height)
	// When the map is smaller than the view, keep it pinned at the top-left of
	// the bounds rather than jittering between two clamps.
	if maxX < c.Bounds.X {
		maxX = c.Bounds.X
	}
	if maxY < c.Bounds.Y {
		maxY = c.Bounds.Y
	}
	c.X = math.Min(math.Max(c.X, c.Bounds.X), maxX)
	c.Y = math.Min(math.Max(c.Y, c.Bounds.Y), maxY)
}

// Offset is the whole-pixel translation to apply when drawing world objects.
func (c *Camera) Offset() (ox, oy float64) {
	return -math.Floor(c.X + 0.5), -math.Floor(c.Y + 0.5)
}

// Visible reports the world-pixel rectangle currently on screen, snapped
// the same way Offset is.
func (c *Camera) Visible() Rect {
	ox, oy := c.Offset()
	return Rect{X: -ox, Y: -oy, W: float64(c.Width), H: float64(c.Height)}
}
