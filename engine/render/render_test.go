package render

import "testing"

func TestFit(t *testing.T) {
	cases := []struct {
		name              string
		cw, ch, ww, wh    int
		scale, offX, offY int
	}{
		{"exact 1x", 320, 180, 320, 180, 1, 0, 0},
		{"1080p is 6x", 320, 180, 1920, 1080, 6, 0, 0},
		{"odd window letterboxes", 320, 180, 1000, 700, 3, 20, 80},
		{"tiny window never scales below 1", 320, 180, 100, 100, 1, -110, -40},
		{"4:3 window pillarboxes vertically", 320, 180, 1280, 1024, 4, 0, 152},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, x, y := Fit(c.cw, c.ch, c.ww, c.wh)
			if s != c.scale || x != c.offX || y != c.offY {
				t.Fatalf("Fit = (%d,%d,%d), want (%d,%d,%d)", s, x, y, c.scale, c.offX, c.offY)
			}
		})
	}
}

func TestCameraClampsToBounds(t *testing.T) {
	cam := NewCamera(320, 180)
	cam.Bounds = &Rect{X: 0, Y: 0, W: 640, H: 360}

	cam.Center(0, 0)
	if cam.X != 0 || cam.Y != 0 {
		t.Fatalf("top-left clamp: got (%v,%v)", cam.X, cam.Y)
	}
	cam.Center(10000, 10000)
	if cam.X != 320 || cam.Y != 180 {
		t.Fatalf("bottom-right clamp: got (%v,%v)", cam.X, cam.Y)
	}
}

func TestCameraSmallMapPinsTopLeft(t *testing.T) {
	cam := NewCamera(320, 180)
	cam.Bounds = &Rect{X: 0, Y: 0, W: 160, H: 90}
	cam.Center(80, 45)
	if cam.X != 0 || cam.Y != 0 {
		t.Fatalf("small map should pin to origin, got (%v,%v)", cam.X, cam.Y)
	}
}

func TestCameraFollowConverges(t *testing.T) {
	cam := NewCamera(320, 180)
	cam.Lerp = 0.5
	for i := 0; i < 60; i++ {
		cam.Follow(1000, 500)
	}
	if d := cam.X - (1000 - 160); d > 0.01 || d < -0.01 {
		t.Fatalf("camera did not converge on x, off by %v", d)
	}
	ox, oy := cam.Offset()
	if ox != -840 || oy != -410 {
		t.Fatalf("offset should be whole pixels, got (%v,%v)", ox, oy)
	}
}
