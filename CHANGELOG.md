# Changelog

All notable changes, newest first. `make release` turns Unreleased into a version.

## [Unreleased]

- First real art: the player wizard (16x40, walk, idle, cast, hurt,
  meditate), spell effects (fireball, impacts, heal, thunderstrike, gust,
  barrier, teleport, dust) and the campus tileset redrawn in style B with
  eight new tiles. The game plays the effects and the cast pose. Previews
  in `docs/art/`; `make art` redraws everything.
- House conventions: `CLAUDE.md`, `Makefile`, `VERSION`, `TODO.md`, `.env.example`, CI and auto-PR workflows. No game changes.

## [0.1.0] - 2026-11-01

- Everything so far: the engine (fixed-step loop, scenes, pixel-perfect
  viewport and camera, input actions, Aseprite sprites, tilemaps, ui), the
  campus hub with the old wizard, dialogue and inventory (step 2), the ritual
  circle binding, grimoire, casting with all ten spell stats and training
  dummies (step 3), placeholder art and the art-style mockup tool, and the
  design notes in `docs/DESIGN.md`.
