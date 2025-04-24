package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
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

	auctionStatesMu.Lock()
	state, exists := auctionStates[i.Message.ID]
	auctionStatesMu.Unlock()

	if !exists || len(state.NominationOrder) == 0 {
		return fmt.Errorf("no nomination order found")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("The nomination phase has begun! It is %s's turn to nominate a Pokemon.", state.NominationOrder[0].Username),
		Description: `Use "/nominate" to pick a Pokemon to nominate.`,
	}

	if user, err := s.User(state.NominationOrder[0].UserID); err == nil && user != nil {
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
	state.NominationPhase = true
	auctionStatesMu.Unlock()

	_, err := s.ChannelMessageEditComplex(edit)

	return err
}

func NominateCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	auctionStatesMu.Lock()
	log.Printf("NominateCallback called in channel: %s\n", i.ChannelID)
	for msgID, state := range auctionStates {
		log.Printf("State: msgID=%s, ChannelID=%s, NominationPhase=%v", msgID, state.ChannelID, state.NominationPhase)
	}

	var activeState *types.AuctionState
	for _, state := range auctionStates {
		if state.NominationPhase && state.ChannelID == i.ChannelID {
			activeState = state
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

	pokemon, err := fetch.FetchPokemon(activeState.GenNumber, pokemonName)
	if err != nil {
		log.Printf("error nominating pokemon: %s\n", err)
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:  "Ability",
			Value: strings.Join(pokemon.Abilities, "\n"),
		},
	}
	image := &discordgo.MessageEmbedImage{
		URL: pokemon.Sprite,
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  fmt.Sprintf("%s was nominated!", pokemon.Name),
					Fields: fields,
					Image:  image,
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Join Auction",
							Style:    discordgo.PrimaryButton,
							CustomID: "join_auction",
						},
						discordgo.Button{
							Label:    "Force Start",
							Style:    discordgo.DangerButton,
							CustomID: "force_start",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("error responding to command: %s", err)
		return
	}
}
