"""Spell effects. 32x32 cells (barrier 48, puff 16), transparent, hard edges.
Run: python3 effects.py OUTDIR  -> OUTDIR/<name>.png + .json"""
import sys, os, math, random
from PIL import Image
from pal import rgba, aseprite_json, zoom

C = {k: rgba(v) for k, v in {
    'white': 'f6f6f0', 'core': 'ffe678', 'glow': 'f0c878', 'fire': 'ff8c28', 'red': 'aa323c',
    'ember': 'a07c54', 'smoke': '605a6c', 'smokeD': '4e4858',
    'life': '78e68c', 'lifeD': '54764c', 'lifeM': '6c8654',
    'thunder': 'faf078', 'thunderD': 'ceaa5a', 'air': 'c8e6f0', 'airD': '9c948a',
    'ward': '6ebef0', 'wardD': '4e708e', 'wardDD': '345470',
    'void': 'a06ef0', 'voidD': '605a6c', 'voidDD': '424670',
    'dust': '92887c', 'dustD': '80766a',
}.items()}

def new(n):
    return Image.new('RGBA', (n, n))

def put(img, x, y, c):
    x, y = int(round(x)), int(round(y))
    if 0 <= x < img.width and 0 <= y < img.height:
        img.putpixel((x, y), c)

def disc(img, cx, cy, r, c, sx=1.0, sy=1.0):
    for y in range(img.height):
        for x in range(img.width):
            dx, dy = (x + .5 - cx) / sx, (y + .5 - cy) / sy
            if dx * dx + dy * dy <= r * r:
                img.putpixel((x, y), c)

def ring(img, cx, cy, r, w, c, sx=1.0, sy=1.0, arc=None):
    for y in range(img.height):
        for x in range(img.width):
            dx, dy = (x + .5 - cx) / sx, (y + .5 - cy) / sy
            d = math.hypot(dx, dy)
            if r - w <= d <= r:
                if arc:
                    a = math.atan2(dy, dx) % (2 * math.pi)
                    a0, a1 = arc
                    if not ((a - a0) % (2 * math.pi) <= (a1 - a0) % (2 * math.pi)):
                        continue
                img.putpixel((x, y), c)

def ellipse(img, cx, cy, rx, ry, c, arc=None):
    """One-pixel ellipse outline, closed (no gaps), optionally only an arc (a0, span)."""
    n = int(8 * (rx + ry)) + 8
    a0, span = arc if arc else (0, 2 * math.pi)
    for k in range(n + 1):
        a = a0 + span * k / n
        put(img, math.floor(cx + math.cos(a) * rx), math.floor(cy + math.sin(a) * ry), c)

def line(img, x0, y0, x1, y1, c):
    x0, y0, x1, y1 = map(int, map(round, (x0, y0, x1, y1)))
    dx, dy = abs(x1 - x0), -abs(y1 - y0)
    sx, sy = (1 if x0 < x1 else -1), (1 if y0 < y1 else -1)
    err = dx + dy
    while True:
        put(img, x0, y0, c)
        if x0 == x1 and y0 == y1:
            break
        e2 = 2 * err
        if e2 >= dy:
            err += dy; x0 += sx
        if e2 <= dx:
            err += dx; y0 += sy

def plus(img, x, y, c, cc=None, big=False):
    for dx, dy in [(0, 0), (1, 0), (-1, 0), (0, 1), (0, -1)]:
        put(img, x + dx, y + dy, c)
    if big:
        for dx, dy in [(2, 0), (-2, 0), (0, 2), (0, -2)]:
            put(img, x + dx, y + dy, c)
    if cc:
        put(img, x, y, cc)

# ---- fireball -------------------------------------------------------------

def fireball_fly(i):
    img = new(32)
    rnd = random.Random(100 + i)
    hx, cy = 21, 16  # head centre; flies to the right
    # tail: flame tongues trailing left, flickering per frame
    for x in range(3, hx + 1):
        t = (hx - x) / (hx - 3)           # 0 at head, 1 at tail end
        half = 5.5 * (1 - t) ** 0.8 + 0.5
        wob = math.sin(x * 0.9 + i * 1.6) * 1.2 * t
        for y in range(32):
            d = abs(y + .5 - (cy + wob))
            if d <= half:
                k = d / half
                if t > 0.7 and rnd.random() < 0.35:
                    continue
                c = C['red'] if k > 0.75 or t > 0.75 else C['fire'] if k > 0.4 or t > 0.45 else C['core']
                put(img, x, y, c)
    disc(img, hx, cy + .5, 5.5, C['red'])
    disc(img, hx, cy + .5, 4.6, C['fire'])
    disc(img, hx + .5, cy + .5, 3.2, C['core'])
    disc(img, hx + 1, cy, 1.6, C['white'])
    # embers falling off the tail
    for k in range(3):
        ex = 4 + ((i * 5 + k * 7) % 12)
        ey = cy + [-6, 5, -4, 7][(i + k) % 4]
        put(img, ex, ey, C['fire'] if k % 2 else C['glow'])
    return img

def fireball_hit(i):
    img = new(32)
    cx, cy = 16, 16
    rnd = random.Random(7)
    if i == 0:
        disc(img, cx, cy, 5, C['core']); disc(img, cx, cy, 3, C['white'])
    elif i == 1:
        disc(img, cx, cy, 9, C['fire']); disc(img, cx, cy, 6.5, C['core']); disc(img, cx, cy, 3.5, C['white'])
    elif i == 2:
        disc(img, cx, cy, 12, C['red']); disc(img, cx, cy, 10, C['fire']); disc(img, cx, cy, 6, C['core'])
    elif i == 3:
        ring(img, cx, cy, 13.5, 3, C['red']); ring(img, cx, cy, 11, 2.5, C['fire'])
        disc(img, cx, cy - 2, 5, C['smoke']); disc(img, cx - 2, cy - 3, 3, C['smokeD'])
    elif i == 4:
        for a in range(8):
            r = 12
            put(img, cx + math.cos(a * .785) * r, cy + math.sin(a * .785) * r, C['fire'])
        disc(img, cx - 3, cy - 5, 4, C['smoke']); disc(img, cx + 3, cy - 3, 3, C['smoke'])
        disc(img, cx, cy - 6, 2.5, C['smokeD'])
    else:
        for a in range(8):
            put(img, cx + math.cos(a * .785 + .3) * 14, cy + math.sin(a * .785 + .3) * 14 + 1, C['red'])
        disc(img, cx - 3, cy - 8, 2.5, C['smokeD']); disc(img, cx + 3, cy - 6, 2, C['smokeD'])
    return img

# ---- heal ------------------------------------------------------------------

def heal(i):
    """Plays on the caster: centre of the cell on the chest, ring at the feet."""
    img = new(32)
    # ring at the feet, expanding then fading
    r = [4, 7, 10, 12, 13, 13, 13, 0][i]
    if r:
        col = C['life'] if i < 5 else C['lifeD']
        ellipse(img, 16, 26, r, r * 0.35, col)
    # sparkles rising
    seeds = [(8, 26, 0), (23, 27, 1), (13, 29, 2), (19, 25, 3), (11, 24, 4), (25, 23, 2)]
    for sx, sy, d in seeds:
        k = i - d
        if 0 <= k < 5:
            y = sy - k * 4
            c = C['white'] if k == 1 else C['life'] if k < 4 else C['lifeD']
            plus(img, sx, y, c, C['white'] if k in (1, 2) else None, big=(k == 2))
    return img

# ---- thunderstrike ---------------------------------------------------------

def bolt_path(seed):
    rnd = random.Random(seed)
    pts, x = [], 16 + rnd.randint(-3, 3)
    for y in range(0, 30, 5):
        pts.append((x, y))
        x = max(8, min(24, x + rnd.choice([-4, -3, 3, 4])))
    pts.append((16, 27))
    return pts

def thunder(i):
    """Bolt from the top of the cell to the impact at the bottom centre."""
    img = new(32)
    cx, cy = 16, 27
    if i == 0:
        for p, q in zip(bolt_path(3), bolt_path(3)[1:]):
            line(img, *p, *q, C['airD'])
        return img
    if i in (1, 2, 3):
        path = bolt_path(3 if i != 3 else 5)
        for p, q in zip(path, path[1:]):
            for o in (-1, 1):
                line(img, p[0] + o, p[1], q[0] + o, q[1], C['thunderD'] if i == 3 else C['thunder'])
        for p, q in zip(path, path[1:]):
            line(img, *p, *q, C['white'])
        if i == 2:  # a branch
            line(img, path[2][0], path[2][1], path[2][0] + 6, path[2][1] + 7, C['thunder'])
    if i in (1, 2):
        disc(img, cx, cy, 5 if i == 1 else 7, C['thunder'], sy=0.6)
        disc(img, cx, cy, 3, C['white'], sy=0.6)
    if i >= 3:
        r = [0, 0, 0, 9, 12, 14][i]
        ellipse(img, cx, cy, r, r * 0.3, C['thunder'] if i < 5 else C['thunderD'])
        for a in range(6):
            ang = math.pi + a * math.pi / 5
            d = r * 0.7
            put(img, cx + math.cos(ang) * d, cy + math.sin(ang) * d * 1.4, C['white'] if i < 5 else C['thunder'])
    return img

# ---- gust ------------------------------------------------------------------

def gust(i):
    """A swirl of wind around the caster."""
    img = new(32)
    rot = i * 0.9
    for k, (r, w) in enumerate([(13, 1.2), (10, 1.2), (7, 1)]):
        a0 = rot + k * 2.1
        span = [1.2, 1.9, 2.4, 2.4, 1.8, 1.0][i]
        col = C['white'] if k == 0 and i in (2, 3) else C['air'] if i < 5 else C['airD']
        ellipse(img, 16, 18, r, r * 0.6, col, arc=(a0, span))
    # streaks
    for k in range(3):
        y = 8 + k * 7 + (i % 2)
        x = 2 + ((i * 4 + k * 9) % 18)
        line(img, x, y, x + 4, y, C['air'] if i < 5 else C['airD'])
    return img

# ---- barrier (48x48) ---------------------------------------------------------

def bubble(img, r, hi, fill_a=60):
    cx = cy = 24
    fill = C['ward'][:3] + (fill_a,)
    disc(img, cx, cy, r, fill)
    ring(img, cx, cy, r, 1.2, C['ward'])
    ring(img, cx, cy, r - 1.2, 1, C['wardD'][:3] + (160,))
    # highlight arc moves round the top left
    ring(img, cx, cy, r - 3, 1.3, C['white'], arc=(hi, hi + 0.7))
    put(img, cx - r * 0.45, cy - r * 0.45, C['white'])

def barrier(tag, i):
    img = new(48)
    if tag == 'appear':
        r = [7, 14, 20][i]
        bubble(img, r, 3.6, [30, 50, 80][i])
        if i == 2:
            ring(img, 24, 24, 21, 1, C['air'])
    elif tag == 'hold':
        bubble(img, 20, 3.3 + i * 0.25)
        # faint hex shimmer
        for k in range(3):
            a = i * 0.8 + k * 2.1
            put(img, 24 + math.cos(a) * 12, 24 + math.sin(a) * 12, C['air'][:3] + (200,))
    else:  # break: shards fly out
        rnd = random.Random(9)
        d = [21, 24, 27][i]
        for k in range(10):
            a = k * 0.628 + rnd.random() * 0.3
            x, y = 24 + math.cos(a) * d, 24 + math.sin(a) * d
            line(img, x, y, x + math.cos(a + 1.3) * (3 - i), y + math.sin(a + 1.3) * (3 - i), C['ward'] if i < 2 else C['wardD'])
        if i == 0:
            disc(img, 24, 24, 16, C['ward'][:3] + (40,))
    return img

# ---- teleport ----------------------------------------------------------------

def teleport_out(i):
    """The caster folds into a column of void that collapses to a point."""
    img = new(32)
    w = [8, 6, 3, 1, 0][i]
    if w:
        for y in range(2, 31):
            for x in range(16 - w, 16 + w):
                c = C['void'] if abs(x + .5 - 16) < w * 0.5 else C['voidDD']
                put(img, x, y, c[:3] + (220,))
        line(img, 16, 2, 16, 30, C['white'] if i in (1, 2) else C['void'])
    r = [6, 9, 11, 12, 13][i]
    ellipse(img, 16, 29, r, r * 0.3, C['void'] if i < 4 else C['voidD'])
    rnd = random.Random(4)
    for k in range(7):
        x = 16 + rnd.randint(-10, 10)
        y = 28 - rnd.randint(0, 10) - i * 4
        put(img, x, y, C['void'] if k % 2 else C['white'] if i < 3 else C['voidD'])
    if i == 3:
        plus(img, 16, 16, C['void'], C['white'], big=True)
    if i == 4:
        put(img, 16, 16, C['white'])
    return img

def teleport_in(i):
    img = teleport_out(4 - i)
    if i == 0:
        disc(img, 16, 16, 3, C['white'])
    return img

# ---- generic hit flash and dust puff --------------------------------------------

def hit(i):
    img = new(32)
    if i == 0:
        for d in range(1, 9):
            for dx, dy in [(d, 0), (-d, 0), (0, d), (0, -d)]:
                put(img, 16 + dx, 16 + dy, C['white'])
        for d in range(1, 5):
            for dx, dy in [(d, d), (-d, d), (d, -d), (-d, -d)]:
                put(img, 16 + dx, 16 + dy, C['white'])
        disc(img, 16.5, 16.5, 3, C['white'])
    elif i == 1:
        ring(img, 16.5, 16.5, 7, 1.3, C['white'])
        for dx, dy in [(10, 0), (-10, 0), (0, 10), (0, -10)]:
            put(img, 16 + dx, 16 + dy, C['white'])
    else:
        for a in range(8):
            put(img, 16 + math.cos(a * .785) * 10, 16 + math.sin(a * .785) * 10, C['air'])
    return img

def puff(i):
    img = new(16)
    r = [2, 3, 3.5, 3][i]
    col = C['dust'] if i < 2 else C['dustD']
    for cx, cy in [(5, 12), (10, 12), (7.5, 10)]:
        if i == 3 and cx == 7.5:
            continue
        disc(img, cx + (cx - 7.5) * i * 0.3, cy - i * 0.7, r * (0.8 if cx == 7.5 else 1), col)
    return img

# ---- sheets ------------------------------------------------------------------

def write(out, name, size, tags, dur=60):
    """tags: list of (tag, [frames]) or (tag, [frames], ms). One row per tag, 8 columns."""
    cols = 8
    nrows = sum((len(t[1]) + cols - 1) // cols for t in tags)
    sheet = Image.new('RGBA', (size * cols, size * nrows))
    cells, tlist, row = [], [], 0
    for t in tags:
        tag, frames = t[0], t[1]
        ms = t[2] if len(t) > 2 else dur
        start = len(cells)
        for k, f in enumerate(frames):
            x, y = (k % cols) * size, (row + k // cols) * size
            sheet.alpha_composite(f, (x, y))
            cells.append((x, y, ms))
        tlist.append((tag, start, len(cells) - 1))
        row += (len(frames) + cols - 1) // cols
    sheet.save(os.path.join(out, name + '.png'))
    with open(os.path.join(out, name + '.json'), 'w') as fh:
        fh.write(aseprite_json(name, name + '.png', cells, tlist, (size, size)))
    return sheet

def build(out):
    sheets = {}
    sheets['fireball'] = write(out, 'fireball', 32, [('fly', [fireball_fly(i) for i in range(4)], 80),
                                                     ('hit', [fireball_hit(i) for i in range(6)])])
    sheets['heal'] = write(out, 'heal', 32, [('self', [heal(i) for i in range(8)], 70)])
    sheets['thunder'] = write(out, 'thunder', 32, [('strike', [thunder(i) for i in range(6)], 50)])
    sheets['gust'] = write(out, 'gust', 32, [('self', [gust(i) for i in range(6)], 60)])
    sheets['barrier'] = write(out, 'barrier', 48, [('appear', [barrier('appear', i) for i in range(3)]),
                                                   ('hold', [barrier('hold', i) for i in range(4)], 120),
                                                   ('break', [barrier('break', i) for i in range(3)])])
    sheets['teleport'] = write(out, 'teleport', 32, [('out', [teleport_out(i) for i in range(5)]),
                                                     ('in', [teleport_in(i) for i in range(5)])])
    sheets['hit'] = write(out, 'hit', 32, [('hit', [hit(i) for i in range(3)], 50)])
    sheets['puff'] = write(out, 'puff', 16, [('puff', [puff(i) for i in range(4)], 80)])
    return sheets

if __name__ == '__main__':
    out = sys.argv[1] if len(sys.argv) > 1 else 'out'
    os.makedirs(out, exist_ok=True)
    sheets = build(out)
    # preview: every sheet stacked on dark ground at 4x
    k = 4
    wmax = max(s.width for s in sheets.values()) * k
    h = sum(s.height * k + 8 for s in sheets.values())
    prev = Image.new('RGBA', (wmax, h), (38, 34, 46, 255))
    y = 0
    for s in sheets.values():
        prev.alpha_composite(zoom(s, k), (0, y)); y += s.height * k + 8
    prev.save(os.path.join(out, 'effects_preview.png'))
