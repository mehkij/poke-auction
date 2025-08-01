package fetch

import (
	"math/rand"
	"time"

	"github.com/mehkij/poke-auction/internal/types"
)

type BabyPokemon struct {
	types.Pokemon
	Generation int
}

func RollRandomBabyPokemon(currentTeam []*types.Pokemon, gen int) []*types.Pokemon {
	remaining := 6 - len(currentTeam)
	// #nosec G404 -- using math/rand is fine as rolling for baby Pokemon is not security-critical
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	alreadyPicked := make(map[string]bool)

	for _, p := range currentTeam {
		alreadyPicked[p.Name] = true
	}

	for remaining > 0 {
		roll := rng.Intn(len(AllBabyPokemon))
		name := AllBabyPokemon[roll].Name

		if alreadyPicked[name] || AllBabyPokemon[roll].Generation > gen {
			continue
		}

		alreadyPicked[name] = true
		currentTeam = append(currentTeam, &types.Pokemon{
			Name:   name,
			Sprite: AllBabyPokemon[roll].Sprite,
		})
		remaining--
	}

	return currentTeam
}

// Hard-coded values of all baby Pokemon to be used in
var AllBabyPokemon = []BabyPokemon{
	// Generation 2
	{
		Pokemon: types.Pokemon{
			Name: "Pichu",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Cleffa",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Igglybuff",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Togepi",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Tyrogue",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Smoochum",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Elekid",
		},
		Generation: 2,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Magby",
		},
		Generation: 2,
	},

	// Generation 3
	{
		Pokemon: types.Pokemon{
			Name: "Azurill",
		},
		Generation: 3,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Wynaut",
		},
		Generation: 3,
	},

	// Generation 4
	{
		Pokemon: types.Pokemon{
			Name: "Budew",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Chingling",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Bonsly",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Mime Jr.",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Happiny",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Munchlax",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Riolu",
		},
		Generation: 4,
	},
	{
		Pokemon: types.Pokemon{
			Name: "Mantyke",
		},
		Generation: 4,
	},

	// Generation 8
	{
		Pokemon: types.Pokemon{
			Name: "Toxel",
		},
		Generation: 8,
	},
}
