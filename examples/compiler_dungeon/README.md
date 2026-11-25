# The Compiler Dungeon

A multi-file turn-based terminal RPG written in ScarScript v1.0.3.

## Structure
- `entities.scar` — `Hero` and `Boss` struct definitions
- `combat.scar` — Damage calculation, enrage/turn modifiers, and health bar formatting
- `main.scar` — Main game loop, action selection, and victory/crash state

## Run
```bash
scar run main.scar