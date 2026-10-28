// Command mockup renders one small campus scene in different art styles so
// tile size and palette can be compared side by side, without the engine.
// It draws straight into PNGs with procedural detail: it is a sketch of
// what each style could look like, not finished art.
//
//	go run ./tools/mockup out/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

// Style is a palette plus proportions.
type Style struct {
	Name string
	T    int // tile size in pixels
	// Character cell height in tiles-worth of pixels (2 = 2 tiles tall).
	CharH float64
	// HeadFrac is the head's share of character height. 0.33 is chibi,
	// 0.2 is adult.
	HeadFrac float64
	Outline  color.RGBA
	Grass    [3]color.RGBA // base, dark, light
	Path     [3]color.RGBA
	Plaster  [2]color.RGBA
	Timber   color.RGBA
	Roof     [3]color.RGBA
	Leaf     [3]color.RGBA
	Trunk    color.RGBA
	Robe     [2]color.RGBA
	Hat      [2]color.RGBA
	Skin     color.RGBA
	Beard    color.RGBA
	Glow     color.RGBA
	Detail   bool // extra texture passes (dither, blades, mortar)
}

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

var current = Style{
	Name: "A_current_16", T: 16, CharH: 2, HeadFrac: 0.3,
	Outline: rgb(24, 20, 37),
	Grass:   [3]color.RGBA{rgb(104, 150, 84), rgb(84, 126, 68), rgb(124, 168, 98)},
	Path:    [3]color.RGBA{rgb(168, 160, 150), rgb(138, 130, 122), rgb(190, 182, 170)},
	Plaster: [2]color.RGBA{rgb(206, 190, 164), rgb(176, 160, 136)},
	Timber:  rgb(120, 100, 84),
	Roof:    [3]color.RGBA{rgb(132, 72, 70), rgb(98, 52, 54), rgb(160, 92, 86)},
	Leaf:    [3]color.RGBA{rgb(60, 120, 70), rgb(40, 88, 52), rgb(90, 150, 90)},
	Trunk:   rgb(88, 60, 40),
	Robe:    [2]color.RGBA{rgb(74, 84, 160), rgb(46, 52, 112)},
	Hat:     [2]color.RGBA{rgb(58, 68, 140), rgb(90, 105, 190)},
	Skin:    rgb(232, 190, 150), Beard: rgb(230, 230, 236),
	Glow: rgb(255, 214, 120),
}

// muted is the same tile size with a desaturated, cooler palette, a
// softer outline and adult proportions.
var muted = Style{
	Name: "B_muted_16", T: 16, CharH: 2.5, HeadFrac: 0.22,
	Outline: rgb(38, 34, 46),
	Grass:   [3]color.RGBA{rgb(92, 118, 72), rgb(74, 96, 58), rgb(108, 134, 84)},
	Path:    [3]color.RGBA{rgb(128, 118, 106), rgb(104, 96, 86), rgb(146, 136, 124)},
	Plaster: [2]color.RGBA{rgb(190, 178, 158), rgb(160, 148, 130)},
	Timber:  rgb(90, 66, 48),
	Roof:    [3]color.RGBA{rgb(78, 72, 88), rgb(58, 52, 66), rgb(96, 90, 108)},
	Leaf:    [3]color.RGBA{rgb(62, 96, 62), rgb(42, 68, 46), rgb(84, 118, 76)},
	Trunk:   rgb(74, 54, 40),
	Robe:    [2]color.RGBA{rgb(56, 60, 100), rgb(38, 40, 70)},
	Hat:     [2]color.RGBA{rgb(44, 46, 80), rgb(66, 70, 112)},
	Skin:    rgb(214, 176, 142), Beard: rgb(206, 204, 208),
	Glow: rgb(240, 200, 120), Detail: true,
}

func main() {
	out := "mockups"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		fail(err)
	}
	big := muted
	big.Name, big.T = "C_muted_32", 32
	for _, st := range []Style{current, muted, big} {
		img := renderScene(st)
		// Scale so every mock-up is the same on-screen size (1280 wide).
		scale := 1280 / img.Bounds().Dx()
		if err := writePNG(filepath.Join(out, st.Name+".png"), upscale(img, scale)); err != nil {
			fail(err)
		}
		fmt.Println("wrote", st.Name)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Scene layout in tiles: 20 x 11.
const sceneW, sceneH = 20, 11

func renderScene(st Style) *image.RGBA {
	T := st.T
	img := image.NewRGBA(image.Rect(0, 0, sceneW*T, sceneH*T))
	rng := rand.New(rand.NewSource(5))
	c := &canvas{img: img, st: st, rng: rng}

	for ty := 0; ty < sceneH; ty++ {
		for tx := 0; tx < sceneW; tx++ {
			c.grass(tx*T, ty*T)
		}
	}
	// Path: vertical down the middle, horizontal across row 7-8.
	for ty := 0; ty < sceneH; ty++ {
		c.path(9*T, ty*T)
		c.path(10*T, ty*T)
	}
	for tx := 0; tx < sceneW; tx++ {
		c.path(tx*T, 7*T)
		c.path(tx*T, 8*T)
	}
	// Building top-left: 6 wide, roof 3 rows, wall 3 rows.
	c.building(1*T, 0*T, 6, 3, 3)
	// Building top-right.
	c.building(13*T, 1*T, 6, 3, 2)
	// Trees.
	for _, p := range [][2]int{{0, 8}, {3, 9}, {15, 9}, {18, 8}, {12, 4}, {17, 5}, {7, 4}} {
		c.tree(p[0]*T, p[1]*T)
	}
	// Lamps by the path.
	c.lamp(8*T, 5*T)
	c.lamp(11*T, 9*T)
	// Wizard on the path, feet near the junction.
	c.wizard(float64(10*T), float64(7*T+T/2))
	return img
}

type canvas struct {
	img *image.RGBA
	st  Style
	rng *rand.Rand
}

func (c *canvas) px(x, y int, col color.RGBA) {
	if image.Pt(x, y).In(c.img.Bounds()) {
		c.img.SetRGBA(x, y, col)
	}
}

func (c *canvas) rect(x0, y0, x1, y1 int, col color.RGBA) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c.px(x, y, col)
		}
	}
}

func (c *canvas) disc(cx, cy, r float64, col color.RGBA) {
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				c.px(x, y, col)
			}
		}
	}
}

func (c *canvas) grass(x, y int) {
	T, st := c.st.T, c.st
	c.rect(x, y, x+T-1, y+T-1, st.Grass[0])
	n := T * T / 40
	for i := 0; i < n; i++ {
		c.px(x+c.rng.Intn(T), y+c.rng.Intn(T), st.Grass[1])
	}
	for i := 0; i < n/2; i++ {
		c.px(x+c.rng.Intn(T), y+c.rng.Intn(T), st.Grass[2])
	}
	if st.Detail {
		// Grass blades: short vertical strokes.
		for i := 0; i < T/4; i++ {
			bx, by := x+c.rng.Intn(T), y+c.rng.Intn(T-2)
			c.px(bx, by, st.Grass[2])
			c.px(bx, by+1, st.Grass[1])
		}
	}
}

func (c *canvas) path(x, y int) {
	T, st := c.st.T, c.st
	c.rect(x, y, x+T-1, y+T-1, st.Path[1])
	if !st.Detail {
		c.rect(x, y, x+T-1, y+T-1, st.Path[0])
		for i := 0; i < T*T/20; i++ {
			c.px(x+c.rng.Intn(T), y+c.rng.Intn(T), st.Path[1])
		}
		return
	}
	// Cobbles: a staggered grid of rounded stones with mortar between.
	stone := T / 4
	if stone < 3 {
		stone = 3
	}
	for row := 0; row*stone < T; row++ {
		off := (row % 2) * stone / 2
		for col := -1; col*stone < T; col++ {
			sx, sy := x+col*stone+off, y+row*stone
			shade := st.Path[0]
			if c.rng.Intn(4) == 0 {
				shade = st.Path[2]
			}
			c.rect(sx+1, sy+1, sx+stone-2, sy+stone-2, shade)
			c.px(sx+1, sy+stone-2, st.Path[1])
			c.px(sx+stone-2, sy+stone-2, st.Path[1])
		}
	}
}

func (c *canvas) building(x, y, w, roofRows, wallRows int) {
	T, st := c.st.T, c.st
	W := w * T
	// Roof.
	c.rect(x, y, x+W-1, y+roofRows*T-1, st.Roof[0])
	rowH := T / 4
	if st.Detail {
		rowH = T / 6
	}
	if rowH < 2 {
		rowH = 2
	}
	for ry := y; ry < y+roofRows*T; ry += rowH {
		c.rect(x, ry, x+W-1, ry, st.Roof[1])
		off := ((ry - y) / rowH % 2) * rowH
		for tx := x + off; tx < x+W; tx += rowH * 2 {
			c.rect(tx, ry+1, tx, ry+rowH-1, st.Roof[1])
			if st.Detail {
				c.px(tx+1, ry+1, st.Roof[2])
			}
		}
	}
	// Ridge and eave.
	c.rect(x, y, x+W-1, y+1, st.Roof[1])
	eave := y + roofRows*T
	c.rect(x-1, eave-2, x+W, eave-1, st.Outline)
	// Wall.
	wallTop := eave
	wallBot := eave + wallRows*T
	c.rect(x, wallTop, x+W-1, wallBot-1, st.Plaster[0])
	if st.Detail {
		// Timber frame and plaster shading.
		c.rect(x, wallTop, x+W-1, wallTop+1, st.Timber)
		for tx := x; tx < x+W; tx += T {
			c.rect(tx, wallTop, tx+1, wallBot-1, st.Timber)
		}
		c.rect(x+W-2, wallTop, x+W-1, wallBot-1, st.Timber)
		for i := 0; i < W*wallRows*T/60; i++ {
			c.px(x+c.rng.Intn(W), wallTop+c.rng.Intn(wallRows*T), st.Plaster[1])
		}
	} else {
		for ry := wallTop; ry < wallBot; ry += T / 4 {
			c.rect(x, ry, x+W-1, ry, st.Plaster[1])
		}
	}
	// Shadow under the eave.
	c.rect(x, wallTop, x+W-1, wallTop+T/8, st.Plaster[1])
	// Door in the middle, windows either side.
	dw, dh := T*3/4, T
	dx := x + W/2 - dw/2
	c.rect(dx-1, wallBot-dh-1, dx+dw, wallBot-1, st.Outline)
	c.rect(dx, wallBot-dh, dx+dw-1, wallBot-1, st.Timber)
	if st.Detail {
		c.rect(dx+dw/2, wallBot-dh, dx+dw/2, wallBot-1, st.Outline)
		c.px(dx+dw-3, wallBot-dh/2, st.Glow)
	}
	for _, wx := range []int{x + T/2, x + W - T/2 - T/2} {
		wy := wallTop + T/2
		c.rect(wx-1, wy-1, wx+T/2, wy+T/2, st.Outline)
		c.rect(wx, wy, wx+T/2-1, wy+T/2-1, rgb(120, 150, 170))
		if st.Detail {
			c.rect(wx+T/4, wy, wx+T/4, wy+T/2-1, st.Outline)
			c.rect(wx, wy+T/4, wx+T/2-1, wy+T/4, st.Outline)
			c.px(wx+1, wy+1, rgb(200, 220, 230))
		}
	}
	// Outline the whole facade sides.
	c.rect(x-1, wallTop, x-1, wallBot-1, st.Outline)
	c.rect(x+W, wallTop, x+W, wallBot-1, st.Outline)
	c.rect(x-1, wallBot, x+W, wallBot, st.Outline)
}

func (c *canvas) tree(x, y int) {
	T, st := c.st.T, c.st
	cx, base := float64(x+T/2), float64(y+2*T)
	// Shadow and trunk.
	c.disc(cx, base-2, float64(T)*0.45, st.Grass[1])
	c.rect(int(cx)-T/10-1, y+T, int(cx)+T/10, int(base)-2, st.Trunk)
	// Canopy: stacked clusters, darker on the right.
	r := float64(T) * 0.62
	c.disc(cx, float64(y)+r+2, r, st.Outline)
	c.disc(cx, float64(y)+r+2, r-1, st.Leaf[0])
	c.disc(cx+r*0.35, float64(y)+r+6, r*0.7, st.Leaf[1])
	c.disc(cx-r*0.35, float64(y)+r*0.9, r*0.55, st.Leaf[2])
	if st.Detail {
		for i := 0; i < T/2; i++ {
			a := c.rng.Float64() * 2 * math.Pi
			d := c.rng.Float64() * (r - 2)
			c.px(int(cx+math.Cos(a)*d), int(float64(y)+r+2+math.Sin(a)*d), st.Leaf[c.rng.Intn(3)])
		}
	}
}

func (c *canvas) lamp(x, y int) {
	T, st := c.st.T, c.st
	cx := x + T/2
	c.rect(cx-1, y+T/4, cx, y+T-1, st.Outline)
	c.rect(cx-T/6, y+T-1, cx+T/6, y+T-1, st.Outline)
	c.rect(cx-T/5, y, cx+T/5, y+T/4, st.Outline)
	c.rect(cx-T/5+1, y+1, cx+T/5-1, y+T/4-1, st.Glow)
	if st.Detail {
		// Soft halo.
		for i := 0; i < T; i++ {
			a := c.rng.Float64() * 2 * math.Pi
			d := float64(T) * (0.3 + c.rng.Float64()*0.5)
			px, py := int(float64(cx)+math.Cos(a)*d), int(float64(y+T/8)+math.Sin(a)*d*0.6)
			c.px(px, py, rgb(130, 140, 90))
		}
	}
}

// wizard draws the character with proportions from the style: a tall
// robe, a head sized by HeadFrac, a pointed hat and a staff.
func (c *canvas) wizard(fx, fy float64) {
	st := c.st
	H := float64(st.T) * st.CharH
	head := H * st.HeadFrac
	hatH := H * 0.28
	bodyTop := fy - H + head + hatH*0.6
	// Shadow.
	c.disc(fx, fy-1, H*0.14, st.Grass[1])
	// Robe: trapezoid, narrow at the shoulders.
	shoulder := head * 0.9
	hem := head * 1.3
	for y := int(bodyTop); y < int(fy); y++ {
		t := (float64(y) - bodyTop) / (fy - bodyTop)
		hw := shoulder + (hem-shoulder)*t
		x0, x1 := int(fx-hw/2), int(fx+hw/2)
		c.rect(x0-1, y, x1+1, y, st.Outline)
		c.rect(x0, y, x1, y, st.Robe[0])
		// Fold shading on the right third and a hem band.
		c.rect(x0+int(hw*0.66), y, x1, y, st.Robe[1])
		if t > 0.9 {
			c.rect(x0, y, x1, y, st.Robe[1])
		}
	}
	// Belt.
	by := int(bodyTop + (fy-bodyTop)*0.35)
	c.rect(int(fx-shoulder/2), by, int(fx+shoulder/2), by, st.Glow)
	// Head.
	hr := head / 2
	hcy := bodyTop - hr*0.6
	c.disc(fx, hcy, hr+1, st.Outline)
	c.disc(fx, hcy, hr, st.Skin)
	// Beard: lower half of the head and a bit below.
	for y := int(hcy); y < int(hcy+hr*1.4); y++ {
		w := hr * (1.1 - (float64(y)-hcy)/(hr*1.6))
		if w <= 0 {
			break
		}
		c.rect(int(fx-w), y, int(fx+w), y, st.Beard)
	}
	// Eyes.
	ey := int(hcy - hr*0.15)
	c.px(int(fx-hr*0.4), ey, st.Outline)
	c.px(int(fx+hr*0.4), ey, st.Outline)
	// Hat: brim then cone.
	brimY := int(hcy - hr*0.55)
	bw := hr * 1.9
	c.rect(int(fx-bw), brimY-1, int(fx+bw), brimY+1, st.Outline)
	c.rect(int(fx-bw)+1, brimY, int(fx+bw)-1, brimY, st.Hat[0])
	tipY := brimY - int(hatH)
	for y := tipY; y < brimY; y++ {
		t := float64(y-tipY) / float64(brimY-tipY)
		hw := hr * 1.2 * t
		lean := (1 - t) * hr * 0.4 // the tip leans to the side
		x0, x1 := int(fx-hw+lean), int(fx+hw+lean)
		c.rect(x0-1, y, x1+1, y, st.Outline)
		c.rect(x0, y, x1, y, st.Hat[0])
		c.rect(x0, y, x0+int(hw*0.4), y, st.Hat[1])
	}
	// Staff on the right, a gem on top.
	sx := int(fx + hem/2 + 2)
	c.rect(sx, int(fy-H*0.85), sx, int(fy), st.Trunk)
	c.rect(sx-1, int(fy-H*0.85), sx+1, int(fy-H*0.85)+1, st.Outline)
	c.disc(float64(sx)+0.5, fy-H*0.88, math.Max(1.5, H*0.03), rgb(140, 220, 240))
}

func upscale(src *image.RGBA, s int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*s, b.Dy()*s))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := src.RGBAAt(x, y)
			for dy := 0; dy < s; dy++ {
				for dx := 0; dx < s; dx++ {
					dst.SetRGBA(x*s+dx, y*s+dy, c)
				}
			}
		}
	}
	return dst
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
