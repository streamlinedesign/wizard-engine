# Art spec and Aseprite export

This is the contract between the art and the engine. If a file follows it,
dropping it into `assets/` is all that is needed; no code changes.

## Sizes

- Tiles: 16x16. One PNG per tileset, tiles packed left-to-right,
  top-to-bottom, any number of columns. Tile 1 is the top-left cell (0 means
  "nothing"). Keep ground tiles opaque and prop tiles (trees, fences, lamps)
  transparent where the ground should show through.
- Characters: 16x32 cells, feet on the bottom row (row 31), centred
  horizontally. Hats and staffs can use the full height. The engine's
  collision box is about 10x6 pixels at the feet, so a character can stand
  closer to things than its sprite suggests.
- Larger things (bosses, ships) can be any multiple; the engine reads the
  size from the export.

## Animation tags for characters

The engine looks these up by name:

| tag          | frames | note                                   |
|--------------|--------|----------------------------------------|
| `idle_down`  | 1+     | facing the camera                      |
| `walk_down`  | 4      |                                        |
| `idle_up`    | 1+     | back to the camera                     |
| `walk_up`    | 4      |                                        |
| `idle_side`  | 1+     | facing **right**; the engine flips it  |
| `walk_side`  | 4      |                                        |

Frame durations come from Aseprite (140 ms per walk frame feels right at
this scale). Tag direction (forward / reverse / ping-pong) is honoured.
Spell casting, hurt and so on will be added to this table when they exist.

## Export settings

File > Export Sprite Sheet (or the command line):

```
aseprite -b wizard.aseprite \
  --sheet assets/sprites/wizard.png \
  --data  assets/sprites/wizard.json \
  --format json-array --list-tags --sheet-type packed
```

- JSON array or hash both work. Trimmed frames work. Rotation is not
  supported, so leave "Allow rotation" off.
- The JSON's `meta.image` must be the PNG filename next to it.
- Keep the PNG and JSON together under `assets/sprites/`.

## Style notes (agreed so far)

- Top-down, 16x16 tiles, taller characters. Aiming for a cozy, slightly
  moody look rather than bright and chunky: a restrained palette, dark
  outlines on characters, softer edges on terrain.
- First real scene to draw: a small winter village hub (snow, pines, a
  frozen pond, lamps, a fence, the old wizard's house).
