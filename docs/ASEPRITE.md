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

## Style (decided 2026-10-28: "style B")

Compared three mock-ups (`go run ./tools/mockup`): the current placeholder
look, the same 16x16 with a muted palette and adult proportions, and that
palette at 32x32. Style B won; 32x32 was judged no improvement ("too
crispy"). So:

- 16x16 tiles stay. Characters about 2.5 tiles tall (16x40 cells are fine,
  feet on the bottom row), head roughly a fifth of the height.
- Muted, slightly cool palette; nothing fully saturated. Outlines are a
  soft dark grey `#26222e`, never pure black. Terrain has no outlines at
  all; characters and props do.
- Reference values (the placeholders use these now):
  grass `#5c7648` / dark `#4a603a` / light `#6c8654`;
  stone path `#80766a` / dark `#68605a` / light `#92887c`;
  slate roof `#4e4858` / dark `#3a3442` / light `#605a6c`;
  plaster `#beb29e` / timber `#5a4230`;
  robe `#383c64` / shade `#262846`; skin `#d6b08e`; beard `#cecccf`;
  lamp glow `#f0c878`.
- First real scene to draw: the campus (grass, cobbles, timber-framed
  buildings with slate roofs, trees, lamps) and the wizard. The winter
  village comes later.
