package types

import (
	"github.com/mehkij/poke-auction/internal/database"
	"github.com/mehkij/poke-auction/internal/dispatcher"
)

type GlobalConfig struct {
	GlobalDispatcher *dispatcher.Dispatcher
	Queries          *database.Queries
}
