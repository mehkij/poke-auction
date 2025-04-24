package types

type AuctionState struct {
	StopSignal      chan bool
	NominationOrder []*Player
	NominationPhase bool
	GenNumber       int
	ChannelID       string
}
