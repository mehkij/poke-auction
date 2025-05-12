package types

type AuctionState struct {
	StopSignal       chan bool
	GenNumber        int
	ChannelID        string
	BalanceMessageID string

	CurrentNominator    int
	NominationOrder     []*Player
	NominationPhase     bool
	NominatedPokemon    *Pokemon
	PreviouslyNominated []string

	BiddingPhase  bool
	BidSoFar      map[string]int
	ProcessingBid bool
	HighestBid    int
}
