package main

import (
	"image"
	"image/color"
	"math/rand"
)

// Campus tileset, 16px tiles, 8 columns. Ids are 1-based in the engine;
// the constants in game/scenes/campus.go mirror this table.
//
//	 1 grass          2 grass (worn)     3 grass (flowers)  4 stone path
//	 5 plaza flags    6 water (solid)    7 shore (top edge) 8 wall (solid)
//	 9 wall top       10 roof            11 roof edge       12 door
//	13 window         14 tree top        15 tree bottom (s) 16 hedge (solid)
//	17 lamp (solid)   18 fountain (s)    19 bench (solid)   20 signpost (s)
//	21 bookshelf (s)  22 desk (solid)    23 wood floor      24 rug
func writeCampusTiles(out string) error {
	const ts, cols = 16, 8
	rng := rand.New(rand.NewSource(11))
	img := image.NewRGBA(image.Rect(0, 0, ts*cols, ts*3))

	grass := color.RGBA{92, 118, 72, 255}
	grassDark := color.RGBA{74, 96, 58, 255}
	grassLight := color.RGBA{108, 134, 84, 255}
	flower := color.RGBA{196, 120, 130, 255}
	flower2 := color.RGBA{206, 190, 120, 255}
	stone := color.RGBA{128, 118, 106, 255}
	stoneDark := color.RGBA{104, 96, 86, 255}
	plaza := color.RGBA{156, 148, 138, 255}
	plazaLine := color.RGBA{132, 124, 114, 255}
	water := color.RGBA{52, 84, 112, 255}
	waterLight := color.RGBA{78, 112, 142, 255}
	wall := color.RGBA{190, 178, 158, 255}
	wallDark := color.RGBA{160, 148, 130, 255}
	wallTop := color.RGBA{90, 66, 48, 255}
	roof := color.RGBA{78, 72, 88, 255}
	roofLight := color.RGBA{96, 90, 108, 255}
	roofDark := color.RGBA{58, 52, 66, 255}
	wood := color.RGBA{130, 96, 62, 255}
	woodDark := color.RGBA{92, 66, 42, 255}
	woodLight := color.RGBA{160, 124, 84, 255}
	glass := color.RGBA{120, 150, 170, 255}
	leaf := color.RGBA{62, 96, 62, 255}
	leafDark := color.RGBA{42, 68, 46, 255}
	trunk := color.RGBA{74, 54, 40, 255}
	iron := color.RGBA{60, 58, 70, 255}
	lampGlow := color.RGBA{240, 200, 120, 255}
	outline := color.RGBA{38, 34, 46, 255}
	paper := color.RGBA{236, 226, 200, 255}
	rug := color.RGBA{150, 60, 70, 255}
	rugLine := color.RGBA{200, 150, 90, 255}

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
	rect := func(ox, oy, x0, y0, x1, y1 int, c color.RGBA) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				img.SetRGBA(ox+x, oy+y, c)
			}
		}
	}

	// 1-3 grass
	ox, oy := fill(1, grass)
	speckle(ox, oy, 6, grassDark)
	speckle(ox, oy, 4, grassLight)
	ox, oy = fill(2, grass)
	speckle(ox, oy, 10, grassDark)
	speckle(ox, oy, 6, stoneDark)
	ox, oy = fill(3, grass)
	speckle(ox, oy, 6, grassDark)
	speckle(ox, oy, 3, flower)
	speckle(ox, oy, 3, flower2)
	// 4 stone path
	ox, oy = fill(4, stone)
	speckle(ox, oy, 12, stoneDark)
	speckle(ox, oy, 4, plaza)
	// 5 plaza
	ox, oy = fill(5, plaza)
	for i := 0; i < ts; i++ {
		img.SetRGBA(ox+i, oy+7, plazaLine)
		img.SetRGBA(ox+7, oy+i, plazaLine)
	}
	img.SetRGBA(ox+15, oy+15, plazaLine)
	// 6 water
	ox, oy = fill(6, water)
	for i := 0; i < ts; i += 4 {
		img.SetRGBA(ox+i, oy+3, waterLight)
		img.SetRGBA(ox+i+1, oy+3, waterLight)
		img.SetRGBA(ox+(i+2)%ts, oy+11, waterLight)
		img.SetRGBA(ox+(i+3)%ts, oy+11, waterLight)
	}
	// 7 shore: grass with a stone lip at the bottom, over water
	ox, oy = fill(7, grass)
	speckle(ox, oy, 4, grassDark)
	rect(ox, oy, 0, 12, 15, 13, stoneDark)
	rect(ox, oy, 0, 14, 15, 15, water)
	// 8 wall
	ox, oy = fill(8, wall)
	for y := 0; y < ts; y += 4 {
		for x := 0; x < ts; x++ {
			img.SetRGBA(ox+x, oy+y, wallDark)
		}
		off := (y / 4 % 2) * 4
		for x := off; x < ts; x += 8 {
			for yy := y; yy < y+4 && yy < ts; yy++ {
				img.SetRGBA(ox+x, oy+yy, wallDark)
			}
		}
	}
	// 9 wall top (the dark band where roof meets wall)
	ox, oy = fill(9, wallTop)
	rect(ox, oy, 0, 0, 15, 1, roofDark)
	rect(ox, oy, 0, 14, 15, 15, wallDark)
	// 10 roof
	ox, oy = fill(10, roof)
	for y := 0; y < ts; y += 4 {
		for x := 0; x < ts; x++ {
			img.SetRGBA(ox+x, oy+y, roofDark)
		}
		off := (y / 4 % 2) * 4
		for x := off; x < ts; x += 8 {
			img.SetRGBA(ox+x, oy+y+1, roofLight)
			img.SetRGBA(ox+x+1, oy+y+1, roofLight)
		}
	}
	// 11 roof edge (ridge)
	ox, oy = fill(11, roof)
	rect(ox, oy, 0, 0, 15, 2, roofDark)
	rect(ox, oy, 0, 3, 15, 3, roofLight)
	for y := 4; y < ts; y += 4 {
		for x := 0; x < ts; x++ {
			img.SetRGBA(ox+x, oy+y, roofDark)
		}
	}
	// 12 door (on wall)
	ox, oy = fill(12, wall)
	rect(ox, oy, 3, 2, 12, 15, outline)
	rect(ox, oy, 4, 3, 11, 15, wood)
	rect(ox, oy, 7, 3, 8, 15, woodDark)
	img.SetRGBA(ox+10, oy+9, lampGlow)
	// 13 window (on wall)
	ox, oy = fill(13, wall)
	rect(ox, oy, 4, 3, 11, 10, outline)
	rect(ox, oy, 5, 4, 10, 9, glass)
	rect(ox, oy, 7, 4, 8, 9, outline)
	rect(ox, oy, 5, 6, 10, 6, outline)
	rect(ox, oy, 3, 11, 12, 11, woodDark)
	// 14 tree top, 15 tree bottom
	ox, oy = fill(14, color.RGBA{})
	for y := 2; y < ts; y++ {
		half := 2 + y/2
		if half > 7 {
			half = 7
		}
		for x := 7 - half; x <= 8+half; x++ {
			c := leaf
			if x > 9 || y > 12 {
				c = leafDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	speckle(ox+3, oy+4, 5, grassLight)
	ox, oy = fill(15, color.RGBA{})
	for y := 0; y < 7; y++ {
		half := 7 - y
		for x := 7 - half; x <= 8+half; x++ {
			c := leaf
			if x > 9 {
				c = leafDark
			}
			img.SetRGBA(ox+x, oy+y, c)
		}
	}
	rect(ox, oy, 6, 6, 9, 13, trunk)
	rect(ox, oy, 4, 14, 11, 14, grassDark)
	// 16 hedge
	ox, oy = fill(16, color.RGBA{})
	rect(ox, oy, 0, 3, 15, 14, leaf)
	rect(ox, oy, 0, 11, 15, 14, leafDark)
	rect(ox, oy, 0, 15, 15, 15, grassDark)
	speckle(ox, oy+3, 6, grassLight)
	// 17 lamp
	ox, oy = fill(17, color.RGBA{})
	rect(ox, oy, 7, 4, 7, 15, outline)
	rect(ox, oy, 8, 4, 8, 15, iron)
	rect(ox, oy, 5, 1, 10, 4, outline)
	rect(ox, oy, 6, 2, 9, 3, lampGlow)
	rect(ox, oy, 5, 15, 10, 15, iron)
	// 18 fountain
	ox, oy = fill(18, plaza)
	rect(ox, oy, 1, 1, 14, 14, stoneDark)
	rect(ox, oy, 2, 2, 13, 13, water)
	rect(ox, oy, 6, 4, 9, 11, waterLight)
	rect(ox, oy, 7, 2, 8, 12, stone)
	// 19 bench
	ox, oy = fill(19, color.RGBA{})
	rect(ox, oy, 1, 6, 14, 8, wood)
	rect(ox, oy, 1, 9, 14, 9, woodDark)
	rect(ox, oy, 2, 10, 3, 13, woodDark)
	rect(ox, oy, 12, 10, 13, 13, woodDark)
	// 20 signpost
	ox, oy = fill(20, color.RGBA{})
	rect(ox, oy, 7, 2, 8, 15, woodDark)
	rect(ox, oy, 2, 3, 13, 8, woodLight)
	rect(ox, oy, 2, 3, 13, 3, woodDark)
	rect(ox, oy, 4, 5, 11, 5, woodDark)
	// 21 bookshelf
	ox, oy = fill(21, color.RGBA{})
	rect(ox, oy, 1, 0, 14, 15, woodDark)
	rect(ox, oy, 2, 1, 13, 14, wood)
	for row := 0; row < 3; row++ {
		y := 2 + row*4
		for x := 3; x < 13; x++ {
			c := []color.RGBA{rug, waterLight, flower2, leaf, roofLight}[(x+row)%5]
			rect(ox, oy, x, y, x, y+2, c)
		}
		rect(ox, oy, 2, y+3, 13, y+3, woodDark)
	}
	// 22 desk
	ox, oy = fill(22, color.RGBA{})
	rect(ox, oy, 1, 5, 14, 9, wood)
	rect(ox, oy, 1, 10, 14, 10, woodDark)
	rect(ox, oy, 2, 11, 3, 14, woodDark)
	rect(ox, oy, 12, 11, 13, 14, woodDark)
	rect(ox, oy, 5, 6, 10, 8, paper)
	// 23 wood floor
	ox, oy = fill(23, woodLight)
	for y := 0; y < ts; y += 4 {
		for x := 0; x < ts; x++ {
			img.SetRGBA(ox+x, oy+y, wood)
		}
		off := (y / 4 % 2) * 6
		img.SetRGBA(ox+off, oy+y+2, wood)
		img.SetRGBA(ox+(off+8)%ts, oy+y+2, wood)
	}
	// 24 rug
	ox, oy = fill(24, rug)
	rect(ox, oy, 0, 0, 15, 0, rugLine)
	rect(ox, oy, 0, 15, 15, 15, rugLine)
	rect(ox, oy, 0, 0, 0, 15, rugLine)
	rect(ox, oy, 15, 0, 15, 15, rugLine)
	rect(ox, oy, 6, 6, 9, 9, rugLine)

	return writePNG(out, img)
}

// Item icons, 16x16, 8 columns. Ids are 1-based and named in
// assets/data/items.json.
//
//	1 empty scroll  2 bound scroll  3 fireball   4 small heal
//	5 minor teleport 6 barrier      7 grimoire   8 coin
func writeItemIcons(out string) error {
	const ts, cols = 16, 8
	img := image.NewRGBA(image.Rect(0, 0, ts*cols, ts))

	outline := color.RGBA{38, 34, 46, 255}
	paper := color.RGBA{236, 226, 200, 255}
	paperDark := color.RGBA{200, 186, 150, 255}
	wax := color.RGBA{170, 50, 60, 255}
	fire := color.RGBA{255, 140, 40, 255}
	fireCore := color.RGBA{255, 230, 120, 255}
	heal := color.RGBA{120, 220, 140, 255}
	healLight := color.RGBA{200, 255, 210, 255}
	tele := color.RGBA{150, 110, 230, 255}
	teleLight := color.RGBA{210, 190, 255, 255}
	barrier := color.RGBA{110, 190, 240, 255}
	barrierLight := color.RGBA{200, 235, 255, 255}
	leather := color.RGBA{96, 56, 48, 255}
	leatherLight := color.RGBA{140, 86, 70, 255}
	gold := color.RGBA{240, 200, 90, 255}
	goldDark := color.RGBA{180, 140, 50, 255}

	rect := func(ox, x0, y0, x1, y1 int, c color.RGBA) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				img.SetRGBA(ox+x, y, c)
			}
		}
	}
	scroll := func(ox int, sealed bool) {
		rect(ox, 3, 2, 12, 13, outline)
		rect(ox, 4, 3, 11, 12, paper)
		rect(ox, 4, 3, 11, 3, paperDark)
		rect(ox, 4, 12, 11, 12, paperDark)
		rect(ox, 2, 2, 2, 4, outline)
		rect(ox, 13, 11, 13, 13, outline)
		if sealed {
			rect(ox, 6, 6, 9, 9, wax)
			rect(ox, 7, 7, 8, 8, outline)
		} else {
			rect(ox, 5, 6, 10, 6, paperDark)
			rect(ox, 5, 8, 10, 8, paperDark)
			rect(ox, 5, 10, 8, 10, paperDark)
		}
	}
	orb := func(ox int, dark, light color.RGBA) {
		rect(ox, 5, 3, 10, 12, outline)
		rect(ox, 3, 5, 12, 10, outline)
		rect(ox, 4, 4, 11, 11, outline)
		rect(ox, 5, 4, 10, 11, dark)
		rect(ox, 4, 5, 11, 10, dark)
		rect(ox, 6, 5, 8, 7, light)
	}

	scroll(0*ts, false)
	scroll(1*ts, true)
	orb(2*ts, fire, fireCore)
	rect(2*ts, 8, 1, 9, 3, fire)
	rect(2*ts, 10, 2, 10, 3, fireCore)
	orb(3*ts, heal, healLight)
	rect(3*ts, 7, 5, 8, 10, healLight)
	rect(3*ts, 5, 7, 10, 8, healLight)
	orb(4*ts, tele, teleLight)
	rect(4*ts, 2, 12, 13, 13, outline)
	rect(4*ts, 3, 12, 12, 12, teleLight)
	orb(5*ts, barrier, barrierLight)
	rect(5*ts, 4, 5, 4, 10, barrierLight)
	rect(5*ts, 11, 5, 11, 10, barrierLight)
	// grimoire
	ox := 6 * ts
	rect(ox, 2, 1, 13, 14, outline)
	rect(ox, 3, 2, 12, 13, leather)
	rect(ox, 4, 3, 11, 12, leatherLight)
	rect(ox, 3, 2, 4, 13, leather)
	rect(ox, 6, 5, 9, 10, gold)
	rect(ox, 7, 6, 8, 9, leatherLight)
	// coin
	ox = 7 * ts
	rect(ox, 5, 3, 10, 12, outline)
	rect(ox, 3, 5, 12, 10, outline)
	rect(ox, 4, 4, 11, 11, outline)
	rect(ox, 5, 4, 10, 11, gold)
	rect(ox, 4, 5, 11, 10, gold)
	rect(ox, 7, 6, 8, 9, goldDark)
	rect(ox, 5, 9, 10, 10, goldDark)

	return writePNG(out, img)
}
