package main

import (
	"image"
	"image/color"
)

// Item icons, 16x16, 8 columns. Ids are 1-based and named in
// assets/data/items.json.
//
//	1 empty scroll  2 bound scroll  3 fireball   4 small heal
//	5 minor teleport 6 barrier      7 grimoire   8 coin
func writeItemIcons(out string) error {
	const ts, cols = 16, 10
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
	// thunderstrike: a bolt
	ox = 8 * ts
	bolt := color.RGBA{250, 240, 120, 255}
	boltDark := color.RGBA{200, 170, 60, 255}
	for _, p := range [][4]int{{8, 1, 10, 3}, {7, 4, 9, 6}, {5, 7, 10, 8}, {7, 9, 8, 11}, {6, 12, 7, 14}} {
		rect(ox, p[0]-1, p[1], p[2]+1, p[3], outline)
	}
	for _, p := range [][4]int{{8, 1, 10, 3}, {7, 4, 9, 6}, {5, 7, 10, 8}, {7, 9, 8, 11}, {6, 12, 7, 14}} {
		rect(ox, p[0], p[1], p[2], p[3], bolt)
	}
	rect(ox, 9, 2, 10, 3, boltDark)
	rect(ox, 9, 7, 10, 8, boltDark)
	// gust of wind: three curls
	ox = 9 * ts
	air := color.RGBA{200, 230, 240, 255}
	airDark := color.RGBA{140, 180, 200, 255}
	for i, y := range []int{4, 8, 12} {
		x0 := 2 + i
		rect(ox, x0, y, x0+8, y, airDark)
		rect(ox, x0, y-1, x0+8, y-1, air)
		rect(ox, x0+9, y-2, x0+10, y-1, air)
		rect(ox, x0+10, y-3, x0+10, y-2, airDark)
	}

	return writePNG(out, img)
}
