// Command mkartpack writes the artist-facing files under artpack/: the
// palette in GIMP .gpl form (which Aseprite loads) and guide sheets that
// show the cell lines at real size.
//
//	go run ./tools/mkartpack
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

type swatch struct {
	name    string
	r, g, b uint8
}

// The style-B palette from docs/ASEPRITE.md, plus a few accents the UI and
// effects already use.
var palette = []swatch{
	{"outline", 38, 34, 46},
	{"shadow", 26, 24, 34},
	{"grass", 92, 118, 72},
	{"grass dark", 74, 96, 58},
	{"grass light", 108, 134, 84},
	{"stone", 128, 118, 106},
	{"stone dark", 104, 96, 86},
	{"stone light", 146, 136, 124},
	{"plaza", 156, 148, 138},
	{"water", 52, 84, 112},
	{"water light", 78, 112, 142},
	{"slate", 78, 72, 88},
	{"slate dark", 58, 52, 66},
	{"slate light", 96, 90, 108},
	{"plaster", 190, 178, 158},
	{"plaster dark", 160, 148, 130},
	{"timber", 90, 66, 48},
	{"wood", 130, 96, 62},
	{"wood light", 160, 124, 84},
	{"leaf", 62, 96, 62},
	{"leaf dark", 42, 68, 46},
	{"leaf light", 84, 118, 76},
	{"trunk", 74, 54, 40},
	{"robe", 56, 60, 100},
	{"robe shade", 38, 40, 70},
	{"hat", 44, 46, 80},
	{"hat light", 66, 70, 112},
	{"skin", 214, 176, 142},
	{"skin shade", 178, 140, 110},
	{"beard", 206, 204, 208},
	{"belt gold", 206, 170, 90},
	{"lamp glow", 240, 200, 120},
	{"parchment", 236, 226, 200},
	{"parchment dark", 200, 186, 150},
	{"ink", 40, 30, 40},
	{"wax red", 170, 50, 60},
	{"fire", 255, 140, 40},
	{"fire core", 255, 230, 120},
	{"life green", 120, 230, 140},
	{"void purple", 160, 110, 240},
	{"ward blue", 110, 190, 240},
	{"thunder yellow", 250, 240, 120},
	{"air", 200, 230, 240},
	{"white", 246, 246, 240},
}

func main() {
	root := "artpack"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	must(os.MkdirAll(filepath.Join(root, "guides"), 0o755))
	must(writeGPL(filepath.Join(root, "palette.gpl")))
	must(writePNG(filepath.Join(root, "guides", "character_cell_16x40.png"), characterGuide()))
	must(writePNG(filepath.Join(root, "guides", "effect_cell_32x32.png"), gridGuide(32, 32, 8, 1, true)))
	must(writePNG(filepath.Join(root, "guides", "tiles_16x16_8cols.png"), gridGuide(16, 16, 8, 4, false)))
	must(writePNG(filepath.Join(root, "guides", "autotile_blob.png"), blobGuide()))
	must(writePNG(filepath.Join(root, "guides", "palette.png"), paletteSheet()))
	fmt.Println("art pack written under", root)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeGPL(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, "GIMP Palette")
	fmt.Fprintln(f, "Name: Wizard Engine style B")
	fmt.Fprintln(f, "Columns: 8")
	fmt.Fprintln(f, "#")
	for _, s := range palette {
		fmt.Fprintf(f, "%3d %3d %3d\t%s\n", s.r, s.g, s.b, s.name)
	}
	return nil
}

var (
	guide     = color.RGBA{255, 0, 180, 140}
	guideSoft = color.RGBA{255, 0, 180, 60}
	ghost     = color.RGBA{90, 90, 120, 110}
)

func set(img *image.RGBA, x, y int, c color.RGBA) {
	if image.Pt(x, y).In(img.Bounds()) {
		img.SetRGBA(x, y, c)
	}
}

// gridGuide draws cell borders; with markCentre a dot marks each centre.
func gridGuide(cw, ch, cols, rows int, markCentre bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, cw*cols, ch*rows))
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x0, y0 := c*cw, r*ch
			for i := 0; i < cw; i++ {
				set(img, x0+i, y0, guide)
				set(img, x0+i, y0+ch-1, guideSoft)
			}
			for i := 0; i < ch; i++ {
				set(img, x0, y0+i, guide)
				set(img, x0+cw-1, y0+i, guideSoft)
			}
			if markCentre {
				set(img, x0+cw/2, y0+ch/2, guide)
				set(img, x0+cw/2-1, y0+ch/2, guide)
				set(img, x0+cw/2, y0+ch/2-1, guide)
				set(img, x0+cw/2-1, y0+ch/2-1, guide)
			}
		}
	}
	return img
}

// characterGuide is a 4-column sheet of 16x40 cells with the feet row,
// the head band and a ghost of the intended proportions.
func characterGuide() *image.RGBA {
	const cw, ch, cols, rows = 16, 40, 4, 3
	img := gridGuide(cw, ch, cols, rows, false)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x0, y0 := c*cw, r*ch
			// Feet row.
			for i := 1; i < cw-1; i++ {
				set(img, x0+i, y0+ch-1, guide)
			}
			// Head band: rows 10..17 (8 px head), hat above.
			for i := 1; i < cw-1; i += 2 {
				set(img, x0+i, y0+10, guideSoft)
				set(img, x0+i, y0+18, guideSoft)
			}
			// Ghost: hat, head, robe to the feet.
			for y := 2; y < 10; y++ {
				hw := (y - 2) / 2
				for x := 8 - hw; x <= 8+hw; x++ {
					set(img, x0+x, y0+y, ghost)
				}
			}
			for y := 10; y < 18; y++ {
				for x := 5; x <= 10; x++ {
					set(img, x0+x, y0+y, ghost)
				}
			}
			for y := 18; y < 39; y++ {
				hw := 3 + (y-18)/7
				for x := 8 - hw; x <= 7+hw; x++ {
					set(img, x0+x, y0+y, ghost)
				}
			}
		}
	}
	return img
}

// blobGuide lays out the 47-tile autotile set as 8 columns x 6 rows, with
// each cell showing which neighbours are "same" as filled edges: the
// standard blob layout, so a tile editor's bitmask matches.
func blobGuide() *image.RGBA {
	const t, cols = 16, 8
	masks := blobMasks()
	rows := (len(masks) + cols - 1) / cols
	img := gridGuide(t, t, cols, rows, false)
	fill := color.RGBA{120, 120, 140, 160}
	for i, m := range masks {
		x0, y0 := (i%cols)*t, (i/cols)*t
		// Centre is always "this tile".
		for y := 5; y < 11; y++ {
			for x := 5; x < 11; x++ {
				set(img, x0+x, y0+y, fill)
			}
		}
		// Edges and corners that continue the same ground.
		edge := func(bit int, x, y, w, h int) {
			if m&bit != 0 {
				for yy := y; yy < y+h; yy++ {
					for xx := x; xx < x+w; xx++ {
						set(img, x0+xx, y0+yy, fill)
					}
				}
			}
		}
		edge(bN, 5, 1, 6, 4)
		edge(bS, 5, 11, 6, 4)
		edge(bW, 1, 5, 4, 6)
		edge(bE, 11, 5, 4, 6)
		edge(bNW, 1, 1, 4, 4)
		edge(bNE, 11, 1, 4, 4)
		edge(bSW, 1, 11, 4, 4)
		edge(bSE, 11, 11, 4, 4)
	}
	return img
}

const (
	bN  = 1
	bNE = 2
	bE  = 4
	bSE = 8
	bS  = 16
	bSW = 32
	bW  = 64
	bNW = 128
)

// blobMasks returns the 47 distinct neighbour masks (a corner only counts
// when both adjacent edges are set), in ascending order.
func blobMasks() []int {
	seen := map[int]bool{}
	var out []int
	for m := 0; m < 256; m++ {
		n := m
		if n&bN == 0 || n&bE == 0 {
			n &^= bNE
		}
		if n&bS == 0 || n&bE == 0 {
			n &^= bSE
		}
		if n&bS == 0 || n&bW == 0 {
			n &^= bSW
		}
		if n&bN == 0 || n&bW == 0 {
			n &^= bNW
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// paletteSheet is the palette as 16x16 swatches, 8 per row, for a quick
// look without opening Aseprite.
func paletteSheet() *image.RGBA {
	const sw, cols = 16, 8
	rows := (len(palette) + cols - 1) / cols
	img := image.NewRGBA(image.Rect(0, 0, sw*cols, sw*rows))
	for i, s := range palette {
		x0, y0 := (i%cols)*sw, (i/cols)*sw
		for y := 0; y < sw; y++ {
			for x := 0; x < sw; x++ {
				img.SetRGBA(x0+x, y0+y, color.RGBA{s.r, s.g, s.b, 255})
			}
		}
	}
	return img
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
