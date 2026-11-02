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
- Real Aseprite art in place of `tools/mkplaceholders` output, drawn to
  `artpack/README.md` (start with the player wizard)
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
