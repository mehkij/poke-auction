package cmd

import (
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/export"
	"github.com/mehkij/poke-auction/internal/types"
	"github.com/mehkij/poke-auction/internal/utils"
)

var BidCommand = &Command{
	Name:        "bid",
	Description: "Bid on a Pokemon.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "amount",
			Description: "How many PokeDollars to bid.",
			Required:    true,
		},
	},
	Callback: BidCallback,
}

func BidTimer(s *discordgo.Session, i *discordgo.InteractionCreate, msg *discordgo.Message, player *types.Player, pokemon *types.Pokemon, stopSignal chan bool) {
	// Add detailed parameter checking
	log.Printf("BidTimer parameters check: session=%v, interaction=%v, msg=%v, player=%v, pokemon=%v, stopSignal=%v",
		s != nil, i != nil, msg != nil, player != nil, pokemon != nil, stopSignal != nil)

	if s == nil || i == nil || msg == nil || player == nil || pokemon == nil || stopSignal == nil {
		log.Printf("BidTimer received nil parameters!")
		return
	}

	// Check msg.Embeds
	if len(msg.Embeds) == 0 {
		log.Printf("BidTimer: message has no embeds!")
		return
	}

	fmt.Println("Timer starting...")
	utils.Timer(1, stopSignal, func(duration int) {
		// Defensive check for embed
		if len(msg.Embeds) == 0 {
			log.Printf("Timer update: message has no embeds!")
			return
		}

		currentEmbed := msg.Embeds[0]
		edit := &discordgo.MessageEdit{
			Channel: msg.ChannelID,
			ID:      msg.ID,
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       currentEmbed.Title,
					Description: currentEmbed.Description,
					Fields:      currentEmbed.Fields,
					Image:       currentEmbed.Image,
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("Timer: %d", duration),
					},
				},
			},
		}

		_, err := s.ChannelMessageEditComplex(edit)
		if err != nil {
			log.Printf("error updating timer: %v", err)
		}
	}, func() {
		var highestBidderID string
		var highestBid int

		mu.Lock()
		for k, v := range bidSoFar {
			if v > highestBid {
				highestBidderID = k
				highestBid = v
			}
		}
		// Clear bids for next round
		bidSoFar = make(map[string]int)
		mu.Unlock()

		auctionStatesMu.Lock()
		state, exists := auctionStates[msg.ID]
		if !exists {
			log.Printf("auction state not found for message ID: %s", msg.ID)
			auctionStatesMu.Unlock()
			return
		}
		auctionStatesMu.Unlock()

		auctionStatesMu.Lock()

		if len(state.NominationOrder) > 0 {
			for i, p := range state.NominationOrder {
				if p.UserID == highestBidderID {
					p.Team = append(p.Team, pokemon)
					log.Printf("Pokemon %s added to player %s's team", pokemon.Name, p.UserID)

					// Remove the player once their team is full
					if len(p.Team) == 6 {
						state.NominationOrder = slices.Delete(state.NominationOrder, i, i+1)
					}
					break
				}
			}
		}

		state.NominatedPokemon = nil
		log.Println("Set nominated Pokemon to nil!")

		if len(state.NominationOrder) == 0 {
			delete(auctionStates, msg.ID)
			log.Printf("Deleted auction state for message ID: %s", msg.ID)
			auctionStatesMu.Unlock()

			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: msg.ChannelID,
				ID:      msg.ID,
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Auction Complete!",
						Description: "All players have completed their teams.",
					},
				},
			})
			if err != nil {
				log.Printf("error updating final message: %v", err)
			}

			export.ExportTeam(s, i, participants, state.GenNumber)
			return
		}

		log.Println("Unlocking Mutex...")
		auctionStatesMu.Unlock()
		log.Println("Mutex unlocked!")

		log.Println("Starting Nomination Phase...")
		newInteraction := &discordgo.InteractionCreate{
			Interaction: i.Interaction,
		}
		newInteraction.Message = msg

		err := NominationPhase(s, newInteraction)
		if err != nil {
			log.Printf("error starting nomination phase: %v", err)
		}
		fmt.Println("Timer stopped.")
	}, func() {
		log.Println("Timer interrupted")
	})
	fmt.Println("Timer started!")
}

/*
Making a bid should be blocking so that the order of incoming bids are preserved.
This also avoids common bugs surrounding timers not properly resetting per bid.
*/
func BidCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Bidding $%s...", i.ApplicationCommandData().Options[0].StringValue()),
		},
	})
	if err != nil {
		log.Printf("error sending initial response: %v", err)
		return
	}

	if len(i.ApplicationCommandData().Options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid bid command usage",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	auctionStatesMu.Lock()
	var activeState *types.AuctionState
	var msgID string
	for id, state := range auctionStates {
		if state.NominationPhase && state.ChannelID == i.ChannelID {
			activeState = state
			msgID = id
			break
		}
	}
	auctionStatesMu.Unlock()

	if activeState == nil {
		utils.CreateFollowupEphemeralError(s, i, "No active auction in bidding phase!")
		return
	}

	if msgID == "" {
		log.Println("no pokemon has been nominated!")
		return
	}

	// Validate that user making a bid does not have a full team
	var found bool
	for _, player := range activeState.NominationOrder {
		if player.UserID == i.Member.User.ID {
			found = true
			break
		}
	}
	if !found {
		utils.CreateFollowupEphemeralError(s, i, "You cannot bid anymore, your team is full!")
		return
	}

	bidAmount, err := strconv.Atoi(i.ApplicationCommandData().Options[0].StringValue())
	if err != nil {
		utils.CreateFollowupEphemeralError(s, i, "Invalid bid amount!")
		return
	}

	msg, err := s.ChannelMessage(i.ChannelID, msgID)
	if err != nil {
		log.Printf("could not fetch message: %s\n", err)
		return
	}

	if msg == nil {
		log.Printf("message is nil for msgID: %s, channelID: %s\n", msgID, i.ChannelID)
		utils.CreateFollowupEphemeralError(s, i, "Internal error: auction message not found.")
		return
	}

	if len(msg.Embeds) == 0 {
		log.Printf("no embeds found in message of ID: %s\n", msg.ID)
		utils.CreateFollowupEphemeralError(s, i, "Internal error: no embed found for auction message.")
		return
	}

	var bidder *types.Player

	for _, p := range participants {
		if i.Member.User.ID == p.UserID {
			// Ensure bid is not 0
			if bidAmount == 0 {
				utils.CreateFollowupEphemeralError(s, i, "Bid amount must be greater than 0!")
				return
			}

			// Ensure user has enough PokeDollars to make a bid
			if p.PokeDollars >= bidAmount {
				p.PokeDollars -= bidAmount
				bidSoFar[i.Member.User.ID] += bidAmount
				bidder = p
			} else {
				utils.CreateFollowupEphemeralError(s, i, "Not enough PokeDollars!")
				return
			}
			break
		}
	}

	if bidder == nil {
		utils.CreateFollowupEphemeralError(s, i, "You are not a participant in this auction!")
		return
	}

	if activeState.NominatedPokemon == nil {
		utils.CreateFollowupEphemeralError(s, i, "No Pokemon has been nominated for bidding!")
		return
	}

	var biddersVal string
	var bids []string
	for k, v := range bidSoFar {
		user, _ := s.User(k)
		bids = append(bids, fmt.Sprintf("%s: $%d", user.Username, v))
	}

	biddersVal = strings.Join(bids, "\n")
	if biddersVal == "" {
		biddersVal = "No bids yet..."
	}

	var fields []*discordgo.MessageEmbedField
	var bidderFieldFound bool
	for _, field := range msg.Embeds[0].Fields {
		if field.Name == "Bidders" {
			field.Value = biddersVal
			bidderFieldFound = true
		}
		fields = append(fields, field)
	}

	if !bidderFieldFound {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Bidders",
			Value: biddersVal,
		})
	}

	edit := &discordgo.MessageEdit{
		Channel: i.ChannelID,
		ID:      msg.ID,
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       msg.Embeds[0].Title,
				Description: msg.Embeds[0].Description,
				Image:       msg.Embeds[0].Image,
				Fields:      fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Timer: 1",
				},
			},
		},
	}

	updatedMsg, err := s.ChannelMessageEditComplex(edit)
	if err != nil {
		log.Printf("error editing embed: %s\n", err)
		return
	}

	auctionStatesMu.Lock()
	// If a timer is currently running, stop it.
	if activeState.StopSignal != nil {
		safeCloseChannel(activeState.StopSignal)
	}

	stopSignal := make(chan bool)
	activeState.StopSignal = stopSignal
	auctionStatesMu.Unlock()

	log.Println("About to call updateBidTimer...")
	updateBidTimer(s, i, activeState, updatedMsg, bidder)
	log.Println("Called updateBidTimer")
}

func updateBidTimer(s *discordgo.Session, i *discordgo.InteractionCreate, state *types.AuctionState, msg *discordgo.Message, bidder *types.Player) {
	log.Println("Entered updateBidTimer function")

	// Validate parameters
	if s == nil || i == nil || state == nil || msg == nil || bidder == nil {
		log.Printf("updateBidTimer parameters check: session=%v, interaction=%v, state=%v, msg=%v, bidder=%v",
			s != nil, i != nil, state != nil, msg != nil, bidder != nil)
		return
	}

	auctionStatesMu.Lock()

	if state.NominatedPokemon == nil {
		log.Println("error updating bid timer: NominatedPokemon is nil")
		auctionStatesMu.Unlock()
		return
	}

	newStopSignal := make(chan bool)
	oldStopSignal := state.StopSignal
	state.StopSignal = newStopSignal

	// Stop existing timer if running
	if oldStopSignal != nil {
		log.Println("Stopping existing timer...")
		select {
		case oldStopSignal <- true:
			log.Println("Successfully sent stop signal")
		default:
			log.Println("Could not send stop signal - channel might be closed")
		}
		close(oldStopSignal)
	}

	log.Println("Unlocking mutex")
	auctionStatesMu.Unlock()
	log.Println("Mutex unlocked")

	log.Println("About to start goroutine for BidTimer")

	// Start new timer
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in BidTimer: %v", r)
				debug.PrintStack()
			}
		}()
		BidTimer(s, i, msg, bidder, state.NominatedPokemon, newStopSignal)
	}()
	log.Println("Goroutine started")
}
