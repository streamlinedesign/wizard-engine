// Package input maps physical keys to named game actions, so the rest of the
// engine never asks "is W down" but "is MoveUp down". That keeps rebinding,
// gamepads and later scripted/network input in one place.
package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Action is a named thing the player can do.
type Action string

const (
	MoveUp    Action = "move_up"
	MoveDown  Action = "move_down"
	MoveLeft  Action = "move_left"
	MoveRight Action = "move_right"
	Interact  Action = "interact"
	Cast      Action = "cast"
	Menu      Action = "menu"
	Inventory Action = "inventory"
	Grimoire  Action = "grimoire"
	Slot1     Action = "slot1"
	Slot2     Action = "slot2"
	Slot3     Action = "slot3"
	Slot4     Action = "slot4"
)

// Bindings maps each action to the keys that trigger it.
type Bindings map[Action][]ebiten.Key

// DefaultBindings is WASD + arrows, E to interact, Space to cast, Esc for menu.
func DefaultBindings() Bindings {
	return Bindings{
		MoveUp:    {ebiten.KeyW, ebiten.KeyArrowUp},
		MoveDown:  {ebiten.KeyS, ebiten.KeyArrowDown},
		MoveLeft:  {ebiten.KeyA, ebiten.KeyArrowLeft},
		MoveRight: {ebiten.KeyD, ebiten.KeyArrowRight},
		Interact:  {ebiten.KeyE, ebiten.KeyEnter},
		Cast:      {ebiten.KeySpace},
		Menu:      {ebiten.KeyEscape},
		Inventory: {ebiten.KeyTab, ebiten.KeyI},
		Grimoire:  {ebiten.KeyG},
		Slot1:     {ebiten.KeyDigit1},
		Slot2:     {ebiten.KeyDigit2},
		Slot3:     {ebiten.KeyDigit3},
		Slot4:     {ebiten.KeyDigit4},
	}
}

// State answers questions about actions for the current tick.
type State struct {
	Bindings Bindings
}

// New returns an input state with the default bindings.
func New() *State {
	return &State{Bindings: DefaultBindings()}
}

// Down reports whether any key bound to the action is held.
func (s *State) Down(a Action) bool {
	for _, k := range s.Bindings[a] {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

// JustPressed reports whether a key bound to the action went down this tick.
func (s *State) JustPressed(a Action) bool {
	for _, k := range s.Bindings[a] {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

// Clicked reports whether the left mouse button went down this tick.
func (s *State) Clicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// RightClicked reports whether the right mouse button went down this tick.
func (s *State) RightClicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
}

// Axis returns the movement direction as a unit vector (or zero). Diagonals
// are normalised so moving diagonally is not faster than straight.
func (s *State) Axis() (dx, dy float64) {
	if s.Down(MoveLeft) {
		dx--
	}
	if s.Down(MoveRight) {
		dx++
	}
	if s.Down(MoveUp) {
		dy--
	}
	if s.Down(MoveDown) {
		dy++
	}
	return Normalize(dx, dy)
}

// Normalize scales a vector to length 1, leaving the zero vector alone.
func Normalize(dx, dy float64) (float64, float64) {
	l := math.Hypot(dx, dy)
	if l == 0 {
		return 0, 0
	}
	return dx / l, dy / l
}
