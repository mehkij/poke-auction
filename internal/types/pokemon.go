package types

// PokemonMap represents the top-level JSON structure where pokemon name is the key
type PokemonMap map[string]*Pokemon

type StatMap map[string]int

type Pokemon struct {
	Name   string `json:"name"`
	Sprite string
}
