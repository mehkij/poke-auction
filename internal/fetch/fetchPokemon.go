package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mehkij/poke-auction/internal/types"
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

	pokemon, exists := pokemonMap[name]
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found", name)
	}

	pokemon.Name = strings.ToUpper(name[:1]) + name[1:]

	return pokemon, nil
}
