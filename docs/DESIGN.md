# Wizard Game: design notes

Uriel's notebook, rewritten as specs (2026-10-28), kept here mostly in his
own words so the game's intent lives next to the code. Headings and bullets
added for navigation; nothing dropped. Engine mapping notes are in italics.

## Pillars

- **Gear = Power. Knowledge = Ease of use. Mana = Capacity.**
- Massive unique spell crafting that works with a lot of stuff, but the
  fireball still has the potential to be broken.
- Massive hidden systems and dynamics: bonus power if you use only basic
  spells for a long time, a staffless bonus, a "naked" bonus.
- Some things allow you to not be a wizard at all.
- Being OP is attainable by everyone and has a clear, meaningful path.

## Story

You are 18 (m/f). Your grandpa taught you four spells when you were a kid
(the tutorial). You accidentally found a spell that opens a portal to a
world (name TBD). You like it and stay, because grandpa died and you have
no family. Somewhere there is a dream where you are an OP wizard on a
rampage: that scene is made after the game is done, to showcase it quickly.

## Basic spells (every character gets them)

1. Fireball
2. Minor heal
3. Thunderstrike
4. Gust of wind: +movement speed and dodge chance for 2 seconds

*Engine: these are rows in `assets/data/items.json` and keywords in
`spells.json`. The build currently has fireball, small heal, minor teleport
and barrier as the four; swapping to this list is a data change.*

## Stats

- HP 100
- Mana 200
- Power
- Wisdom: age. Creates survivability spells; potions make you older or
  younger. Dead at year 126, 131 days, at a random hour. Age is not real
  time.
- Def
- Luck
- Movespeed

## Gear slots

Amulet, Lucky item, Hat, Chest, Pants, Boots, Cape, Ring, Weapon, Tome,
Staff (dual wielding, unique OP stuff). Gimmicks like a fireball minigun: a
0-mana-cost fireball shooter. What could go wrong.

## Knowledge (the private wiki book; the character comes with a copy)

- Grimoire
- Diary
- Bestiary
- Stories
- Mount

## Spells

A spell has these stats:

1. Mana cost
2. Cast time
3. Cooldown
4. Damage
5. Spell amp
6. Radius
7. Fail rate (yes, a cast can fail)
8. Misfire
9. Continuous yes/no
10. Free fire yes/no

The grimoire is a permanent item that holds all of them and can be opened
to change what is on the toggle bar. So you can cast everything: stand
there, list the pages and swap spells if everything is on cooldown. A spell
is added to the grimoire when the crafted scroll is used.

Voicecast is available, gives a buff, and must start with "I cast". A
little gimmick, but a way to stack buffs.

### Crafting example

- Lightning in a bottle: projectile
- plus cloud / thundercloud: free use, continuous
- plus rosemary: +radius, -misfire
- plus some candle: +10% spell amp, -10% cooldown
- plus eternal flame: burn continuously, plus future burn amp
- plus gladiator spear: projectile
- plus circle of life: +50% AoE

Gives a name if picked: "Thundercloud of eternal spears". And that has to
be tankable by someone.

*Engine: this is exactly the keyword model in `game/spells`: each item
contributes to shared stats, synergies name and tweak combinations. Spell
stats 2, 3, 5, 7, 8, 9, 10 are not in the resolver yet.*

### The 12 fireballs

Fireball, Waterbolt, Air blast, Cannon ball, Boulder, Ice shard, Demon's
breath, Angel's light, Pinball, Photosphere, Lightning ball (made by hitting
an orb with a thunderstrike), Prismatic punch. Together they make the
Ultimate Fireball. Twelve ultimate fireballs create a spell called
Singularity.

## Weapons

- All weapons buff spells and are used to cast them.
- Unique weapons take in specific spells, which they buff.
- Certain weapons are spells themselves.
- Basic magic wand: takes one basic spell and casts it cheaper and quicker.
  Not bad throughout the game.
- Trident: takes 3 spells and fires them simultaneously.
- Quickspell wand: 1 spell, reduces mana cost and cooldown.

## Crafting

- 18 unique herbs, into a dehydrator (made from general crafting).
- Bonuses based on harvest requirements (picked during events, e.g. full
  moon).
- Categories: Gear, Ritual, Spell, Orb, Potion, Bottle, Glass.
- Unique materials and enchantments for crafting boats and staffs. Same wood
  for a staff or a boat, if the log is big enough. Cutting trees is its own
  dynamic.
- TBD: the whole crafting system has some glue missing.

## Rituals

Most bosses, portals and worlds are done through rituals with their own
recipes. A ritual is a crafting grid of four squares: item on top, blood on
the right, candle on the left, and an enchantment on the bottom, for an
event of better loot chance or the super rare "everyone gets a unique item"
enchantment.

*Engine: distinct from the 12-slot spell-binding circle; both exist. The
ritual grid is a second station type with a 4-slot recipe matcher.*

## Bosses

Killing a boss increases mana, gives bonus rewards on specific events, and
levels up the spell that killed it (a permanent buff to that specific
spell). Unique loot, spell ingredients, weapons.

- Every boss drops blood, which is why a bottle is essential to always
  carry; the same bottle serves potions and all crafting including glass,
  except some machinery.
- Every boss drops literal body parts: Head of Medusa, Hydra scales...
- There is always a unique drop that only one player can take even if the
  party killed it.
- Mythicals have real-world interactions: the werewolf during a full moon;
  a vampire fought during the day is enraged and impossible to beat unless
  you survive long enough on pure defence, because of the sun.

Boss list: Medusa, Gladiator, Minotaur, Zeus, Krampus, Witch Supreme,
Goblin King, Dice Master, Jester, Prince, Dark & White Knight, Gilgamesh,
Balthus, Plague Doctor, Big Brain, some classic horror, Cerberus, Ghosts of
the Past, Hydra.

## World

- Portal-based exploration; world exploration.
- Giant hub with minigames that give rewards, usually to jumpstart skilled
  players who are new. Think old Minecraft multiplayer server hubs; those
  had soul.

### Hub

Free roam, Casino, PvP arena, Tarot shop, Moon shrine, Guild house (maybe),
Library (serves as the wiki, literally has encyclopedias), Auction house,
Shops.

### Shops

- Seller: used only to sell to him; he sells only what was sold to him.
  Items are real and counted.
- Clothes
- Materials

## See also

The campus to-do (buildings, sprites, UI, sound) is in the project doc
"Wizard game to-do". `docs/ARCHITECTURE.md` says how the engine is laid out.
