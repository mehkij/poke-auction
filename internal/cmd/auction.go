package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/types"
	"github.com/mehkij/poke-auction/internal/utils"
)

// "/auction" command definition.
var AuctionCommand = &Command{
	Name:        "auction",
	Description: "Start a Pokemon auction.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "generation",
			Description: "The generation of the pool of Pokemon to choose from.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "timer",
			Description: "Set the time before the auction begins in seconds.",
			Required:    true,
		},
	},
	Callback: AuctionCallback,
}

var (
	mu            sync.Mutex
	auctionStates = make(map[string]*types.AuctionState)
)

func JoinAuction(i *discordgo.InteractionCreate) []*types.Player {
	user := i.Member.User
	username := user.Username
	id := user.ID

	mu.Lock()
	state, exists := auctionStates[i.Message.ID]
	mu.Unlock()
	if !exists {
		log.Printf("auction state not found!")
		return nil
	}

	state.AuctionStateMu.Lock()
	defer state.AuctionStateMu.Unlock()
	// Check if user is already in participants
	for _, p := range state.Participants {
		if p.UserID == id {
			return state.Participants
		}
	}

	// Add new participant
	state.Participants = append(state.Participants, &types.Player{
		Username:    username,
		UserID:      id,
		PokeDollars: 10000,
	})

	return state.Participants
}

func AuctionTimer(s *discordgo.Session, i *discordgo.InteractionCreate, timerStr string, stopSignal chan bool) {
	timeLeft, err := strconv.Atoi(timerStr)
	if err != nil {
		log.Printf("error converting timer to int: %s", err)
		return
	}

	utils.Timer(timeLeft, stopSignal, func(d int) {
		var usernames []string

		mu.Lock()
		state, exists := auctionStates[i.Message.ID]
		mu.Unlock()

		if exists {
			state.AuctionStateMu.Lock()
			for _, p := range state.Participants {
				usernames = append(usernames, p.Username)
			}
			state.AuctionStateMu.Unlock()
		}

		participantsValue := "No participants yet..."
		if len(usernames) > 0 {
			participantsValue = strings.Join(usernames, "\n")
		}

		edit := &discordgo.MessageEdit{
			Channel: i.ChannelID,
			ID:      i.Message.ID,
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title:       i.Message.Embeds[0].Title,
					Description: i.Message.Embeds[0].Description,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Participants",
							Value:  participantsValue,
							Inline: false,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("Timer: %d", d),
					},
				},
			},
		}

		s.ChannelMessageEditComplex(edit)
	}, func() {
		mu.Lock()
		state, exists := auctionStates[i.Message.ID]
		mu.Unlock()

		if exists {
			state.AuctionStateMu.Lock()
			state.CurrentNominator = -1 // Starts at -1 so that when NominationPhase is called, it is incremented to 0
			state.NominationOrder = RollNominationOrder(state)
			state.NominationPhase = true
			state.BidSoFar = make(map[string]int)
			state.AuctionStateMu.Unlock()
		}

		err := NominationPhase(s, i, state)
		if err != nil {
			fmt.Printf("error starting nomination phase on timer end: %s", err)
			return
		}
	}, func() {})
}

// Called when "Join Auction" button is clicked.
func HandleAuctionInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("error responding to interaction: %s\n", err)
		return
	}

	JoinAuction(i)
}

// Called when "Force Start" button is clicked.
func HandleForceStartAuction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("error responding to interaction: %s\n", err)
		return
	}

	mu.Lock()
	state, exists := auctionStates[i.Message.ID]
	mu.Unlock()
	if exists {
		state.AuctionStateMu.Lock()
		state.StopSignal <- true
		state.CurrentNominator = -1 // Starts at -1 so that when NominationPhase is called, it is incremented to 0
		state.NominationOrder = RollNominationOrder(state)
		state.NominationPhase = true
		state.BidSoFar = make(map[string]int)
		state.AuctionStateMu.Unlock()
	}

	err = NominationPhase(s, i, state)
	if err != nil {
		log.Printf("error starting nomination phase: %s\n", err)
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Error starting nomination phase: %s", err))
	}
}

func AuctionCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Reset participants list at start of new auction
	participants := make([]*types.Player, 0)

	// Get timer value directly from command options
	timerStr := i.ApplicationCommandData().Options[1].StringValue()

	gen, err := strconv.Atoi(i.ApplicationCommandData().Options[0].Value.(string))
	if err != nil {
		log.Println("error while converting string to int")
		return
	}

	if gen < 1 || gen > 9 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please choose a valid generation! (1-9)",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("error responding to command: %s\n", err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("The Gen %d auction is beginning!", gen),
					Description: "Please register by clicking the button.",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Participants",
							Value:  "",
							Inline: false,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("Timer: %s", timerStr),
					},
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

	// Get the message we just created
	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Printf("error getting interaction response: %s", err)
		return
	}

	newInteraction := &discordgo.InteractionCreate{
		Interaction: i.Interaction,
	}
	newInteraction.Message = msg

	stopSignal := make(chan bool)
	mu.Lock()
	if oldState, exists := auctionStates[msg.ID]; exists {
		close(oldState.StopSignal)
		delete(auctionStates, msg.ID)
	}
	auctionStates[msg.ID] = &types.AuctionState{Participants: participants, StopSignal: stopSignal, GenNumber: gen, ChannelID: msg.ChannelID}
	log.Printf("Auction state set: msgID=%s, ChannelID=%s", msg.ID, msg.ChannelID)
	mu.Unlock()

	// Start the timer with the original timer value with a cleanup channel
	go AuctionTimer(s, newInteraction, timerStr, stopSignal)
}

func safeCloseChannel(ch chan bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from channel close panic: %v", r)
		}
	}()

	select {
	case <-ch:
		return
	default:
		close(ch)
	}
}

// Signal cleanup helper function
func CleanupAuctionTimer(messageID string) {
	mu.Lock()
	defer mu.Unlock()

	if state, exists := auctionStates[messageID]; exists && state.StopSignal != nil {
		state.AuctionStateMu.Lock()
		safeCloseChannel(state.StopSignal)
		state.StopSignal = nil
		state.AuctionStateMu.Unlock()
	}
}
