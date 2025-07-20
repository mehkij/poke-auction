package cmd

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/export"
	"github.com/mehkij/poke-auction/internal/fetch"
	"github.com/mehkij/poke-auction/internal/types"
)

func PickCallback(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *types.GlobalConfig) {
	gd := cfg.GlobalDispatcher

	mu.Lock()
	var activeState *types.AuctionState
	var msgID string
	for id, state := range auctionStates {
		if len(state.Participants) > 0 && state.ChannelID == i.ChannelID {
			activeState = state
			msgID = id
			break
		}
	}
	mu.Unlock()

	if activeState == nil {
		gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "No active auction found in this channel.",
			},
		})
		return
	}

	activeState.AuctionStateMu.Lock()
	if len(activeState.NominationOrder) != 1 {
		gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "You are not the only participant in the auction!",
			},
		})
		activeState.AuctionStateMu.Unlock()
		return
	}
	activeState.AuctionStateMu.Unlock()

	pickedPokemon := i.ApplicationCommandData().Options[0].StringValue()

	if slices.Contains(activeState.PreviouslyNominated, pickedPokemon) {
		done := gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Cannot nominate a Pokemon that has already been picked by someone! Please choose a different Pokemon!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		<-done
		return
	}

	gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Picking... %s", i.ApplicationCommandData().Options[0].StringValue()),
		},
	})

	// Validate if chosen Pokemon actually exists
	pokemon, err := fetch.FetchPokemon(activeState.GenNumber, pickedPokemon)
	if err != nil {
		log.Printf("error picking pokemon: %s\n", err)
		return
	}

	activeState.AuctionStateMu.Lock()
	activeState.NominationOrder[0].Team = append(activeState.NominationOrder[0].Team, &types.Pokemon{
		Name: pokemon.Name,
	})
	activeState.AuctionStateMu.Unlock()

	msg, err := s.ChannelMessage(i.ChannelID, msgID)
	if err != nil {
		log.Printf("could not fetch message: %s\n", err)
		return
	}

	var team []string
	activeState.AuctionStateMu.Lock()
	for _, playersPokemon := range activeState.NominationOrder[0].Team {
		team = append(team, playersPokemon.Name)
	}
	activeState.AuctionStateMu.Unlock()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s, please choose your remaining Pokemon!", activeState.NominationOrder[0].Username),
		Description: strings.Join(team, "\n"),
	}

	if user, err := s.User(activeState.NominationOrder[0].UserID); err == nil && user != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		}
	} else {
		log.Printf("Warning: Could not fetch nominator avatar: %v", err)
	}

	gd.QueueEditMessage(s, i.ChannelID, msg.ID, &discordgo.MessageEdit{
		Channel: msg.ChannelID,
		ID:      msg.ID,
		Embed:   embed,
	})

	activeState.AuctionStateMu.Lock()
	activeState.PreviouslyNominated = append(activeState.PreviouslyNominated, pickedPokemon)
	activeState.AuctionStateMu.Unlock()

	if len(team) == 6 {
		gd.QueueEditMessage(s, msg.ChannelID, msg.ID, &discordgo.MessageEdit{
			Channel: msg.ChannelID,
			ID:      msg.ID,
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       "Auction Complete!",
					Description: "All players have completed their teams.",
				},
			},
		})

		export.ExportTeam(s, i, activeState.Participants, activeState.GenNumber, gd)
		delete(auctionStates, msg.ID)
	}
}
