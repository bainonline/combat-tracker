# D&D Combat Tracker

A terminal-based combat tracking tool for Dungeons & Dragons game masters. Keep track of initiative order, hit points, status effects, and more during encounters.

## Features

- Manage player characters and monsters in combat
- Track initiative order automatically
- Record hit points, temporary HP, and status effects
- Auto-save feature to preserve combat state
- Load and save encounters to JSON files
- Track campaign and encounter names

## Installation

### Prerequisites

- Go 1.13 or higher

### Building from Source

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/dnd-combat-tracker.git
   cd dnd-combat-tracker
   ```

2. Build the executable:
   ```
   go build combat-tracker.go
   ```

## Usage

### Basic Usage

Run the tracker without arguments to start with a blank combat state:

```
./combat-tracker
```

### Auto-save Mode

Start the tracker with a file path to enable auto-saving:

```
./combat-tracker mysave.json
```

This will:
- Load the initial state from the file if it exists
- Create a new state if the file doesn't exist
- Automatically save to this file after every action

### Commands

The tracker provides a text-based interface with the following commands:

1. **Add Combatant**: Add a player or monster to the encounter
2. **Start Combat**: Begin the combat, sort by initiative
3. **Next Turn**: Advance to the next combatant's turn
4. **HP**: Adjust hit points (positive for healing, negative for damage)
5. **TempHP**: Add temporary hit points
6. **AddStatus**: Add a status effect (like "Poisoned" or "Stunned")
7. **RemStatus**: Remove a status effect
8. **Display**: Refresh the combat state display
9. **End**: End the current combat
10. **Details**: Set campaign and encounter names
11. **Save**: Manually save the state to a file
12. **Load**: Load a previously saved state
13. **Duplicate**: Create multiple copies of a combatant
14. **Change Initiative**: Update a combatant's initiative value
0. **Exit**: Quit the application

## Combat Display

The combat tracker displays combatants with the following information:

```
→ P  1. Gandalf            Init: 18 HP:  75/75
  M  2. Goblin Chief       Init: 15 HP:  45/45
  P  3. Aragorn            Init: 14 HP:  60/60 (Temp: 5)
  M  4. Warg               Init: 12 HP:  30/30 [Poisoned]
  M  5. Orc                Init:  8 HP:   0/25 (Unconscious)
```

Legend:
- `→` indicates the current turn
- `P` indicates a player character
- `M` indicates a monster/NPC
- Numbers show initiative order
- Status effects are shown in brackets
- Unconscious combatants are marked
- Temporary HP is displayed when present

## File Format

Combat states are saved as JSON files with the following structure:

```json
{
    "combatTracker": {
        "combatants": [
            {
                "name": "Wizard",
                "initiative": 18,
                "maxHP": 45,
                "currentHP": 32,
                "isPlayer": true,
                "isConscious": true,
                "temporaryHP": 5,
                "statusEffects": ["Concentration"]
            },
            ...
        ],
        "round": 2,
        "currentTurnIdx": 1,
        "isActive": true,
        "campaignName": "Lost Mine of Phandelver",
        "encounterName": "Goblin Ambush"
    },
    "saveTime": "2025-03-30T14:32:25Z",
    "version": "1.0.0"
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the needs of D&D Dungeon Masters everywhere
- Built with love for the tabletop gaming community

## Example Combat Session

Here's an example of how to use each command in a typical combat session:

### 1. Adding Combatants
