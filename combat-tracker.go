package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Combatant represents any entity in combat (player or monster)
type Combatant struct {
	Name          string   `json:"name"`
	Initiative    int      `json:"initiative"`
	MaxHP         int      `json:"maxHP"`
	CurrentHP     int      `json:"currentHP"`
	IsPlayer      bool     `json:"isPlayer"`
	IsConscious   bool     `json:"isConscious"`
	TemporaryHP   int      `json:"temporaryHP"`
	StatusEffects []string `json:"statusEffects"`
}

// CombatTracker manages the combat encounter
type CombatTracker struct {
	Combatants     []Combatant `json:"combatants"`
	Round          int         `json:"round"`
	CurrentTurnIdx int         `json:"currentTurnIdx"`
	IsActive       bool        `json:"isActive"`
	CampaignName   string      `json:"campaignName"`
	EncounterName  string      `json:"encounterName"`
	SaveFilePath   string      `json:"-"`             // Track the save file path but don't include in JSON
	StatusEffects  []string    `json:"statusEffects"` // List of available status effects
}

// SaveState represents the full state for saving/loading
type SaveState struct {
	CombatTracker CombatTracker `json:"combatTracker"`
	SaveTime      string        `json:"saveTime"`
	Version       string        `json:"version"`
}

// NewCombatTracker creates a new combat tracker
func NewCombatTracker() *CombatTracker {
	// Define default status effects
	defaultEffects := []string{
		"Blinded",
		"Charmed",
		"Deafened",
		"Frightened",
		"Grappled",
		"Incapacitated",
		"Invisible",
		"Paralyzed",
		"Petrified",
		"Poisoned",
		"Prone",
		"Restrained",
		"Stunned",
		"Unconscious",
		"Custom Status Effect",
	}

	return &CombatTracker{
		Combatants:     []Combatant{},
		Round:          0,
		CurrentTurnIdx: -1,
		IsActive:       false,
		CampaignName:   "Default Campaign",
		EncounterName:  "Unknown Encounter",
		SaveFilePath:   "",
		StatusEffects:  defaultEffects,
	}
}

// AddCombatant adds a new combatant to the encounter
func (ct *CombatTracker) AddCombatant(name string, initiative, maxHP int, isPlayer bool) {
	combatant := Combatant{
		Name:          name,
		Initiative:    initiative,
		MaxHP:         maxHP,
		CurrentHP:     maxHP,
		IsPlayer:      isPlayer,
		IsConscious:   true,
		TemporaryHP:   0,
		StatusEffects: []string{},
	}
	ct.Combatants = append(ct.Combatants, combatant)

	// Auto-save state
	ct.AutoSave()
}

// SortByInitiative sorts combatants by initiative (highest first)
func (ct *CombatTracker) SortByInitiative() {
	sort.Slice(ct.Combatants, func(i, j int) bool {
		return ct.Combatants[i].Initiative > ct.Combatants[j].Initiative
	})
}

// StartCombat begins the combat encounter
func (ct *CombatTracker) StartCombat() {
	if len(ct.Combatants) == 0 {
		fmt.Println("Cannot start combat with no combatants!")
		return
	}

	ct.SortByInitiative()
	ct.Round = 1
	ct.CurrentTurnIdx = 0
	ct.IsActive = true

	fmt.Println("\n===== COMBAT BEGINS =====")
	fmt.Printf("Round %d\n", ct.Round)
	fmt.Printf("It's %s's turn!\n", ct.Combatants[ct.CurrentTurnIdx].Name)

	// Auto-save state
	ct.AutoSave()
}

// NextTurn advances to the next combatant's turn
func (ct *CombatTracker) NextTurn() {
	if !ct.IsActive {
		fmt.Println("Combat hasn't started yet!")
		return
	}

	ct.CurrentTurnIdx++
	if ct.CurrentTurnIdx >= len(ct.Combatants) {
		ct.Round++
		ct.CurrentTurnIdx = 0
		fmt.Printf("\n===== ROUND %d =====\n", ct.Round)
	}

	fmt.Printf("It's %s's turn!\n", ct.Combatants[ct.CurrentTurnIdx].Name)

	// Auto-save state
	ct.AutoSave()
}

// AdjustHP changes a combatant's hit points
func (ct *CombatTracker) AdjustHP(index int, amount int) {
	if index < 0 || index >= len(ct.Combatants) {
		fmt.Println("Invalid combatant index!")
		return
	}

	c := &ct.Combatants[index]

	// Handle damage
	if amount < 0 {
		damage := -amount

		// Apply temporary HP first
		if c.TemporaryHP > 0 {
			if damage <= c.TemporaryHP {
				c.TemporaryHP -= damage
				damage = 0
			} else {
				damage -= c.TemporaryHP
				c.TemporaryHP = 0
			}
		}

		// Apply remaining damage to current HP
		if damage > 0 {
			c.CurrentHP -= damage
		}

		// Check if unconscious
		if c.CurrentHP <= 0 {
			c.CurrentHP = 0
			c.IsConscious = false
			fmt.Printf("%s falls unconscious!\n", c.Name)
		}
	} else {
		// Handle healing
		c.CurrentHP += amount
		if c.CurrentHP > c.MaxHP {
			c.CurrentHP = c.MaxHP
		}
		if !c.IsConscious && c.CurrentHP > 0 {
			c.IsConscious = true
			fmt.Printf("%s regains consciousness!\n", c.Name)
		}
	}

	fmt.Printf("%s HP: %d/%d", c.Name, c.CurrentHP, c.MaxHP)
	if c.TemporaryHP > 0 {
		fmt.Printf(" (Temp: %d)", c.TemporaryHP)
	}
	fmt.Println()

	// Auto-save state
	ct.AutoSave()
}

// AddTemporaryHP adds temporary hit points to a combatant
func (ct *CombatTracker) AddTemporaryHP(index int, amount int) {
	if index < 0 || index >= len(ct.Combatants) {
		fmt.Println("Invalid combatant index!")
		return
	}

	c := &ct.Combatants[index]

	// Temporary HP doesn't stack, take the higher value
	if amount > c.TemporaryHP {
		c.TemporaryHP = amount
		fmt.Printf("%s now has %d temporary hit points!\n", c.Name, c.TemporaryHP)
	} else {
		fmt.Printf("%s already has %d temporary hit points, which is higher!\n", c.Name, c.TemporaryHP)
	}

	// Auto-save state
	ct.AutoSave()
}

// AddStatusEffect adds a status effect to a combatant
func (ct *CombatTracker) AddStatusEffect(index int, effect string) {
	if index < 0 || index >= len(ct.Combatants) {
		fmt.Println("Invalid combatant index!")
		return
	}

	c := &ct.Combatants[index]
	c.StatusEffects = append(c.StatusEffects, effect)
	fmt.Printf("%s is now affected by: %s\n", c.Name, effect)

	// Auto-save state
	ct.AutoSave()
}

// RemoveStatusEffect removes a status effect from a combatant
func (ct *CombatTracker) RemoveStatusEffect(index int, effect string) {
	if index < 0 || index >= len(ct.Combatants) {
		fmt.Println("Invalid combatant index!")
		return
	}

	c := &ct.Combatants[index]

	for i, e := range c.StatusEffects {
		if e == effect {
			// Remove the effect by replacing it with the last element and then truncating
			c.StatusEffects[i] = c.StatusEffects[len(c.StatusEffects)-1]
			c.StatusEffects = c.StatusEffects[:len(c.StatusEffects)-1]
			fmt.Printf("%s is no longer affected by: %s\n", c.Name, effect)

			// Auto-save state
			ct.AutoSave()
			return
		}
	}

	fmt.Printf("%s was not affected by: %s\n", c.Name, effect)
}

// DisplayCombatState shows the current state of all combatants
func (ct *CombatTracker) DisplayCombatState() {
	fmt.Println("\n===== COMBAT STATE =====")
	if ct.CampaignName != "" || ct.EncounterName != "" {
		fmt.Printf("Campaign: %s | Encounter: %s\n", ct.CampaignName, ct.EncounterName)
	}
	if ct.SaveFilePath != "" {
		fmt.Printf("Auto-saving to: %s\n", ct.SaveFilePath)
	}
	fmt.Printf("Round: %d\n", ct.Round)
	fmt.Println("-------------------")

	for i, c := range ct.Combatants {
		currentTurnMarker := " "
		if i == ct.CurrentTurnIdx && ct.IsActive {
			currentTurnMarker = "→"
		}

		statusStr := ""
		if len(c.StatusEffects) > 0 {
			statusStr = fmt.Sprintf(" [%s]", strings.Join(c.StatusEffects, ", "))
		}

		consciousnessStr := ""
		if !c.IsConscious {
			consciousnessStr = " (Unconscious)"
		}

		tempHPStr := ""
		if c.TemporaryHP > 0 {
			tempHPStr = fmt.Sprintf(" (Temp: %d)", c.TemporaryHP)
		}

		playerMarker := " "
		if c.IsPlayer {
			playerMarker = "P"
		} else {
			playerMarker = "M"
		}

		fmt.Printf("%s %s %2d. %-20s Init: %2d HP: %3d/%-3d%s%s%s\n",
			currentTurnMarker, playerMarker, i+1, c.Name, c.Initiative,
			c.CurrentHP, c.MaxHP, tempHPStr, consciousnessStr, statusStr)
	}
	fmt.Println("-------------------")
}

// EndCombat ends the current combat
func (ct *CombatTracker) EndCombat() {
	if !ct.IsActive {
		fmt.Println("No active combat to end!")
		return
	}

	ct.IsActive = false
	fmt.Println("\n===== COMBAT ENDED =====")

	// Display final combat state
	ct.DisplayCombatState()

	// Auto-save state
	ct.AutoSave()
}

// SaveToFile saves the current combat state to a file
func (ct *CombatTracker) SaveToFile(filename string) error {
	// Create a save state object
	saveState := SaveState{
		CombatTracker: *ct,
		SaveTime:      time.Now().Format(time.RFC3339),
		Version:       "1.0.0",
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(saveState, "", "    ")
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	// Write to file
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("Combat state saved to %s\n", filename)
	return nil
}

// AutoSave automatically saves the current state to the configured save file
func (ct *CombatTracker) AutoSave() {
	if ct.SaveFilePath == "" {
		return // No auto-save file configured
	}

	err := ct.SaveToFile(ct.SaveFilePath)
	if err != nil {
		fmt.Printf("Auto-save failed: %v\n", err)
	}
}

// LoadFromFile loads a combat state from a file
func LoadFromFile(filename string) (*CombatTracker, error) {
	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse JSON
	var saveState SaveState
	err = json.Unmarshal(data, &saveState)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Set the save file path for auto-save
	saveState.CombatTracker.SaveFilePath = filename

	// If no status effects are present, initialize with defaults
	if len(saveState.CombatTracker.StatusEffects) == 0 {
		// Define default status effects
		defaultEffects := []string{
			"Blinded",
			"Charmed",
			"Deafened",
			"Frightened",
			"Grappled",
			"Incapacitated",
			"Invisible",
			"Paralyzed",
			"Petrified",
			"Poisoned",
			"Prone",
			"Restrained",
			"Stunned",
			"Unconscious",
			"Custom Status Effect",
		}
		saveState.CombatTracker.StatusEffects = defaultEffects
		fmt.Println("Initialized default status effects list.")
	}

	fmt.Printf("Loaded save from: %s\n", saveState.SaveTime)
	return &saveState.CombatTracker, nil
}

// SetEncounterDetails sets campaign and encounter names
func (ct *CombatTracker) SetEncounterDetails(campaignName, encounterName string) {
	ct.CampaignName = campaignName
	ct.EncounterName = encounterName
	fmt.Printf("Set encounter details - Campaign: %s, Encounter: %s\n", campaignName, encounterName)

	// Auto-save state
	ct.AutoSave()
}

// ClearScreen clears the terminal (platform dependent)
func ClearScreen() {
	fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen
}

// DisplayMenuHorizontal displays the menu options horizontally
func DisplayMenuHorizontal() {
	fmt.Println("\n====================== COMMANDS ======================")
	fmt.Print("1:Add  2:Start  3:Next  4:HP  5:TempHP  6:AddStatus  ")
	fmt.Println("7:RemStatus  8:Display  9:End  10:Details  11:Save  12:Load  0:Exit")
	fmt.Println("======================================================")
}

func main() {
	var ct *CombatTracker

	// Check if a save file was provided as a command-line argument
	if len(os.Args) > 1 {
		saveFilePath := os.Args[1]
		var err error

		// Try to load the file
		ct, err = LoadFromFile(saveFilePath)
		if err != nil {
			fmt.Printf("Failed to load save file: %v\n", err)
			fmt.Println("Creating a new combat tracker instead.")
			ct = NewCombatTracker()
			ct.SaveFilePath = saveFilePath // Set for future auto-saves
		}
	} else {
		// No save file provided, start fresh
		ct = NewCombatTracker()
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("===== D&D COMBAT TRACKER =====")
	if ct.SaveFilePath != "" {
		fmt.Printf("Auto-saving enabled to: %s\n", ct.SaveFilePath)
	}

	for {
		// Always display the current combat state
		ct.DisplayCombatState()

		// Display menu options horizontally
		DisplayMenuHorizontal()

		fmt.Print("\nEnter command: ")
		scanner.Scan()
		cmd := scanner.Text()

		ClearScreen() // Clear screen before processing command

		switch cmd {
		case "1": // Add Combatant
			var name string
			var initiative, hp int
			var isPlayer string

			fmt.Println("=== ADD COMBATANT ===")
			fmt.Print("Enter name: ")
			scanner.Scan()
			name = scanner.Text()

			fmt.Print("Enter initiative: ")
			scanner.Scan()
			initiative, _ = strconv.Atoi(scanner.Text())

			fmt.Print("Enter max HP: ")
			scanner.Scan()
			hp, _ = strconv.Atoi(scanner.Text())

			fmt.Print("Is this a player? (y/n): ")
			scanner.Scan()
			isPlayer = strings.ToLower(scanner.Text())

			ct.AddCombatant(name, initiative, hp, isPlayer == "y" || isPlayer == "yes")
			fmt.Printf("Added %s to combat with initiative %d and %d HP\n", name, initiative, hp)

		case "2": // Start Combat
			ct.StartCombat()

		case "3": // Next Turn
			ct.NextTurn()

		case "4": // Damage/Heal
			var index, amount int

			fmt.Println("=== ADJUST HIT POINTS ===")
			ct.DisplayCombatState()

			// If combat is active, show current player and use as default
			if ct.IsActive && ct.CurrentTurnIdx >= 0 {
				currentPlayer := ct.Combatants[ct.CurrentTurnIdx]
				fmt.Printf("Current player: %s (index: %d)\n", currentPlayer.Name, ct.CurrentTurnIdx+1)
			}

			fmt.Print("Enter combatant number (press Enter for current player): ")
			scanner.Scan()
			indexStr := scanner.Text()

			if indexStr == "" && ct.IsActive && ct.CurrentTurnIdx >= 0 {
				// Use current player if no input and combat is active
				index = ct.CurrentTurnIdx
			} else {
				index, _ = strconv.Atoi(indexStr)
				index-- // Convert to 0-based index
			}

			fmt.Print("Enter amount (+heal, -damage): ")
			scanner.Scan()
			amount, _ = strconv.Atoi(scanner.Text())

			ct.AdjustHP(index, amount)

		case "5": // Add Temporary HP
			var index, amount int

			fmt.Println("=== ADD TEMPORARY HP ===")
			ct.DisplayCombatState()

			// If combat is active, show current player and use as default
			if ct.IsActive && ct.CurrentTurnIdx >= 0 {
				currentPlayer := ct.Combatants[ct.CurrentTurnIdx]
				fmt.Printf("Current player: %s (index: %d)\n", currentPlayer.Name, ct.CurrentTurnIdx+1)
			}

			fmt.Print("Enter combatant number (press Enter for current player): ")
			scanner.Scan()
			indexStr := scanner.Text()

			if indexStr == "" && ct.IsActive && ct.CurrentTurnIdx >= 0 {
				// Use current player if no input and combat is active
				index = ct.CurrentTurnIdx
			} else {
				index, _ = strconv.Atoi(indexStr)
				index-- // Convert to 0-based index
			}

			// Validate index
			if index < 0 || index >= len(ct.Combatants) {
				fmt.Println("Invalid combatant index!")
				continue
			}

			fmt.Print("Enter temporary HP amount: ")
			scanner.Scan()
			amount, _ = strconv.Atoi(scanner.Text())

			ct.AddTemporaryHP(index, amount)

		case "6": // Add Status Effect
			var index int
			var effect string

			fmt.Println("=== ADD STATUS EFFECT ===")
			ct.DisplayCombatState()

			// If combat is active, show current player and use as default
			if ct.IsActive && ct.CurrentTurnIdx >= 0 {
				currentPlayer := ct.Combatants[ct.CurrentTurnIdx]
				fmt.Printf("Current player: %s (index: %d)\n", currentPlayer.Name, ct.CurrentTurnIdx+1)
			}

			fmt.Print("Enter combatant number (press Enter for current player): ")
			scanner.Scan()
			indexStr := scanner.Text()

			if indexStr == "" && ct.IsActive && ct.CurrentTurnIdx >= 0 {
				// Use current player if no input and combat is active
				index = ct.CurrentTurnIdx
			} else {
				index, _ = strconv.Atoi(indexStr)
				index-- // Convert to 0-based index
			}

			// Validate index
			if index < 0 || index >= len(ct.Combatants) {
				fmt.Println("Invalid combatant index!")
				continue
			}

			fmt.Printf("\nSelect status effect for %s:\n", ct.Combatants[index].Name)
			fmt.Println("0. Custom Status Effect")
			for i, effect := range ct.StatusEffects {
				fmt.Printf("%d. %s\n", i+1, effect)
			}

			fmt.Print("\nEnter number of status effect (or 0 for custom): ")
			scanner.Scan()
			effectIndex, err := strconv.Atoi(scanner.Text())
			if err != nil || effectIndex < 0 || effectIndex > len(ct.StatusEffects) {
				fmt.Println("Invalid selection!")
				continue
			}

			// If custom status effect was selected
			if effectIndex == 0 {
				fmt.Print("Enter custom status effect name: ")
				scanner.Scan()
				effect = scanner.Text()
			} else {
				effect = ct.StatusEffects[effectIndex-1]
			}

			ct.AddStatusEffect(index, effect)

		case "7": // Remove Status Effect
			var index int
			var effect string

			fmt.Println("=== REMOVE STATUS EFFECT ===")
			ct.DisplayCombatState()

			// If combat is active, show current player and use as default
			if ct.IsActive && ct.CurrentTurnIdx >= 0 {
				currentPlayer := ct.Combatants[ct.CurrentTurnIdx]
				fmt.Printf("Current player: %s (index: %d)\n", currentPlayer.Name, ct.CurrentTurnIdx+1)
			}

			fmt.Print("Enter combatant number (press Enter for current player): ")
			scanner.Scan()
			indexStr := scanner.Text()

			if indexStr == "" && ct.IsActive && ct.CurrentTurnIdx >= 0 {
				// Use current player if no input and combat is active
				index = ct.CurrentTurnIdx
			} else {
				index, _ = strconv.Atoi(indexStr)
				index-- // Convert to 0-based index
			}

			// Validate index
			if index < 0 || index >= len(ct.Combatants) {
				fmt.Println("Invalid combatant index!")
				continue
			}

			// Show current status effects
			combatant := ct.Combatants[index]
			if len(combatant.StatusEffects) == 0 {
				fmt.Printf("%s has no status effects to remove.\n", combatant.Name)
				continue
			}

			fmt.Printf("\nCurrent status effects for %s:\n", combatant.Name)
			for i, effect := range combatant.StatusEffects {
				fmt.Printf("%d. %s\n", i+1, effect)
			}

			fmt.Print("\nEnter number of status effect to remove: ")
			scanner.Scan()
			effectIndex, err := strconv.Atoi(scanner.Text())
			if err != nil || effectIndex < 1 || effectIndex > len(combatant.StatusEffects) {
				fmt.Println("Invalid selection!")
				continue
			}

			// Get the effect name from the selected index
			effect = combatant.StatusEffects[effectIndex-1]
			ct.RemoveStatusEffect(index, effect)

		case "8": // Display Combat State
			// State is already displayed at top of loop
			fmt.Println("Combat state refreshed.")

		case "9": // End Combat
			ct.EndCombat()

		case "10": // Set Encounter Details
			var campaign, encounter string

			fmt.Println("=== SET ENCOUNTER DETAILS ===")
			fmt.Print("Enter campaign name: ")
			scanner.Scan()
			campaign = scanner.Text()

			fmt.Print("Enter encounter name: ")
			scanner.Scan()
			encounter = scanner.Text()

			ct.SetEncounterDetails(campaign, encounter)

		case "11": // Save Combat State
			var filename string

			fmt.Println("=== SAVE COMBAT STATE ===")

			defaultFilename := ct.SaveFilePath
			if defaultFilename == "" {
				// Generate a default filename with timestamp
				defaultFilename = fmt.Sprintf("combat_%s_%s_%s.json",
					strings.ReplaceAll(ct.CampaignName, " ", "_"),
					strings.ReplaceAll(ct.EncounterName, " ", "_"),
					time.Now().Format("2006-01-02_15-04-05"))
			}

			fmt.Printf("Enter filename (default: %s): ", defaultFilename)
			scanner.Scan()
			filename = scanner.Text()

			if filename == "" {
				filename = defaultFilename
			}

			// Add .json extension if not present
			if !strings.HasSuffix(filename, ".json") {
				filename += ".json"
			}

			err := ct.SaveToFile(filename)
			if err != nil {
				fmt.Printf("Error saving: %v\n", err)
			} else {
				// Update the save file path for auto-saves
				ct.SaveFilePath = filename
			}

		case "12": // Load Combat State
			var filename string

			fmt.Println("=== LOAD COMBAT STATE ===")
			fmt.Print("Enter filename to load: ")
			scanner.Scan()
			filename = scanner.Text()

			loadedCT, err := LoadFromFile(filename)
			if err != nil {
				fmt.Printf("Error loading: %v\n", err)
			} else {
				ct = loadedCT
				fmt.Println("Combat state loaded successfully!")
			}

		case "0": // Exit
			fmt.Println("Exiting D&D Combat Tracker. Farewell, adventurer!")

			// Perform one final auto-save before exiting
			if ct.SaveFilePath != "" {
				fmt.Printf("Performing final save to %s before exit.\n", ct.SaveFilePath)
				ct.AutoSave()
			}

			return

		default:
			fmt.Println("Invalid command. Please try again.")
		}
	}
}
