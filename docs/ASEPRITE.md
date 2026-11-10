# Art spec and Aseprite export

This is the contract between the art and the engine. If a file follows it,
dropping it into `assets/` is all that is needed; no code changes.

## Sizes

- Tiles: 16x16. One PNG per tileset, tiles packed left-to-right,
  top-to-bottom, any number of columns. Tile 1 is the top-left cell (0 means
  "nothing"). Keep ground tiles opaque and prop tiles (trees, fences, lamps)
  transparent where the ground should show through.
- Characters: 16x40 cells (the player wizard; the old wizard is still a
  16x32 placeholder), feet on the bottom row, centred horizontally. Hats and staffs can use the full height. The engine's
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
| `cast_down`, `cast_up`, `cast_side` | 4 | staff raised, release on frame 2; played once on every cast |
| `hurt`       | 2      | not played yet                         |
| `meditate`   | 4      | loop, sitting; not played yet          |

Frame durations come from Aseprite (140 ms per walk frame feels right at
this scale). Tag direction (forward / reverse / ping-pong) is honoured.
One-shot tags (cast, hurt, effects that play once) set Aseprite's repeat
to 1 so the engine knows when they are done.

## Spell effects

`assets/effects/<name>.png` + `.json`, same Aseprite export, drawn with the
centre of the cell on the point the game gives. 32x32 cells unless noted.

| file | tags | used for |
|---|---|---|
| `fireball` | `fly` (4, loop, points right; the engine rotates it), `hit` (6) | fire projectiles and their impact |
| `heal` | `self` (8) | minor heal, on the caster |
| `thunder` | `strike` (6) | thunderstrike; impact at the bottom centre, row 27. The engine draws the bolt above the cell |
| `gust` | `self` (6) | gust of wind, on the caster |
| `barrier` | `appear` (3), `hold` (4, loop), `break` (3); 48x48 | barrier bubble (`break` not played yet) |
| `teleport` | `out` (5), `in` (5) | minor teleport, at both ends |
| `hit` | `hit` (3) | impact of any projectile without its own art |
| `puff` | `puff` (4); 16x16 | footstep dust (not played yet) |

Elements other than fire still draw their projectile as a circle until a
recolour of the fireball exists.

## Where the current art comes from

The wizard, the effects and the campus tileset are drawn by the scripts in
`art/gen/` (Python + Pillow, every pixel placed from the palette below).
`make art` redraws them into `assets/` and the previews into `docs/art/`.
An Aseprite export of the same name replaces any of them; the scripts are
then retired for that file.

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
