package cmd

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/database"
	"github.com/mehkij/poke-auction/internal/dispatcher"
	"github.com/mehkij/poke-auction/internal/export"
	"github.com/mehkij/poke-auction/internal/fetch"
	"github.com/mehkij/poke-auction/internal/types"
	"github.com/mehkij/poke-auction/internal/utils"
)

func BidTimer(s *discordgo.Session, i *discordgo.InteractionCreate, msg *discordgo.Message, player *types.Player, pokemon *types.Pokemon, stopSignal chan bool, activeState *types.AuctionState, gd *dispatcher.Dispatcher, timerString string) {
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

	timerInt, err := strconv.Atoi(timerString)
	if err != nil {
		log.Printf("error converting timer string to int: %s", err)
	}

	utils.Timer(timerInt, stopSignal, func(duration int) {
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

		gd.QueueEditMessage(s, msg.ChannelID, msg.ID, edit)
	}, func() {
		var highestBidderID string
		var highestBid int

		activeState.AuctionStateMu.Lock()

		if activeState.ProcessingBid {
			activeState.AuctionStateMu.Unlock()

			// Wait up to 2 seconds for bid to process
			for i := 0; i < 20; i++ {
				time.Sleep(100 * time.Millisecond)

				activeState.AuctionStateMu.Lock()
				if !activeState.ProcessingBid {
					activeState.AuctionStateMu.Unlock()
					break
				}
				activeState.AuctionStateMu.Unlock()
			}

			activeState.AuctionStateMu.Lock()
		}

		for k, v := range activeState.BidSoFar {
			if v > highestBid {
				highestBidderID = k
				highestBid = v
			}
		}

		log.Println("PokeDollars after bid: ")
		for _, p := range activeState.Participants {
			log.Printf("%s: %d", p.Username, p.PokeDollars)
		}

		if len(activeState.NominationOrder) > 0 {
			for i, p := range activeState.NominationOrder {
				if p.UserID == highestBidderID {
					p.PokeDollars -= activeState.HighestBid
					p.Team = append(p.Team, pokemon)
					log.Printf("Pokemon %s added to player %s's team", pokemon.Name, p.UserID)

					gd.QueueSendMessage(s, activeState.ChannelID, &discordgo.MessageSend{
						Content: fmt.Sprintf("Congratulations, %s! You obtained: %s", p.Username, pokemon.Name),
					})

					if p.PokeDollars == 0 {
						team := fetch.RollRandomBabyPokemon(p.Team, activeState.GenNumber)
						p.Team = team
					}

					// Remove the player once their team is full
					if len(p.Team) == 6 {
						activeState.NominationOrder = slices.Delete(activeState.NominationOrder, i, i+1)
					}
					break
				}
			}
		}

		// Clear bid state for next round
		activeState.BidSoFar = make(map[string]int)
		activeState.HighestBid = 0

		// Notify participants of their remaining balances and the current teams
		log.Println("Notifying users of their remaining balance and the current teams...")
		var remaining []string
		var currentTeams []string
		for _, p := range activeState.Participants {
			remaining = append(remaining, fmt.Sprintf("%s's Balance: %d", p.Username, p.PokeDollars))

			var team []string
			for _, pokemon := range p.Team {
				team = append(team, pokemon.Name)
			}
			currentTeams = append(currentTeams, fmt.Sprintf("%s's Team: %s", p.Username, strings.Join(team, ", ")))
		}

		activeState.AuctionStateMu.Unlock()

		if activeState.BalanceMessageID == "" {
			m := &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Player Balances",
						Description: strings.Join(remaining, "\n"),
					},
					{
						Title:       "Current Teams",
						Description: strings.Join(currentTeams, "\n"),
					},
				},
			}
			done := gd.QueueSendMessage(s, activeState.ChannelID, m)
			sentMsg := <-done
			if sentMsg != nil {
				activeState.BalanceMessageID = sentMsg.ID
			}
		} else {
			edit := &discordgo.MessageEdit{
				Channel: msg.ChannelID,
				ID:      activeState.BalanceMessageID,
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Player Balances",
						Description: strings.Join(remaining, "\n"),
					},
					{
						Title:       "Current Teams",
						Description: strings.Join(currentTeams, "\n"),
					},
				},
			}
			done := gd.QueueEditMessage(s, activeState.ChannelID, activeState.BalanceMessageID, edit)
			<-done
		}

		activeState.AuctionStateMu.Lock()

		activeState.NominatedPokemon = nil
		log.Println("Set nominated Pokemon to nil!")

		if len(activeState.NominationOrder) == 0 {
			delete(auctionStates, msg.ID)
			log.Printf("Deleted auction state for message ID: %s", msg.ID)
			activeState.AuctionStateMu.Unlock()

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

			export.ExportTeam(s, i, activeState.Participants, activeState.GenNumber)
			return
		}

		log.Println("Unlocking Mutex...")
		activeState.AuctionStateMu.Unlock()
		log.Println("Mutex unlocked!")

		log.Println("Starting Nomination Phase...")
		newInteraction := &discordgo.InteractionCreate{
			Interaction: i.Interaction,
		}
		newInteraction.Message = msg

		err := NominationPhase(s, newInteraction, activeState, gd)
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
func BidCallback(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *types.GlobalConfig) {
	gd := cfg.GlobalDispatcher

	gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Bidding $%s...", i.ApplicationCommandData().Options[0].StringValue()),
		},
	})
	// <-done

	if len(i.ApplicationCommandData().Options) == 0 {
		gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid bid command usage",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	mu.Lock()
	var activeState *types.AuctionState
	var msgID string
	for id, state := range auctionStates {
		if state.BiddingPhase && state.ChannelID == i.ChannelID {
			activeState = state
			msgID = id
			break
		}
	}
	mu.Unlock()

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

	if bidAmount <= activeState.HighestBid {
		utils.CreateFollowupEphemeralError(s, i, "Your bid must be higher than the highest bid!")
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

	activeState.AuctionStateMu.Lock()
	bidder, ok := activeState.Participants[i.Member.User.ID]
	if !ok {
		log.Println("user not an active participant in this auction")
		activeState.AuctionStateMu.Unlock()
		return
	}

	if bidAmount == 0 {
		utils.CreateFollowupEphemeralError(s, i, "Bid amount must be greater than 0!")
		activeState.AuctionStateMu.Unlock()
		return
	}

	if bidder.PokeDollars >= bidAmount {
		activeState.BidSoFar[i.Member.User.ID] = bidAmount
	} else {
		utils.CreateFollowupEphemeralError(s, i, "Not enough PokeDollars!")
		activeState.AuctionStateMu.Unlock()
		return
	}
	activeState.AuctionStateMu.Unlock()

	if bidder == nil {
		utils.CreateFollowupEphemeralError(s, i, "You are not a participant in this auction!")
		return
	}

	if activeState.NominatedPokemon == nil {
		utils.CreateFollowupEphemeralError(s, i, "No Pokemon has been nominated for bidding!")
		return
	}

	activeState.AuctionStateMu.Lock()
	activeState.ProcessingBid = true
	activeState.AuctionStateMu.Unlock()

	defer func() {
		activeState.AuctionStateMu.Lock()
		activeState.ProcessingBid = false
		activeState.AuctionStateMu.Unlock()
	}()

	var biddersField string
	var highestBid string

	activeState.AuctionStateMu.Lock()
	var highestBidValue int
	var highestBidderID string
	for k, v := range activeState.BidSoFar {
		if v > highestBidValue {
			highestBidValue = v
			highestBidderID = k
		}
	}

	if highestBidValue > activeState.HighestBid {
		activeState.HighestBid = highestBidValue
	}

	if highestBidderID != "" {
		user, _ := s.User(highestBidderID)
		highestBid = fmt.Sprintf("%s: $%d", user.Username, highestBidValue)
	}
	activeState.AuctionStateMu.Unlock()

	biddersField = highestBid
	if biddersField == "" {
		biddersField = "No bids yet..."
	}

	var fields []*discordgo.MessageEmbedField
	var bidderFieldFound bool
	for _, field := range msg.Embeds[0].Fields {
		if field.Name == "Highest Bid" {
			field.Value = biddersField
			bidderFieldFound = true
		}
		fields = append(fields, field)
	}

	if !bidderFieldFound {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Highest Bid",
			Value: biddersField,
		})
	}

	var timerVal string
	val, err := cfg.Queries.GetConfigOption(context.Background(), database.GetConfigOptionParams{
		ServerID: i.GuildID,
		Key:      "BidTimerDuration",
	})
	if err != nil {
		log.Printf("error getting config option from DB: %s", err)
		timerVal = "30"
	} else {
		timerVal = val
	}

	done := gd.QueueEditMessage(s, i.ChannelID, msg.ID, &discordgo.MessageEdit{
		Channel: i.ChannelID,
		ID:      msg.ID,
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       msg.Embeds[0].Title,
				Description: msg.Embeds[0].Description,
				Image:       msg.Embeds[0].Image,
				Fields:      fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Timer: %s", timerVal),
				},
			},
		},
	})
	updatedMsg := <-done

	activeState.AuctionStateMu.Lock()
	// If a timer is currently running, stop it.
	if activeState.StopSignal != nil {
		safeCloseChannel(activeState.StopSignal)
	}

	stopSignal := make(chan bool)
	activeState.StopSignal = stopSignal
	activeState.AuctionStateMu.Unlock()

	log.Println("About to call updateBidTimer...")
	updateBidTimer(s, i, activeState, updatedMsg, bidder, gd, timerVal)
	log.Println("Called updateBidTimer")
}

func updateBidTimer(s *discordgo.Session, i *discordgo.InteractionCreate, state *types.AuctionState, msg *discordgo.Message, bidder *types.Player, gd *dispatcher.Dispatcher, timerString string) {
	log.Println("Entered updateBidTimer function")

	// Validate parameters
	if s == nil || i == nil || state == nil || msg == nil || bidder == nil {
		log.Printf("updateBidTimer parameters check: session=%v, interaction=%v, state=%v, msg=%v, bidder=%v",
			s != nil, i != nil, state != nil, msg != nil, bidder != nil)
		return
	}

	state.AuctionStateMu.Lock()

	if state.NominatedPokemon == nil {
		log.Println("error updating bid timer: NominatedPokemon is nil")
		state.AuctionStateMu.Unlock()
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
	state.AuctionStateMu.Unlock()
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
		BidTimer(s, i, msg, bidder, state.NominatedPokemon, newStopSignal, state, gd, timerString)
	}()
	log.Println("Goroutine started")
}
