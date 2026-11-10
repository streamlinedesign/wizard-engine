"""The player wizard, 16x40 cells, feet on row 39. Run: python3 wizard.py OUTDIR"""
import sys, os
from PIL import Image
from pal import PAL, grid, aseprite_json, zoom

W, H = 16, 40

HAT = [
    "................",
    "..........OO....",
    ".........OhHO...",
    ".........OhO....",
    "........OhHO....",
    "........OhHO....",
    ".......OhHHO....",
    ".......OhHHO....",
    "......OhHHHO....",
    ".....OhHHHHO....",
    "....OhHHHHHO....",
    "....OYYYYYYO....",
    ".OhhHHHHHHHHHHO.",
    "..OOOOOOOOOOOO..",
]

DOWN = HAT + [
    "....OssssssO....",
    "....OSESSESO....",
    "....OSSssSSO....",
    "....OBBSSBBO....",
    "..ORBBBBBBBBRO..",
    ".ORRBBBBBBBBRRO.",
    ".ORRbBBBBBBbRRO.",
    ".ORRRbBBBBbRRRO.",
    ".ORRRRbBBbRRRRO.",
    ".OrRRRRbbRRRRrO.",
    ".OrRYYYyyYYYRrO.",
    ".OrRRRRRRRRRRrO.",
    ".OSrRRRRRRRRrSO.",
    ".OrRRRRRRRRRRrO.",
    ".OrRRRRRRRRRRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OrrRRRrRRRRrrO.",
    ".OrrRRRrRRRRrrO.",
    "OrrRRRRrRRRRRrrO",
    "OrrRRRRrRRRRRrrO",
    "OrrrRRRrRRRRrrrO",
    "OOOOOOOOOOOOOOOO",
]

UP = HAT + [
    "....OBBBBBBO....",
    "....OBBBBBBO....",
    "....ObBBBBbO....",
    ".....ObbbbO.....",
    "..ORRRRRRRRRRO..",
    ".ORRRRRRRRRRRRO.",
    ".ORRRRRRRRRRRRO.",
    ".ORRRRRrRRRRRRO.",
    ".ORRRRRrRRRRRRO.",
    ".OrRRRRrRRRRRrO.",
    ".OrRYYYyyYYYRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OSrRRRrRRRRrSO.",
    ".OrRRRRrRRRRRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OrRRRRrRRRRRrO.",
    ".OrrRRRrRRRRrrO.",
    ".OrrRRRrRRRRrrO.",
    "OrrRRRRrRRRRRrrO",
    "OrrRRRRrRRRRRrrO",
    "OrrrRRRrRRRRrrrO",
    "OOOOOOOOOOOOOOOO",
]

SIDE_HAT = [
    "................",
    "...OO...........",
    "...OhO..........",
    "....OhO.........",
    "....OhHO........",
    ".....OhHO.......",
    ".....OhHHO......",
    ".....OhHHHO.....",
    ".....OhHHHO.....",
    "....OhHHHHHO....",
    "....OhHHHHHO....",
    "....OYYYYYYO....",
    "..OhhHHHHHHHHHO.",
    "...OOOOOOOOOOOO.",
]

SIDE = SIDE_HAT + [
    ".....OBBssssO...",
    ".....OBBSSESO...",
    ".....OBSSSSSSO..",
    ".....OBSSBBBO...",
    "....ORBBBBBBO...",
    "....ORRBBBBBBO..",
    "....ORRbBBBBO...",
    "....ORRRbBBBO...",
    "....OrRRRbbO....",
    "....OrRRRRRO....",
    "....OrYYYYYYO...",
    "....OrRRrRRO....",
    "....OrRRrRRSO...",
    "....OrRRRrRO....",
    "....OrRRRRRO....",
    "....OrrRRRRO....",
    "...OrrRRRRRO....",
    "...OrrRRRRRRO...",
    "...OrrRRRRRRO...",
    "..OrrrRRRRRRRO..",
    "..OrrrRRRRRRRO..",
    "..OrrrrRRRRRrO..",
    "..OOOOOOOOOOOO..",
]

def lit(t, a, b):
    """Light comes from the top left: the first shade pixel of a robe row becomes highlight."""
    t = list(t)
    for i in range(a, b):
        r = t[i]
        k = r.find('r')
        if 0 <= k <= 3:
            t[i] = r[:k] + 'h' + r[k + 1:]
    return t

DOWN = lit(DOWN, 23, 36)
UP = lit(UP, 23, 36)

MEDITATE = HAT + [
    "....OssssssO....",
    "....OSsSSsSO....",
    "....OSSssSSO....",
    "....OBBSSBBO....",
    "..ORBBBBBBBBRO..",
    ".ORRBBBBBBBBRRO.",
    ".OhRbBBBBBBbRrO.",
    ".OhRRbBBBBbRRrO.",
    "OhRRRRbBBbRRRRrO",
    "OhRYYYYyyYYYYRrO",
    "OSsRRRRRRRRRRsSO",
    "OhrRRRRrRRRRRrrO",
    "OhRRRRRrRRRRRRrO",
    "OKtrRRRrRRRRrtKO",
    "OOOOOOOOOOOOOOOO",
]
assert len(MEDITATE) == 29

for name, t in [('DOWN', DOWN), ('UP', UP), ('SIDE', SIDE), ('MEDITATE', MEDITATE)]:
    assert name == 'MEDITATE' or len(t) == 37, (name, len(t))
    for i, r in enumerate(t):
        assert len(r) == 16, (name, i, r)

# Feet below the hem: rows 37-39 (row 37 only shows when the body bobs up).
FEET_DOWN = [
    ["...OKKO..OKKO...", "...OKKO..OKKO...", "...OttO..OttO...", "...OOOO..OOOO..."],
    ["..OKKO....OKKO..", "..OKKO....OOOO..", "..OttO..........", "..OOOO.........."],  # left forward
    ["...OKKO..OKKO...", "...OKKO..OKKO...", "...OttO..OttO...", "...OOOO..OOOO..."],
    ["..OKKO....OKKO..", "..OOOO....OKKO..", "..........OttO..", "..........OOOO.."],  # right forward
]
FEET_SIDE = [
    [".....OKKKKO.....", ".....OKKKKO.....", ".....OttOttO....", ".....OOOOOOO...."],
    ["...OKKO.OKKO....", "...OKKO..OKKO...", "...OttO..OttO...", "...OOOO..OOOO..."],
    [".....OKKKKO.....", ".....OKKKKO.....", ".....OttOttO....", ".....OOOOOOO...."],
    ["....OKKOKKO.....", "....OKKOOKKO....", "....OttO.OttO...", "....OOOO.OOOO..."],
]

def staff(img, x, top, glow=0, side=False):
    """A wooden staff, gem on top. x is the shaft column, top the gem's top row."""
    for y in range(top + 4, H):
        img.putpixel((x, y), PAL['T'])
        if 0 <= x + 1 < W:
            img.putpixel((x + 1, y), PAL['t'])
    # crook around the gem
    gem = [
        ".OO.",
        "OgGO",
        "OGGO",
        ".tT.",
    ]
    grid(img, gem, x - 1, top)
    if glow:
        halo = {1: [(x - 2, top + 1), (x + 3, top + 1), (x, top - 1), (x + 1, top - 1)],
                2: [(x - 2, top + 1), (x + 3, top + 1), (x, top - 1), (x + 1, top - 1),
                    (x - 2, top + 2), (x + 3, top + 2), (x - 3, top + 1), (x + 4, top + 2),
                    (x + 1, top - 2), (x, top - 2)]}[glow]
        for p in halo:
            if 0 <= p[0] < W and 0 <= p[1] < H:
                img.putpixel(p, PAL['g'])
        if glow == 2:
            for p in [(x, top + 1), (x + 1, top + 1)]:
                img.putpixel(p, PAL['W'])

def hand(img, x, y):
    img.putpixel((x, y), PAL['S'])
    img.putpixel((x, y + 1), PAL['s'])

def body(img, tmpl, feet, bob, sway=0):
    rows = list(tmpl)
    if sway:
        # hem sways: shift the last four hem rows
        for i in range(33, 37):
            r = rows[i]
            r = (r[1:] + '.') if sway < 0 else ('.' + r[:-1])
            r = list(r)
            xs = [k for k, c in enumerate(r) if c != '.']
            r[xs[0]] = 'O'; r[xs[-1]] = 'O'
            rows[i] = ''.join(r)
    grid(img, feet, 0, 36)
    grid(img, rows, 0, -bob)
    if bob:
        # fill the gap between hem and feet with shins
        pass

def frame_down(i, staff_lift=0, glow=0):
    img = Image.new('RGBA', (W, H))
    bob = 1 if i in (1, 3) else 0
    body(img, DOWN, FEET_DOWN[i], bob, {1: -1, 3: 1}.get(i, 0))
    staff(img, 1, 6 - bob - staff_lift, glow)
    hand(img, 2, 26 - bob - staff_lift)
    return img

def frame_up(i, staff_lift=0, glow=0):
    img = Image.new('RGBA', (W, H))
    bob = 1 if i in (1, 3) else 0
    # draw staff behind? when facing away the staff is in front of the body edge
    body(img, UP, FEET_DOWN[i], bob, {1: -1, 3: 1}.get(i, 0))
    staff(img, 13, 6 - bob - staff_lift, glow)
    hand(img, 13, 26 - bob - staff_lift)
    return img

def frame_side(i, staff_lift=0, glow=0, reach=0):
    img = Image.new('RGBA', (W, H))
    bob = 1 if i in (1, 3) else 0
    sx = 13 + reach
    staff(img, sx, 6 - bob - staff_lift, glow)
    body(img, SIDE, FEET_SIDE[i], bob)
    hand(img, sx - 1, 26 - bob - staff_lift)
    return img

def frame_hurt(i):
    img = Image.new('RGBA', (W, H))
    if i == 0:
        # white flash: everything but the outline goes pale
        f = frame_down(0)
        for y in range(H):
            for x in range(W):
                p = f.getpixel((x, y))
                if p[3] and p[:3] != PAL['O'][:3]:
                    f.putpixel((x, y), PAL['W'] if sum(p[:3]) > 300 else PAL['g'])
        return f
    # recoil: squint, body leans back one pixel, staff tips out
    t = list(DOWN)
    t[15] = "....OSsSSsSO...."
    t[16] = "....OSSOOSSO...."
    grid(img, FEET_DOWN[0], 0, 36)
    grid(img, t[22:24], 0, 22)
    grid(img, t[:24], 1, -1)
    grid(img, t[24:], 0, 24)
    staff(img, 0, 7)
    hand(img, 2, 26)
    return img

def frame_meditate(i):
    """Sitting cross-legged, eyes shut, breathing. A mote of mana rises."""
    img = Image.new('RGBA', (W, H))
    breath = 1 if i in (1, 2) else 0
    top = 40 - len(MEDITATE)
    grid(img, MEDITATE[:22], 0, top - breath)
    grid(img, MEDITATE[22:], 0, top + 22)
    if breath:
        # close the one-pixel gap under the shoulders
        grid(img, [MEDITATE[21]], 0, top + 21)
    mx, my = [(13, 12), (14, 9), (13, 6), (12, 3)][i]
    img.putpixel((mx, my), PAL['g'])
    if i in (1, 2):
        img.putpixel((mx, my + 1), PAL['G'])
    mx, my = [(2, 6), (1, 3), (2, 12), (1, 9)][i]
    img.putpixel((mx, my), PAL['G'])
    return img

def build(out):
    rows = []
    rows.append(('walk_down', [frame_down(i) for i in range(4)], 140))
    rows.append(('walk_up', [frame_up(i) for i in range(4)], 140))
    rows.append(('walk_side', [frame_side(i) for i in range(4)], 140))
    lifts = [(2, 0), (4, 1), (4, 2), (1, 0)]
    rows.append(('cast_down', [frame_down(0, l, g) for l, g in lifts], 100))
    rows.append(('cast_up', [frame_up(0, l, g) for l, g in lifts], 100))
    rows.append(('cast_side', [frame_side(0, l, g) for l, g in lifts], 100))
    rows.append(('hurt', [frame_hurt(i) for i in range(2)], 120))
    rows.append(('meditate', [frame_meditate(i) for i in range(4)], 400))
    sheet = Image.new('RGBA', (W * 4, H * len(rows)))
    cells, tags = [], []
    for r, (tag, frames, dur) in enumerate(rows):
        start = len(cells)
        for c, f in enumerate(frames):
            sheet.alpha_composite(f, (c * W, r * H))
            cells.append((c * W, r * H, dur))
        tags.append((tag, start, len(cells) - 1))
        if tag.startswith('walk_'):
            tags.append(('idle_' + tag[5:], start, start))
    sheet.save(os.path.join(out, 'wizard.png'))
    with open(os.path.join(out, 'wizard.json'), 'w') as f:
        f.write(aseprite_json('wizard', 'wizard.png', cells, tags, (W, H)))
    return sheet

if __name__ == '__main__':
    out = sys.argv[1] if len(sys.argv) > 1 else 'out'
    s = build(out)
    zoom(s, 8, (92, 118, 72, 255)).save(os.path.join(out, 'wizard_preview.png'))
