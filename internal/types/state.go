package types

import "sync"

type AuctionState struct {
	Participants   map[string]*Player // key: UserID
	AuctionStateMu sync.Mutex

	StopSignal       chan bool
	GenNumber        int
	ChannelID        string
	BalanceMessageID string
	NatDexEnabled    bool

	CurrentNominator    int
	NominationOrder     []*Player
	NominationPhase     bool
	NominatedPokemon    *Pokemon
	PreviouslyNominated []string

	BiddingPhase  bool
	BidSoFar      map[string]int // key: UserID
	ProcessingBid bool
	HighestBid    int
}
