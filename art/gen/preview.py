"""Animated previews from the exported sheets. python3 preview.py ASSETS OUTDIR"""
import sys, os, json
from PIL import Image

assets, out = sys.argv[1], sys.argv[2]
os.makedirs(out, exist_ok=True)
GRASS = (92, 118, 72, 255)
INK = (38, 34, 46, 255)

def load(j):
    d = json.load(open(j))
    img = Image.open(os.path.join(os.path.dirname(j), d['meta']['image'])).convert('RGBA')
    frames = [img.crop((f['frame']['x'], f['frame']['y'], f['frame']['x'] + f['frame']['w'], f['frame']['y'] + f['frame']['h'])) for f in d['frames']]
    durs = [f['duration'] for f in d['frames']]
    tags = {t['name']: (t['from'], t['to']) for t in d['meta']['frameTags']}
    return frames, durs, tags

def gif(path, strips, bg, k, tick=20, total=None, flip=()):
    """strips: list of (frames, durations); all play side by side on their own clocks."""
    cw = max(f[0].width for f, _ in strips); ch = max(f[0].height for f, _ in strips)
    loops = [sum(d) for _, d in strips]
    total = total or max(loops)
    canv = []
    for t in range(0, total, tick):
        c = Image.new('RGBA', ((cw + 4) * len(strips), ch + 4), bg)
        for i, (fr, du) in enumerate(strips):
            m = t % sum(du); j = 0
            while m >= du[j]:
                m -= du[j]; j += 1
            f = fr[j]
            if i in flip:
                f = f.transpose(Image.FLIP_LEFT_RIGHT)
            c.alpha_composite(f, (i * (cw + 4) + 2 + (cw - f.width) // 2, 2 + ch - f.height))
        canv.append(c.resize((c.width * k, c.height * k), Image.NEAREST).convert('RGB'))
    canv[0].save(path, save_all=True, append_images=canv[1:], duration=tick, loop=0, disposal=1)

fr, du, tg = load(os.path.join(assets, 'sprites/wizard.json'))
strip = lambda tag: ([fr[i] for i in range(tg[tag][0], tg[tag][1] + 1)], [du[i] for i in range(tg[tag][0], tg[tag][1] + 1)])
walk = [strip('walk_down'), strip('walk_side'), strip('walk_up'), strip('walk_side')]
gif(os.path.join(out, 'wizard-walk.gif'), walk, GRASS, 6, flip=(3,))
hold = lambda s, ms: (s[0], s[1][:-1] + [s[1][-1] + ms])
gif(os.path.join(out, 'wizard-cast.gif'), [hold(strip('cast_down'), 500), hold(strip('cast_side'), 500), hold(strip('cast_up'), 500),
                                           hold(strip('hurt'), 600), strip('meditate')], GRASS, 6, total=1600)
fx = []
for name, tag in [('fireball', 'fly'), ('fireball', 'hit'), ('heal', 'self'), ('thunder', 'strike'), ('gust', 'self'),
                  ('barrier', 'hold'), ('teleport', 'out'), ('teleport', 'in'), ('hit', 'hit'), ('puff', 'puff')]:
    f, d, t = load(os.path.join(assets, 'effects', name + '.json'))
    a, b = t[tag]
    fr2, du2 = f[a:b + 1], d[a:b + 1]
    if tag not in ('fly', 'hold'):
        du2 = du2[:-1] + [du2[-1] + 300]
    fx.append((fr2, du2))
gif(os.path.join(out, 'effects.gif'), fx, INK, 4, total=1200)
