"""Campus tileset, 16x16, 8 columns. Ids 1-24 keep the order the game uses
(game/scenes/campus.go); new tiles are appended. Run: python3 tiles.py OUTDIR"""
import sys, os, random
from PIL import Image
from pal import rgba, zoom

T, COLS = 16, 8
X = {k: rgba(v) for k, v in {
    'O': '26222e',
    'g': '5c7648', 'gd': '4a603a', 'gl': '6c8654',
    'st': '80766a', 'sd': '686056', 'sl': '92887c',
    'pz': '9c948a', 'pzd': '80766a', 'pzl': 'a09482',
    'wa': '345470', 'wl': '4e708e', 'air': 'c8e6f0',
    'sla': '4e4858', 'sld': '3a3442', 'sll': '605a6c',
    'pl': 'beb29e', 'pld': 'a09482', 'tim': '5a4230',
    'wo': '82603e', 'wol': 'a07c54', 'trunk': '4a3628',
    'lf': '3e603e', 'lfd': '2a442e', 'lfl': '54764c',
    'glow': 'f0c878', 'gold': 'ceaa5a', 'parch': 'ece2c8', 'parchd': 'c8ba96',
    'wax': 'aa323c', 'ink': '281e28', 'robe': '383c64', 'void': 'a06ef0',
    'iron': '3a3442', 'ironl': '605a6c', 'white': 'f6f6f0', 'fire': 'ff8c28',
}.items()}
SHADOW = (26, 24, 34, 90)

class Tile:
    def __init__(self, img, idx):
        self.img, self.ox, self.oy = img, (idx % COLS) * T, (idx // COLS) * T
    def p(self, x, y, c, wrap=True):
        if wrap:
            x, y = x % T, y % T
        elif not (0 <= x < T and 0 <= y < T):
            return
        if len(c) == 4 and c[3] < 255:
            base = self.img.getpixel((self.ox + x, self.oy + y))
            if base[3]:
                a = c[3] / 255
                c = tuple(int(base[k] * (1 - a) + c[k] * a) for k in range(3)) + (255,)
        self.img.putpixel((self.ox + x, self.oy + y), c)
    def fill(self, c):
        for y in range(T):
            for x in range(T):
                self.p(x, y, c)
    def rect(self, x0, y0, x1, y1, c):
        for y in range(y0, y1 + 1):
            for x in range(x0, x1 + 1):
                self.p(x, y, c, wrap=False)
    def rows(self, rows, key, ox=0, oy=0):
        for y, r in enumerate(rows):
            for x, ch in enumerate(r):
                if ch != '.':
                    self.p(ox + x, oy + y, X[key[ch]], wrap=False)

def grass(t, rnd, tufts=7):
    t.fill(X['g'])
    for _ in range(14):
        t.p(rnd.randrange(T), rnd.randrange(T), X['gd'])
    for _ in range(tufts):  # little blades: dark root, light tip
        x, y = rnd.randrange(T), rnd.randrange(T)
        t.p(x, y, X['gd']); t.p(x + 1, y, X['gd']); t.p(x - 1, y - 1, X['gl']); t.p(x + 1, y - 1, X['gl'])
    for _ in range(6):
        t.p(rnd.randrange(T), rnd.randrange(T), X['gl'])

def cobbles(t, rnd):
    t.fill(X['sd'])
    # running bond of rounded stones, 5x4 with 1px mortar, wraps on both axes
    for row in range(4):
        y0 = row * 4
        off = (row % 2) * 3 + rnd.randrange(2)
        x = off
        while x < off + T:
            w = rnd.choice([4, 5, 5, 6])
            for yy in range(y0, y0 + 3):
                for xx in range(x, x + w - 1):
                    corner = (yy in (y0, y0 + 2)) and (xx in (x, x + w - 2))
                    if not corner:
                        t.p(xx, yy, X['st'])
            for xx in range(x + 1, x + w - 2):
                t.p(xx, y0, X['sl'])
            t.p(x, y0 + 1, X['sl'])
            x += w

def flags(t):
    t.fill(X['pz'])
    for i in range(T):
        t.p(i, 15, X['pzd']); t.p(15, i, X['pzd'])
        t.p(i, 7, X['pzd'])
    for i in range(8, 15):
        t.p(7, i, X['pzd'])
    for i in range(0, 7):
        t.p(11, i, X['pzd'])
    for x, y in [(0, 0), (12, 0), (0, 8), (8, 8)]:
        for k in range(3):
            t.p(x + k, y, X['pzl'])
        t.p(x, y + 1, X['pzl'])
    for x, y in [(4, 3), (13, 4), (3, 12), (11, 11), (14, 10)]:
        t.p(x, y, X['pzd'])

def water(t, rnd):
    t.fill(X['wa'])
    for y, x, n in [(2, 1, 3), (5, 9, 4), (9, 3, 3), (12, 11, 3), (14, 5, 2)]:
        for k in range(n):
            t.p(x + k, y, X['wl'])
        t.p(x + 1, y - 1, X['wl'])
    t.p(12, 5, X['air']); t.p(4, 12, X['air'])

def build(out):
    rnd = random.Random(16)
    img = Image.new('RGBA', (T * COLS, T * 4))
    tile = lambda i: Tile(img, i - 1)

    # 1 grass, 2 worn, 3 flowers
    grass(tile(1), rnd)
    t = tile(2); grass(t, rnd, 3)
    for x, y in [(3, 4), (4, 4), (5, 5), (9, 10), (10, 10), (10, 11), (11, 11), (13, 3), (6, 13), (2, 9)]:
        t.p(x, y, X['sd']); t.p(x, y + 1, X['st'])
    t = tile(3); grass(t, rnd, 4)
    for (x, y), c in zip([(3, 3), (11, 5), (6, 10), (13, 12), (1, 13)], ['parch', 'glow', 'wax', 'parch', 'glow']):
        t.p(x, y, X[c]); t.p(x - 1, y, X[c]); t.p(x + 1, y, X[c]); t.p(x, y - 1, X[c]); t.p(x, y + 1, X['gd'])
        t.p(x, y, X['gold'] if c != 'glow' else X['wol'])
    # 4 stone path
    cobbles(tile(4), rnd)
    # 5 plaza flags
    flags(tile(5))
    # 6 water
    water(tile(6), rnd)
    # 7 shore: grass to the west, stone lip, water to the east
    t = tile(7); grass(t, rnd, 4)
    for y in range(T):
        t.p(10, y, X['sl']); t.p(11, y, X['st']); t.p(12, y, X['sd'])
        for x in range(13, 16):
            t.p(x, y, X['wa'])
        if y % 4 == 0:
            t.p(11, y, X['sl'])
    t.p(14, 4, X['wl']); t.p(15, 4, X['wl']); t.p(13, 11, X['wl']); t.p(14, 11, X['wl'])
    # 8 wall: plaster with a timber post on the left edge
    t = tile(8); t.fill(X['pl'])
    for x, y in [(4, 3), (9, 6), (13, 11), (5, 12), (11, 2), (7, 9)]:
        t.p(x, y, X['pld'])
    t.rect(0, 0, 1, 15, X['tim'])
    # 9 wall top: eave shadow, beam, plaster
    t = tile(9); t.fill(X['pl'])
    t.rect(0, 0, 15, 2, X['sld']); t.rect(0, 3, 15, 3, X['O'])
    t.rect(0, 4, 15, 6, X['tim']); t.rect(0, 7, 15, 7, X['pld'])
    t.rect(0, 8, 1, 15, X['tim'])
    for x in range(3, 16, 5):
        t.p(x, 5, X['wo'])
    # 10 roof slates, 11 ridge
    def slates(t, y0=0):
        t.rect(0, y0, 15, 15, X['sla'])
        for row, y in enumerate(range(y0, T, 4)):
            t.rect(0, y + 3, 15, y + 3, X['sld']) if y + 3 < T else None
            off = (row % 2) * 4
            for x in range(off, T + off, 8):
                for yy in range(y, min(y + 3, T)):
                    t.p(x % T, yy, X['sld'])
                for xx in range(x + 1, x + 5):
                    t.p(xx % T, y, X['sll'])
    slates(tile(10))
    t = tile(11); slates(t, 4)
    t.rect(0, 0, 15, 0, X['O']); t.rect(0, 1, 15, 1, X['sll']); t.rect(0, 2, 15, 3, X['sld'])
    for x in range(0, 16, 4):
        t.p(x, 1, X['sla'])
    # 12 door (closed), 13 window
    def door(t, open_=False):
        t.fill(X['pl']); t.rect(0, 0, 1, 15, X['tim'])
        t.rect(3, 1, 12, 15, X['tim'])
        if open_:
            t.rect(4, 2, 11, 15, X['ink']); t.rect(4, 13, 11, 15, X['sld'])
            t.rect(4, 2, 5, 15, X['wo'])
        else:
            t.rect(4, 2, 11, 15, X['wo'])
            t.rect(7, 2, 8, 15, X['tim'])
            for x in (5, 10):
                t.rect(x, 3, x, 14, X['wol'])
            t.p(6, 9, X['glow']); t.p(9, 9, X['glow'])
            t.rect(4, 6, 11, 6, X['tim'])
    def window(t, lit=False):
        t.fill(X['pl']); t.rect(0, 0, 1, 15, X['tim'])
        t.rect(4, 2, 12, 11, X['tim'])
        glass, glint = (X['glow'], X['parch']) if lit else (X['wl'], X['air'])
        t.rect(5, 3, 11, 10, glass)
        t.rect(8, 3, 8, 10, X['tim']); t.rect(5, 6, 11, 6, X['tim'])
        t.p(5, 3, glint); t.p(9, 3, glint); t.p(6, 4, glint) if not lit else None
        if not lit:
            t.rect(5, 9, 7, 10, X['wa']); t.rect(9, 9, 11, 10, X['wa'])
        t.rect(3, 12, 13, 12, X['wo']); t.rect(3, 13, 13, 13, X['tim'])
    door(tile(12)); window(tile(13))
    # 14 tree top, 15 tree bottom: one tree split across two tiles
    tree = [
        "......OOOO......",
        "....OOllLLOO....",
        "...OlllLLLLLO...",
        "..OllLLLLLLLLO..",
        "..OlLLLLLLLLdO..",
        ".OlLLLLLLLLLLdO.",
        ".OlLLLLlLLLLddO.",
        ".OLLLLllLLLLLdO.",
        "OlLLLLLLLLLLLddO",
        "OLLLLLLLLLLLdddO",
        "OLLLLLLLLLLLLddO",
        "OLlLLLLLLLLLdddO",
        "OLLLLLLLLLLLLddO",
        ".OLLLLLLLLLLddO.",
        ".OLLLLLLLLLdddO.",
        ".ODLLLLLLLLdddO.",
        "..ODLLLLLLdddO..",
        "..OdDLLLdLddO...",
        "...OddOddddO....",
        "....OOOtTOOO....",
        "......OtTO......",
        "......OtTO......",
        "......OtTO......",
        ".....OttTTO.....",
        ".....OOOOOO.....",
    ]
    key = {'O': 'O', 'L': 'lf', 'l': 'lfl', 'd': 'lfd', 'D': 'lfd', 't': 'trunk', 'T': 'wo'}
    top, bot = tile(14), tile(15)
    for y in range(T):
        for x in range(T):
            pass
    # leaf dapple on the light side
    trows = [list(r) for r in tree]
    for x, y in [(4, 4), (7, 3), (5, 8), (9, 6), (3, 10), (6, 12), (10, 9), (8, 14)]:
        if trows[y][x] == 'L':
            trows[y][x] = 'l'
    for x, y in [(11, 13), (6, 16), (9, 15), (4, 14)]:
        if trows[y][x] == 'L':
            trows[y][x] = 'd'
    tree = [''.join(r) for r in trows]
    top.rows(tree[:T], key, 0, 0)
    for x in range(3, 13):  # soft ground shadow round the trunk's foot
        for y in (8, 9):
            if abs(x - 7.5) < (5 if y == 8 else 3.5):
                bot.p(x, y, SHADOW, wrap=False)
    bot.rows(tree[T:], key, 0, 0)
    # 16 hedge: a dense leafy block that tiles both ways
    t = tile(16); t.fill(X['lf'])
    for _ in range(26):
        x, y = rnd.randrange(T), rnd.randrange(T)
        t.p(x, y, X['lfl']); t.p(x + 1, y + 1, X['lfd'])
    for x in range(T):
        t.p(x, 14, X['lfd']); t.p(x, 15, X['lfd'] if x % 3 else X['O'])
        if x % 4 == 1:
            t.p(x, 0, X['lfl']); t.p(x + 1, 0, X['lfl'])
    # 17 lamp
    t = tile(17)
    t.rows([
        "......OOOO......",
        ".....OgggGO.....",
        ".....OGwwGO.....",
        ".....OGwwGO.....",
        ".....OOOOOO.....",
        ".......OO.......",
        ".......Oi.......",
        ".......Oi.......",
        ".......Oi.......",
        ".......Oi.......",
        ".......Oi.......",
        ".......Oi.......",
        ".......Oi.......",
        "......OOiO......",
        ".....OiiiiO.....",
        ".....OOOOOO.....",
    ], {'O': 'O', 'g': 'ironl', 'G': 'iron', 'w': 'glow', 'i': 'ironl'})
    t.p(6, 2, X['parch'])
    # 18 fountain on flags
    t = tile(18); flags(t)
    t.rows([
        "................",
        "....OOOOOOOO....",
        "..OOllllllllOO..",
        ".OlsWWWWWWWWslO.",
        ".OsWWwwwwwwWWsO.",
        "OlsWwwwAAwwwWslO",
        "OlsWwwAAAAwwWslO",
        "OlsWwwwSSwwwWslO",
        "OlsWwwwSSwwwWslO",
        "OlsWWwwSSwwWWslO",
        ".OsWWWWSSWWWWsO.",
        ".OssssssssssssO.",
        "..OOssssssssOO..",
        "....OOOOOOOO....",
        "................",
        "................",
    ], {'O': 'O', 'l': 'sl', 's': 'st', 'W': 'wl', 'w': 'wa', 'A': 'air', 'S': 'sl'})
    t.p(7, 4, X['white']); t.p(8, 3, X['air'])
    # 19 bench
    tile(19).rows([
        "................",
        "................",
        "................",
        "................",
        "OOOOOOOOOOOOOOOO",
        "OllllllllllllllO",
        "OwwwwwwwwwwwwwwO",
        "OOOOOOOOOOOOOOOO",
        "OllllllllllllllO",
        "OwwwwwwwwwwwwwwO",
        "OddddddddddddddO",
        "OOdOOOOOOOOOOdOO",
        ".OdO........OdO.",
        ".OdO........OdO.",
        ".OOO........OOO.",
        "................",
    ], {'O': 'O', 'l': 'wol', 'w': 'wo', 'd': 'tim'})
    # 20 signpost
    tile(20).rows([
        "................",
        ".OOOOOOOOOOOOO..",
        ".OlllllllllllO..",
        ".OlwwwwwwwwwlwO.",
        ".OlwiiiwiiwwlwO.",
        ".OlwwwwwwwwwwwO.",
        ".OlwiiwiiiiwwO..",
        ".OddddddddddO...",
        ".OOOOOdtOOOOO...",
        "......OdtO......",
        "......OdtO......",
        "......OdtO......",
        "......OdtO......",
        "......OdtO......",
        ".....OOdtOO.....",
        "......OOOO......",
    ], {'O': 'O', 'l': 'wol', 'w': 'wo', 'd': 'tim', 't': 'wo', 'i': 'tim'})
    # 21 bookshelf
    t = tile(21)
    t.rect(0, 0, 15, 15, X['O']); t.rect(1, 1, 14, 14, X['tim'])
    spines = ['wax', 'robe', 'lfl', 'gold', 'sll', 'wl', 'parchd', 'wax', 'lf', 'void']
    for row in range(3):
        y = 2 + row * 4
        x = 2
        k = row * 3
        while x < 14:
            w = 1 + (k % 3 == 0)
            h = 3 - (k % 4 == 1)
            c = X[spines[k % len(spines)]]
            for xx in range(x, min(x + w, 14)):
                for yy in range(y + 3 - h, y + 3):
                    t.p(xx, yy, c, wrap=False)
            x += w + (k % 5 == 4)
            k += 1
        t.rect(1, y + 3, 14, y + 3, X['wo'])
    t.rect(1, 1, 14, 1, X['wo'])
    # 22 desk
    tile(22).rows([
        "................",
        "................",
        "................",
        "....pppp....i...",
        "OOOOpPPpOOOOOOOO",
        "OllllllllllllllO",
        "OwwwwwwwwwwwwwwO",
        "OddddddddddddddO",
        "OwwwwwwwOwwwwwwO",
        "OwOOOOwwOwwwwwwO",
        "OwwwwwwwOwOOOOwO",
        "OddddddwOwwwwwwO",
        "OOOOOOOOOOOOOOOO",
        ".OdO........OdO.",
        ".OdO........OdO.",
        ".OOO........OOO.",
    ], {'O': 'O', 'l': 'wol', 'w': 'wo', 'd': 'tim', 'p': 'parchd', 'P': 'parch', 'i': 'ink'})
    # 23 wood floor
    t = tile(23); t.fill(X['wo'])
    for y in range(0, T, 4):
        t.rect(0, y + 3, 15, y + 3, X['tim'])
        off = (y // 4 % 2) * 7 + 2
        t.p(off % T, y + 1, X['tim']); t.p(off % T, y + 2, X['tim'])
        for x in range(T):
            if (x + y) % 9 == 0:
                t.p(x, y, X['wol'])
    # 24 rug
    t = tile(24); t.fill(X['wax'])
    for i in range(T):
        t.p(i, 1, X['gold']); t.p(i, 14, X['gold']); t.p(1, i, X['gold']); t.p(14, i, X['gold'])
    t.rows(["..g..", ".gwg.", "gw.wg", ".gwg.", "..g.."], {'g': 'gold', 'w': 'parchd'}, 5, 5)
    for i in range(0, T, 2):
        t.p(i, 0, X['parchd']); t.p(i, 15, X['parchd'])

    # ---- new tiles (25+) ----
    # 25 grass with a stone, 26 grass with mushrooms
    t = tile(25); grass(t, rnd, 4)
    t.rows([".OOOO.", "OllssO", "OlsssO", ".OOOO."], {'O': 'O', 'l': 'sl', 's': 'st'}, 6, 7)
    t = tile(26); grass(t, rnd, 4)
    t.rows([".OOO.", "OwPwO", ".OpO.", "..p.."], {'O': 'O', 'w': 'wax', 'P': 'parch', 'p': 'parchd'}, 4, 6)
    t.rows([".OO.", "OwwO", ".p.."], {'O': 'O', 'w': 'wax', 'p': 'parchd'}, 9, 10)
    # 27 dirt, 28 gravel
    t = tile(27); t.fill(X['wo'])
    for _ in range(18):
        t.p(rnd.randrange(T), rnd.randrange(T), X['tim'])
    for _ in range(8):
        t.p(rnd.randrange(T), rnd.randrange(T), X['wol'])
    t = tile(28); t.fill(X['sd'])
    for _ in range(40):
        x, y = rnd.randrange(T), rnd.randrange(T)
        t.p(x, y, X['st']); t.p(x + 1, y, X['sl'] if rnd.random() < .4 else X['st'])
    # 29 window lit, 30 door open
    window(tile(29), lit=True); door(tile(30), open_=True)
    # 31 barrel, 32 crate
    tile(31).rows([
        "................",
        "................",
        "....OOOOOOOO....",
        "...OllwwwwwdO...",
        "...OOOOOOOOOO...",
        "...OlwwwwwwdO...",
        "..OlwwwwwwwwdO..",
        "..OiiiiiiiiiiO..",
        "..OlwwwwwwwwdO..",
        "..OlwwwwwwwwdO..",
        "..OiiiiiiiiiiO..",
        "..OlwwwwwwwwdO..",
        "...OlwwwwwwdO...",
        "...OOOOOOOOOO...",
        "................",
        "................",
    ], {'O': 'O', 'l': 'wol', 'w': 'wo', 'd': 'tim', 'i': 'ironl'})
    tile(32).rows([
        "................",
        "................",
        "..OOOOOOOOOOOO..",
        "..OllllllllllO..",
        "..OlddddddddlO..",
        "..OldwwwwwwdlO..",
        "..OlwdwwwwdwlO..",
        "..OlwwdwwdwwlO..",
        "..OlwwwddwwwlO..",
        "..OlwwdwwdwwlO..",
        "..OlwdwwwwdwlO..",
        "..OldwwwwwwdlO..",
        "..OlddddddddlO..",
        "..OllllllllllO..",
        "..OOOOOOOOOOOO..",
        "................",
    ], {'O': 'O', 'l': 'wol', 'w': 'wo', 'd': 'tim'})

    img.save(os.path.join(out, 'campus.png'))
    return img

if __name__ == '__main__':
    out = sys.argv[1] if len(sys.argv) > 1 else 'out'
    os.makedirs(out, exist_ok=True)
    img = build(out)
    zoom(img, 6, (20, 18, 26, 255)).save(os.path.join(out, 'tiles_preview.png'))
