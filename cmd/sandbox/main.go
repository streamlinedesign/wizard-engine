// Command sandbox runs the engine's proving-ground scene.
//
//	go run ./cmd/sandbox
//
// With -screenshot it runs a number of ticks and writes the canvas to a PNG,
// then exits. A window is still created, so on a server run it under
// xvfb-run.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine"
	"github.com/streamlinedesign/wizard-engine/game/scenes"
)

func main() {
	screenshot := flag.String("screenshot", "", "write the canvas to this PNG after -ticks ticks and exit")
	ticks := flag.Uint64("ticks", 90, "ticks to run before taking the screenshot")
	scale := flag.Int("scale", 3, "initial window zoom")
	flag.Parse()

	g := engine.New(engine.Config{
		Title:        "Wizard Engine - sandbox",
		CanvasWidth:  320,
		CanvasHeight: 180,
		WindowScale:  *scale,
	}, &scenes.Sandbox{})
	g.Context().Input.Bindings["debug"] = []ebiten.Key{ebiten.KeyF3}

	if *screenshot != "" {
		g.AfterDraw = func(canvas *ebiten.Image, tick uint64) {
			if tick < *ticks {
				return
			}
			if err := savePNG(*screenshot, canvas); err != nil {
				log.Fatal(err)
			}
			fmt.Println("wrote", *screenshot)
			os.Exit(0)
		}
	}

	if err := g.Run(); err != nil {
		log.Fatal(err)
	}
}

func savePNG(path string, img *ebiten.Image) error {
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	img.ReadPixels(rgba.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, rgba)
}
