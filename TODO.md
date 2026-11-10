# Todo

## Now

- Step 4: the first things the design notes need after binding and casting
  (`docs/DESIGN.md`, `docs/ARCHITECTURE.md` known gaps): player stats beyond
  HP and mana (power, wisdom/age, def, luck) and gear slots, element effects
  beyond the tint, and something that attacks so dodge and heal matter
- Run make configure-repo once

## Next

- Ritual grid station (4 slots: item, blood, candle, enchantment), distinct
  from the 12-slot binding circle
- Doors: let the campus facades open once a building has something inside
- Replace the 12px stand-in bitmap font with the 5x7 pixel font once UI art
  exists
- More art, in the art pack's order: the old wizard and NPCs at 16x40,
  portraits, the 47-tile autotile sets for grass, path, plaza and water (the
  engine needs autotiling first), the rest of the campus props, item icons,
  element recolours of the fireball
- Play the art that exists but is not wired: `hurt`, `meditate`, barrier
  `break`, footstep `puff`
- Audio and gamepad bindings
- Bring the "Wizard game to-do" project doc that `docs/DESIGN.md` points at
  into the repo, or link it

## Later

- Map editor import once the hub outgrows maps built in code
- An authoritative server running the same fixed-step `Update` for several
  wizards (PORT 4160 is reserved; nothing listens yet)
- The dream sequence showcase, after the game is done

## Done

- 0.1.0: engine, campus hub, dialogue and inventory, ritual circle binding,
  grimoire, casting, training dummies, design notes
