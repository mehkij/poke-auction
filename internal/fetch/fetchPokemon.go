package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/mehkij/poke-auction/internal/types"
	"github.com/mehkij/poke-auction/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func FetchPokemon(gen int, name string) (*types.Pokemon, error) {
	// Validate name to contain only allowed characters (letters, numbers, hyphens)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\-\s]+$`)
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("invalid pokemon name: %s", name)
	}

	// Even though the randbat data isn't used, it helps validate that a Pokemon is useable in a certain generation.
	u, _ := url.Parse("https://pkmn.github.io/randbats/data/gen")
	u.Path += fmt.Sprintf("%drandombattle.json", gen)
	res, err := http.Get(u.String())

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", res.StatusCode)
	}

	var pokemonMap types.PokemonMap
	if err := json.NewDecoder(res.Body).Decode(&pokemonMap); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	titleCase := cases.Title(language.English)
	pokemonName := titleCase.String(name)
	pokemon, exists := pokemonMap[pokemonName]
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found", name)
	}

	// Replace whitespace with dashes, as PokeAPI doesn't support spaces in Pokemon names
	var validatedName string
	if strings.Contains(name, " ") {
		validatedName = strings.ReplaceAll(pokemonName, " ", "-")
	}

	sprite, err := FetchPokemonImage(gen, validatedName)
	if err != nil {
		return nil, err
	}

	pokemon.Name = pokemonName
	pokemon.Sprite = sprite

	return pokemon, nil
}

func FetchPokemonImage(gen int, name string) (string, error) {
	// Validate name to contain only allowed characters (letters, numbers, hyphens)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid pokemon name: %s", name)
	}

	u, _ := url.Parse("https://pokeapi.co/api/v2/pokemon/")
	u.Path += name + "/"
	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request to PokeAPI failed with status code: %d", res.StatusCode)
	}

	var data map[string]any

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	sprites := data["sprites"].(map[string]any)
	var front_default string

	if gen == 9 {
		front_default = sprites["front_default"].(string)
	} else {
		genNum := fmt.Sprintf("generation-%s", utils.ToRoman(gen))
		versions := sprites["versions"].(map[string]any)
		generation := versions[genNum].(map[string]any)

		var spriteMap map[string]any
		for _, v := range generation {
			spriteMap = v.(map[string]any)
			break
		}

		front_default = spriteMap["front_default"].(string)
	}

	return front_default, nil
}
