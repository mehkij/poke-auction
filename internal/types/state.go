package types

type AuctionState struct {
	StopSignal       chan bool
	NominationOrder  []*Player
	NominationPhase  bool
	NominatedPokemon *Pokemon
	BiddingPhase     bool
	GenNumber        int
	ChannelID        string
}
