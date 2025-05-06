package types

type AuctionState struct {
	StopSignal          chan bool
	CurrentNominator    int
	NominationOrder     []*Player
	NominationPhase     bool
	NominatedPokemon    *Pokemon
	PreviouslyNominated []string
	BiddingPhase        bool
	GenNumber           int
	ChannelID           string
}
