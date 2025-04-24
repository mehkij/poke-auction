package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mehkij/poke-auction/internal/types"
	"github.com/mehkij/poke-auction/internal/utils"
)

func FetchPokemon(gen int, name string) (*types.Pokemon, error) {
	url := "https://pkmn.github.io/randbats/data/gen" + fmt.Sprint(gen) + "randombattle.json"
	res, err := http.Get(url)
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

	pokemonName := strings.ToUpper(name[:1]) + name[1:]

	pokemon, exists := pokemonMap[pokemonName]
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found", name)
	}

	sprite, err := FetchPokemonImage(gen, pokemonName)
	if err != nil {
		return nil, err
	}

	pokemon.Name = pokemonName
	pokemon.Sprite = sprite

	return pokemon, nil
}

func FetchPokemonImage(gen int, name string) (string, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + name + "/"
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", res.StatusCode)
	}

	var data map[string]any

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	genNum := fmt.Sprintf("generation-%s", utils.ToRoman(gen))

	sprites := data["sprites"].(map[string]any)
	versions := sprites["versions"].(map[string]any)
	generation := versions[genNum].(map[string]any)

	var spriteMap map[string]any
	for _, v := range generation {
		spriteMap = v.(map[string]any)
		break
	}

	front_default := spriteMap["front_default"].(string)

	return front_default, nil
}
