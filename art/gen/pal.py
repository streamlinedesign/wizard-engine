"""Style B palette (artpack/palette.gpl) as single-letter keys, plus helpers."""
import json
from PIL import Image

HEX = {
    'O': '26222e',  # outline
    'K': '1a1822',  # shadow
    'R': '383c64',  # robe
    'r': '262846',  # robe shade
    'H': '2c2e50',  # hat
    'h': '424670',  # hat light
    'S': 'd6b08e',  # skin
    's': 'b28c6e',  # skin shade
    'B': 'ceccd0',  # beard
    'b': '9c948a',  # beard shade
    'Y': 'ceaa5a',  # belt gold
    'y': 'a07c54',  # gold shade
    'T': '82603e',  # staff wood
    't': '5a4230',  # staff dark
    'G': '6ebef0',  # gem (ward blue)
    'g': 'c8e6f0',  # gem light (air)
    'W': 'f6f6f0',  # white
    'E': '281e28',  # eye (ink)
    'L': 'f0c878',  # lamp glow
    'F': 'ff8c28',  # fire
    'f': 'ffe678',  # fire core
    'w': 'aa323c',  # wax red
    'V': 'a06ef0',  # void
    'P': '78e68c',  # life
    'U': 'faf078',  # thunder
}
def rgba(h, a=255):
    return (int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), a)
PAL = {k: rgba(v) for k, v in HEX.items()}

def grid(img, rows, ox=0, oy=0, pal=PAL):
    for y, row in enumerate(rows):
        for x, ch in enumerate(row):
            if ch in '. ':
                continue
            px, py = ox + x, oy + y
            if 0 <= px - ox < 10**6 and 0 <= px < img.width and 0 <= py < img.height:
                img.putpixel((px, py), pal[ch])

ONCE = ('hit', 'self', 'strike', 'appear', 'break', 'out', 'in', 'puff', 'hurt')

def tag(t):
    n, a, b = t
    d = {"name": n, "from": a, "to": b, "direction": "forward", "color": "#000000ff"}
    if n in ONCE or n.startswith('cast_'):
        d["repeat"] = "1"  # one-shot: the engine reports Finished() after one play
    return d

def aseprite_json(name, png, cells, tags, size):
    """cells: list of (x, y, duration_ms). tags: list of (name, from, to)."""
    w, h = size
    frames = []
    for i, (x, y, d) in enumerate(cells):
        frames.append({
            "filename": f"{name} {i}.aseprite",
            "frame": {"x": x, "y": y, "w": w, "h": h},
            "rotated": False, "trimmed": False,
            "spriteSourceSize": {"x": 0, "y": 0, "w": w, "h": h},
            "sourceSize": {"w": w, "h": h},
            "duration": d,
        })
    meta = {
        "app": "https://www.aseprite.org/", "version": "1.3",
        "image": png, "format": "RGBA8888",
        "size": {"w": 0, "h": 0}, "scale": "1",
        "frameTags": [tag(t) for t in tags],
        "layers": [{"name": "Layer", "opacity": 255, "blendMode": "normal"}],
        "slices": [],
    }
    return json.dumps({"frames": frames, "meta": meta}, indent=1)

def zoom(img, k, bg=None):
    out = img.resize((img.width * k, img.height * k), Image.NEAREST)
    if bg:
        b = Image.new('RGBA', out.size, bg)
        b.alpha_composite(out)
        return b
    return out
