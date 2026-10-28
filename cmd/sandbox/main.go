// Command sandbox runs the game's scenes for development.
//
//	go run ./cmd/sandbox            # the campus hub
//	go run ./cmd/sandbox -scene snow # the step-1 winter clearing
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
	scene := flag.String("scene", "campus", "scene to start in: campus or snow")
	at := flag.String("at", "", "start the player at tile x,y (campus only), e.g. -at 33,26")
	say := flag.String("say", "", "open this dialogue id on start, to preview text (campus only)")
	inv := flag.Bool("inventory", false, "start with the old wizard's gifts and the inventory open (campus only)")
	flag.Parse()

	var start engine.Scene
	switch *scene {
	case "campus":
		c := &scenes.Campus{StartDialogue: *say, StartInventory: *inv}
		if *at != "" {
			if _, err := fmt.Sscanf(*at, "%d,%d", &c.StartX, &c.StartY); err != nil {
				log.Fatalf("-at wants x,y tiles: %v", err)
			}
		}
		start = c
	case "snow":
		start = &scenes.Sandbox{}
	default:
		log.Fatalf("unknown scene %q", *scene)
	}

	g := engine.New(engine.Config{
		Title:        "Wizard Engine - " + *scene,
		CanvasWidth:  320,
		CanvasHeight: 180,
		WindowScale:  *scale,
	}, start)
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
