// Command mkplaceholders writes the stand-in art under assets/ so the game
// runs before any real sprites exist. It also writes the Aseprite-style JSON
// for the wizard, so the import pipeline is exercised by real files from day
// one. Real art from Aseprite replaces these files without code changes.
//
//	go run ./tools/mkplaceholders
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
)

var palette = map[byte]color.RGBA{
	'.': {0, 0, 0, 0},
	'O': {24, 20, 37, 255},    // outline
	'H': {58, 68, 140, 255},   // hat/robe blue
	'h': {90, 105, 190, 255},  // hat highlight
	'R': {74, 84, 160, 255},   // robe
	'r': {46, 52, 112, 255},   // robe shade
	'S': {232, 190, 150, 255}, // skin
	'B': {230, 230, 236, 255}, // beard
	'E': {24, 20, 37, 255},    // eye
	'T': {120, 82, 50, 255},   // staff wood
	'G': {120, 220, 255, 255}, // staff gem
	'Y': {240, 200, 90, 255},  // belt / trim
}

// Each template is 16 wide x 28 tall; the frame cell is 16x32 and the body
// sits at the bottom so the feet are on row 31. Legs are the last 4 rows and
// get swapped per walk frame.
var down = []string{
	"................",
	".......OO.......",
	"......OhHO......",
	"......OhHO......",
	".....OhHHHO.....",
	".....OhHHHO.....",
	"....OhHHHHHO....",
	"...OOOOOOOOOO...",
	"..OhHHHHHHHHHO..",
	"...OOOOOOOOOO...",
	"....OSSSSSSO....",
	"....OSESSESO....",
	"....OBSSSSBO..T.",
	"....OBBBBBBO..T.",
	".....OBBBBO...T.",
	"....ORRRRRRO..T.",
	"...ORRRYYRRRO.T.",
	"...ORRRRRRRRO.T.",
	"...ORRRRRRRRO.T.",
	"...ORrRRRRrRO.T.",
	"...ORrRRRRrRO.T.",
	"...OrrRRRRrrO.T.",
	"...OrrrRRrrrO.T.",
	"...OrrrrrrrrO.T.",
}

var up = []string{
	"................",
	".......OO.......",
	"......OhHO......",
	"......OhHO......",
	".....OhHHHO.....",
	".....OhHHHO.....",
	"....OhHHHHHO....",
	"...OOOOOOOOOO...",
	"..OhHHHHHHHHHO..",
	"...OOOOOOOOOO...",
	"....OHHHHHHO....",
	"....OHHHHHHO....",
	"....OHHHHHHO..T.",
	"....OHHHHHHO..T.",
	".....OHHHHO...T.",
	"....ORRRRRRO..T.",
	"...ORRRRRRRRO.T.",
	"...ORRRRRRRRO.T.",
	"...ORRRRRRRRO.T.",
	"...ORrRRRRrRO.T.",
	"...ORrRRRRrRO.T.",
	"...OrrRRRRrrO.T.",
	"...OrrrRRrrrO.T.",
	"...OrrrrrrrrO.T.",
}

var side = []string{
	"................",
	".......OO.......",
	"......OhHO......",
	"......OhHO......",
	".....OhHHHO.....",
	".....OhHHHO.....",
	"....OhHHHHHO....",
	"...OOOOOOOOOO...",
	"..OhHHHHHHHHHO..",
	"...OOOOOOOOOO...",
	".....OSSSSO.....",
	".....OSSSEO.....",
	".....OBSSSO...T.",
	".....OBBBBO...T.",
	"......OBBO....T.",
	".....ORRRRO...T.",
	"....ORRRYRRO..T.",
	"....ORRRRRRO..T.",
	"....ORRRRRRO..T.",
	"....ORrRRRRO..T.",
	"....ORrRRRRO..T.",
	"....OrrRRrrO..T.",
	"....OrrrRrrO..T.",
	"....OrrrrrrO..T.",
}

// legs are the bottom 4 rows for each walk frame (stand, left, stand, right).
var legsWide = [][]string{
	{"...OrOOOOOOrO.T.", "...OrO....OrO.T.", "...OOO....OOO.T.", "................"},
	{"...OrOOOOOOrO.T.", "..OrO.....OrO.T.", "..OOO......OO.T.", "................"},
	{"...OrOOOOOOrO.T.", "...OrO....OrO.T.", "...OOO....OOO.T.", "................"},
	{"...OrOOOOOOrO.T.", "....OrO...OrO.T.", "....OOO....OOOT.", "................"},
}

var legsNarrow = [][]string{
	{"....OrOOOOrO..T.", "....OrO..OrO..T.", "....OOO..OOO..T.", "................"},
	{"....OrOOOOrO..T.", "...OrO....OrO.T.", "...OOO....OOO.T.", "................"},
	{"....OrOOOOrO..T.", "....OrO..OrO..T.", "....OOO..OOO..T.", "................"},
	{"....OrOOOOrO..T.", ".....OrOOrO...T.", ".....OOOOOO...T.", "................"},
}

const (
	cellW, cellH = 16, 32
	walkFrames   = 4
)

type aseFrame struct {
	Filename         string  `json:"filename"`
	Frame            aseRect `json:"frame"`
	Rotated          bool    `json:"rotated"`
	Trimmed          bool    `json:"trimmed"`
	SpriteSourceSize aseRect `json:"spriteSourceSize"`
	SourceSize       aseSize `json:"sourceSize"`
	Duration         int     `json:"duration"`
}
type aseRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}
type aseSize struct {
	W int `json:"w"`
	H int `json:"h"`
}
type aseTag struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

func main() {
	root := "assets"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	must(os.MkdirAll(filepath.Join(root, "sprites"), 0o755))
	must(os.MkdirAll(filepath.Join(root, "tiles"), 0o755))
	must(os.MkdirAll(filepath.Join(root, "icons"), 0o755))
	must(writeWizard(filepath.Join(root, "sprites"), "wizard", palette))
	must(writeWizard(filepath.Join(root, "sprites"), "oldwizard", oldWizardPalette()))
	must(writeWinterTiles(filepath.Join(root, "tiles", "winter.png")))
	must(writeCampusTiles(filepath.Join(root, "tiles", "campus.png")))
	must(writeItemIcons(filepath.Join(root, "icons", "items.png")))
	fmt.Println("placeholders written under", root)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// oldWizardPalette recolours the player template into the teacher: grey
// robe, white hair, a warmer gem.
func oldWizardPalette() map[byte]color.RGBA {
	p := map[byte]color.RGBA{}
	for k, v := range palette {
		p[k] = v
	}
	p['H'] = color.RGBA{96, 86, 110, 255}
	p['h'] = color.RGBA{138, 126, 154, 255}
	p['R'] = color.RGBA{118, 108, 132, 255}
	p['r'] = color.RGBA{78, 70, 90, 255}
	p['Y'] = color.RGBA{200, 160, 90, 255}
	p['G'] = color.RGBA{255, 190, 90, 255}
	return p
}

// writeWizard lays out rows: down (0-3), up (4-7), side (8-11).
func writeWizard(dir, name string, pal map[byte]color.RGBA) error {
	rows := []struct {
		name string
		body []string
		legs [][]string
	}{
		{"down", down, legsWide},
		{"up", up, legsWide},
		{"side", side, legsNarrow},
	}
	img := image.NewRGBA(image.Rect(0, 0, cellW*walkFrames, cellH*len(rows)))
	var frames []aseFrame
	var tags []aseTag
	for r, row := range rows {
		for f := 0; f < walkFrames; f++ {
			// Bob the body up one pixel on the stepping frames.
			bob := 0
			if f%2 == 1 {
				bob = 1
			}
			ox, oy := f*cellW, r*cellH
			top := cellH - len(row.body) - len(row.legs[f]) - bob
			blit(img, ox, oy+top, row.body, pal)
			blit(img, ox, oy+cellH-len(row.legs[f]), row.legs[f], pal)
			idx := len(frames)
			frames = append(frames, aseFrame{
				Filename:         fmt.Sprintf("%s %d.aseprite", name, idx),
				Frame:            aseRect{ox, oy, cellW, cellH},
				SpriteSourceSize: aseRect{0, 0, cellW, cellH},
				SourceSize:       aseSize{cellW, cellH},
				Duration:         140,
			})
		}
		base := r * walkFrames
		tags = append(tags,
			aseTag{"idle_" + row.name, base, base, "forward"},
			aseTag{"walk_" + row.name, base, base + walkFrames - 1, "forward"},
		)
	}
	if err := writePNG(filepath.Join(dir, name+".png"), img); err != nil {
		return err
	}
	doc := map[string]any{
		"frames": frames,
		"meta": map[string]any{
			"app":       "mkplaceholders (Aseprite-compatible)",
			"image":     name + ".png",
			"format":    "RGBA8888",
			"size":      aseSize{img.Bounds().Dx(), img.Bounds().Dy()},
			"scale":     "1",
			"frameTags": tags,
		},
	}
	data, err := json.MarshalIndent(doc, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".json"), data, 0o644)
}

func blit(img *image.RGBA, ox, oy int, rows []string, pal map[byte]color.RGBA) {
	for y, row := range rows {
		for x := 0; x < len(row); x++ {
			c, ok := pal[row[x]]
			if !ok || c.A == 0 {
				continue
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
}

// Winter tileset, 16px tiles, 8 columns. Ids are 1-based in the engine:
//
//	1 snow         2 snow (speckled)  3 packed path   4 ice
//	5 water (solid) 6 rock (solid)    7 pine top      8 pine bottom (solid)
//	9 fence (solid) 10 snow edge top  11 shrub (solid) 12 lamp (solid)
func writeWinterTiles(out string) error {
	const ts, cols = 16, 8
	rng := rand.New(rand.NewSource(7))
	img := image.NewRGBA(image.Rect(0, 0, ts*cols, ts*2))

	snow := color.RGBA{228, 236, 246, 255}
	snowShade := color.RGBA{206, 218, 236, 255}
	snowBright := color.RGBA{246, 250, 255, 255}
	path := color.RGBA{178, 170, 160, 255}
	pathDark := color.RGBA{150, 142, 132, 255}
	ice := color.RGBA{176, 214, 236, 255}
	iceLight := color.RGBA{214, 236, 248, 255}
	water := color.RGBA{54, 96, 150, 255}
	waterLight := color.RGBA{82, 130, 184, 255}
	rock := color.RGBA{110, 112, 124, 255}
	rockDark := color.RGBA{72, 74, 86, 255}
	pine := color.RGBA{44, 92, 70, 255}
	pineDark := color.RGBA{28, 62, 48, 255}
	trunk := color.RGBA{88, 60, 40, 255}
	wood := color.RGBA{130, 96, 62, 255}
	woodDark := color.RGBA{92, 66, 42, 255}
	lampGlow := color.RGBA{255, 214, 120, 255}
	outline := color.RGBA{24, 20, 37, 255}

	fill := func(id int, base color.RGBA) (ox, oy int) {
		i := id - 1
		ox, oy = (i%cols)*ts, (i/cols)*ts
		for y := 0; y < ts; y++ {
			for x := 0; x < ts; x++ {
				img.SetRGBA(ox+x, oy+y, base)
			}
		}
		return
	}
	speckle := func(ox, oy, n int, c color.RGBA) {
		for i := 0; i < n; i++ {
			img.SetRGBA(ox+rng.Intn(ts), oy+rng.Intn(ts), c)
		}
	}

	// 1 snow
	ox, oy := fill(1, snow)
	speckle(ox, oy, 3, snowBright)
	// 2 speckled snow
	ox, oy = fill(2, snow)
	speckle(ox, oy, 6, snowShade)
	speckle(ox, oy, 3, snowBright)
	// 3 path
	ox, oy = fill(3, path)
	speckle(ox, oy, 10, pathDark)
	speckle(ox, oy, 4, snowShade)
	// 4 ice
	ox, oy = fill(4, ice)
	for i := 0; i < ts; i++ {
		img.SetRGBA(ox+i, oy+(i+5)%ts, iceLight)
	}
	speckle(ox, oy, 4, iceLight)
	// 5 water
	ox, oy = fill(5, water)
	for i := 0; i < ts; i += 4 {
		img.SetRGBA(ox+i, oy+3, waterLight)
		img.SetRGBA(ox+i+1, oy+3, waterLight)
		img.SetRGBA(ox+(i+2)%ts, oy+11, waterLight)
		img.SetRGBA(ox+(i+3)%ts, oy+11, waterLight)
	}
	// 6 rock
	ox, oy = fill(6, snow)
	for y := 4; y < 15; y++ {
		for x := 2; x < 14; x++ {
			if (x == 2 || x == 13) && (y == 4 || y == 14) {
				continue
			}
			c := rock
			if y > 10 || x < 4 {
				c = rockDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	for x := 3; x < 13; x++ {
		img.SetRGBA(ox+x, oy+15, outline)
		img.SetRGBA(ox+x, oy+5, snowBright)
	}
	// 7 pine top (transparent background so it layers over ground)
	ox, oy = fill(7, color.RGBA{})
	for y := 1; y < 16; y++ {
		half := y / 2
		if y%3 == 0 {
			half++
		}
		for x := 7 - half; x <= 8+half; x++ {
			if x < 0 || x > 15 {
				continue
			}
			c := pine
			if x > 8 {
				c = pineDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	for y := 2; y < 16; y += 3 {
		img.SetRGBA(ox+7-y/3, oy+y, snowBright)
	}
	// 8 pine bottom
	ox, oy = fill(8, color.RGBA{})
	for y := 0; y < 9; y++ {
		half := 5 + y/3
		for x := 7 - half; x <= 8+half; x++ {
			if x < 0 || x > 15 {
				continue
			}
			c := pine
			if x > 8 {
				c = pineDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	for y := 9; y < 14; y++ {
		img.SetRGBA(ox+7, oy+y, trunk)
		img.SetRGBA(ox+8, oy+y, trunk)
	}
	for x := 4; x < 12; x++ {
		img.SetRGBA(ox+x, oy+14, snowShade)
	}
	// 9 fence
	ox, oy = fill(9, color.RGBA{})
	for x := 0; x < ts; x++ {
		img.SetRGBA(ox+x, oy+6, wood)
		img.SetRGBA(ox+x, oy+7, woodDark)
		img.SetRGBA(ox+x, oy+11, wood)
		img.SetRGBA(ox+x, oy+12, woodDark)
	}
	for _, px := range []int{2, 12} {
		for y := 3; y < 15; y++ {
			img.SetRGBA(ox+px, oy+y, wood)
			img.SetRGBA(ox+px+1, oy+y, woodDark)
		}
		img.SetRGBA(ox+px, oy+2, snowBright)
		img.SetRGBA(ox+px+1, oy+2, snowBright)
	}
	// 10 snow edge (darker band along the top, for pond rims)
	ox, oy = fill(10, snow)
	for x := 0; x < ts; x++ {
		img.SetRGBA(ox+x, oy, snowShade)
		img.SetRGBA(ox+x, oy+1, snowShade)
	}
	// 11 shrub
	ox, oy = fill(11, color.RGBA{})
	for y := 5; y < 15; y++ {
		for x := 3; x < 13; x++ {
			if (x == 3 || x == 12) && (y == 5 || y == 14) {
				continue
			}
			c := pine
			if y > 10 {
				c = pineDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	speckle(ox+3, oy+5, 5, snowBright)
	// 12 lamp post
	ox, oy = fill(12, color.RGBA{})
	for y := 4; y < 16; y++ {
		img.SetRGBA(ox+7, oy+y, outline)
		img.SetRGBA(ox+8, oy+y, rockDark)
	}
	for y := 1; y < 5; y++ {
		for x := 5; x < 11; x++ {
			c := lampGlow
			if y == 1 || y == 4 || x == 5 || x == 10 {
				c = outline
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	for x := 5; x < 11; x++ {
		img.SetRGBA(ox+x, oy+15, rockDark)
	}

	return writePNG(out, img)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
