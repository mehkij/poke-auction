package cmd

import (
	"fmt"
	"log"
	"math/rand"
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
	// Ensure that incrementing the pointer doesn't exceed length of array, and set the pointer to 0 if it does.
	state.CurrentNominator++
	if state.CurrentNominator == len(state.NominationOrder) {
		state.CurrentNominator = 0
		log.Println("Pointer is equal to NominationOrder length, resetting pointer!")
	}
	auctionStatesMu.Unlock()
	log.Println("Mutex Unlocked.")

	log.Println("Creating embed...")
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("The nomination phase has begun! It is %s's turn to nominate a Pokemon.", state.NominationOrder[state.CurrentNominator].Username),
		Description: `Use "/nominate" to pick a Pokemon to nominate.`,
	}

	log.Println("Setting embed image...")
	if user, err := s.User(state.NominationOrder[state.CurrentNominator].UserID); err == nil && user != nil {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: user.AvatarURL(""),
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

	pokemonName, ok := i.ApplicationCommandData().Options[0].Value.(string)
	if !ok {
		log.Println("pokemon name not type of string")
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

	edit := &discordgo.MessageEdit{
		Channel: msg.ChannelID,
		ID:      msg.ID,
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("%s was nominated!", pokemon.Name),
				Image: image,
			},
		},
	}

	_, err = s.ChannelMessageEditComplex(edit)
	if err != nil {
		log.Printf("error editing message: %s\n", err)
		return
	}

	auctionStatesMu.Lock()
	activeState.BiddingPhase = true
	auctionStatesMu.Unlock()
}
