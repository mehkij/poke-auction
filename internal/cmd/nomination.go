package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/fetch"
	"github.com/mehkij/poke-auction/internal/types"
)

// "/nominate" command definition.
var NominateCommand = &Command{
	Name:        "nominate",
	Description: "Nominate a Pokemon.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "The name of the Pokemon.",
			Required:    true,
		},
	},
	Callback: NominateCallback,
}

// Returns the order in which players will nominate a Pokemon to be auctioned.
func RollNominationOrder() []*types.Player {
	remaining := make([]*types.Player, len(participants))
	copy(remaining, participants)

	var order []*types.Player
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for len(remaining) > 0 {
		roll := rng.Intn(len(remaining))

		order = append(order, remaining[roll])

		// Remove the selected player by replacing with last element and truncating
		remaining[roll] = remaining[len(remaining)-1]
		remaining = remaining[:len(remaining)-1]
	}

	return order
}

func NominationPhase(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(participants) == 0 {
		edit := &discordgo.MessageEdit{
			Channel: i.ChannelID,
			ID:      i.Message.ID,
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Error Starting Nomination Phase",
					Description: "Cannot start nomination phase with no participants!",
				},
			},
			Components: &[]discordgo.MessageComponent{},
		}

		_, err := s.ChannelMessageEditComplex(edit)
		return err
	}

	log.Println("Locking Mutex...")
	auctionStatesMu.Lock()
	state, exists := auctionStates[i.Message.ID]
	if !exists || len(state.NominationOrder) == 0 {
		return fmt.Errorf("no nomination order found")
	}

	log.Println("Incrementing CurrentNominator...")
	state.CurrentNominator++

	// Ensure that incrementing the pointer doesn't exceed length of array, and set the pointer to 0 if it does.
	if state.CurrentNominator >= len(state.NominationOrder) {
		state.CurrentNominator = 0
		log.Println("Pointer is equal to NominationOrder length, resetting pointer!")
	}

	currentNominator := state.NominationOrder[state.CurrentNominator]
	if currentNominator == nil {
		return fmt.Errorf("current nomintor is nil")
	}

	auctionStatesMu.Unlock()
	log.Println("Mutex Unlocked.")

	log.Println("Creating embed...")
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("The nomination phase has begun! It is %s's turn to nominate a Pokemon.", currentNominator.Username),
		Description: `Use "/nominate" to pick a Pokemon to nominate.`,
	}

	log.Println("Setting embed image...")
	if user, err := s.User(state.NominationOrder[state.CurrentNominator].UserID); err == nil && user != nil {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: user.AvatarURL("256"),
		}
	} else {
		log.Printf("Warning: Could not fetch nominator avatar: %v", err)
	}

	edit := &discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{},
	}

	auctionStatesMu.Lock()
	log.Println("Setting Nomination Phase to true...")
	state.NominationPhase = true
	auctionStatesMu.Unlock()

	_, err := s.ChannelMessageEditComplex(edit)

	log.Println("Done setting up Nomination Phase!")
	return err
}

func NominateCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	auctionStatesMu.Lock()

	log.Printf("NominateCallback called in channel: %s\n", i.ChannelID)
	for msgID, state := range auctionStates {
		log.Printf("State: msgID=%s, ChannelID=%s, NominationPhase=%v", msgID, state.ChannelID, state.NominationPhase)
	}

	var activeState *types.AuctionState
	var messageID string
	for id, state := range auctionStates {
		if state.NominationPhase && state.ChannelID == i.ChannelID {
			activeState = state
			messageID = id
			break
		}
	}

	auctionStatesMu.Unlock()

	if activeState == nil || !activeState.NominationPhase {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No active nomination phase found!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("error responding to command: %s\n", err)
		}
		return
	}

	if i.Member.User.ID != activeState.NominationOrder[activeState.CurrentNominator].UserID {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "It's not your turn to nominate!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("error responding to command: %s\n", err)
		}
		return
	}

	pokemonName, ok := i.ApplicationCommandData().Options[0].Value.(string)
	if !ok {
		log.Println("pokemon name not type of string")
		return
	}

	if slices.Contains(activeState.PreviouslyNominated, pokemonName) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Cannot nominate a Pokemon that has already been nominated! Please nominate a different Pokemon!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("error responding to command: %s\n", err)
		}
		return
	}

	// Immediately acknowledge interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Nominating %s...", pokemonName),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("error responding to interaction: %s\n", err)
		return
	}

	pokemon, err := fetch.FetchPokemon(activeState.GenNumber, pokemonName)
	if err != nil {
		log.Printf("error nominating pokemon: %s\n", err)
		return
	}

	auctionStatesMu.Lock()
	activeState.NominatedPokemon = pokemon
	auctionStatesMu.Unlock()

	msg, err := s.ChannelMessage(i.ChannelID, messageID)
	if err != nil {
		log.Printf("could not fetch message: %s\n", err)
		return
	}

	image := &discordgo.MessageEmbedImage{
		URL: pokemon.Sprite,
	}

	activeState.BidSoFar[i.Member.User.ID] = 50
	activeState.HighestBid = 50
	edit := &discordgo.MessageEdit{
		Channel: msg.ChannelID,
		ID:      msg.ID,
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("%s was nominated!", pokemon.Name),
				Image: image,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Highest Bid",
						Value: fmt.Sprintf("%s: $%d", i.Member.User.Username, 50),
					},
				},
			},
		},
	}

	updatedMsg, err := s.ChannelMessageEditComplex(edit)
	if err != nil {
		log.Printf("error editing message: %s\n", err)
		return
	}

	auctionStatesMu.Lock()
	activeState.PreviouslyNominated = append(activeState.PreviouslyNominated, pokemonName)
	activeState.NominationPhase = false
	activeState.BiddingPhase = true
	auctionStatesMu.Unlock()

	var player *types.Player
	for _, p := range participants {
		if p.UserID == i.Member.User.ID {
			player = p
		}
	}

	updateBidTimer(s, i, activeState, updatedMsg, player)
}
