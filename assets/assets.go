// Package assets embeds the game's art and data into the binary so a build
// is a single file. Loading helpers decode on demand and cache.
package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"path"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/streamlinedesign/wizard-engine/engine/sprite"
)

//go:embed sprites tiles
var FS embed.FS

var (
	mu     sync.Mutex
	images = map[string]*ebiten.Image{}
)

// Image decodes a PNG from the embedded tree, e.g. "tiles/winter.png".
func Image(name string) (*ebiten.Image, error) {
	mu.Lock()
	defer mu.Unlock()
	if img, ok := images[name]; ok {
		return img, nil
	}
	data, err := FS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("asset %s: %w", name, err)
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("asset %s: %w", name, err)
	}
	img := ebiten.NewImageFromImage(src)
	images[name] = img
	return img, nil
}

// Sprite loads an Aseprite export by its JSON path, e.g. "sprites/wizard.json",
// and the PNG it names alongside.
func Sprite(jsonName string) (*sprite.Sprite, error) {
	data, err := FS.ReadFile(jsonName)
	if err != nil {
		return nil, fmt.Errorf("asset %s: %w", jsonName, err)
	}
	sheet, err := sprite.ParseAseprite(data)
	if err != nil {
		return nil, fmt.Errorf("asset %s: %w", jsonName, err)
	}
	imgName := path.Join(path.Dir(jsonName), path.Base(sheet.Image))
	img, err := Image(imgName)
	if err != nil {
		return nil, err
	}
	return &sprite.Sprite{Sheet: sheet, Image: img}, nil
}

// Must panics on error; for assets that are compiled in and must exist.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
