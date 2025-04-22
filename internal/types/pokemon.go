package types

type StatMap map[string]int

type Role struct {
	Abilities []string `json:"abilities,omitempty"`
	Items     []string `json:"items,omitempty"`
	Moves     []string `json:"moves"`
	TeraTypes []string `json:"teratypes,omitempty"`
	IVs       StatMap  `json:"ivs,omitempty"`
}

type Pokemon struct {
	Name      string          `json:"name"`
	Level     int             `json:"level"`
	Abilities []string        `json:"abilities,omitempty"`
	Items     []string        `json:"items,omitempty"`
	Moves     []string        `json:"moves,omitempty"`
	Roles     map[string]Role `json:"roles,omitempty"`
	IVs       StatMap         `json:"ivs,omitempty"`
}

// PokemonMap represents the top-level JSON structure where pokemon name is the key
type PokemonMap map[string]*Pokemon
