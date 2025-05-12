package types

type AuctionState struct {
	StopSignal          chan bool
	CurrentNominator    int
	NominationOrder     []*Player
	NominationPhase     bool
	NominatedPokemon    *Pokemon
	PreviouslyNominated []string
	BiddingPhase        bool
	BidSoFar            map[string]int
	ProcessingBid       bool
	HighestBid          int
	GenNumber           int
	ChannelID           string
}
