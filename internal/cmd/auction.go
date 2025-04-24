package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/types"
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
	participants    []*types.Player
	mu              sync.Mutex
	auctionStates   = make(map[string]*types.AuctionState)
	auctionStatesMu sync.Mutex
)

func JoinAuction(i *discordgo.InteractionCreate) []*types.Player {
	user := i.Member.User
	username := user.Username
	id := user.ID

	// Check if user is already in participants
	for _, p := range participants {
		if p.UserID == id {
			return participants
		}
	}

	// Add new participant
	participants = append(participants, &types.Player{
		Username:    username,
		UserID:      id,
		PokeDollars: 10000,
	})

	return participants
}

func AuctionTimer(s *discordgo.Session, i *discordgo.InteractionCreate, timerStr string, stopSignal chan bool) {
	defer cleanupAuction(i.Message.ID)

	timeLeft, err := strconv.Atoi(timerStr)
	if err != nil {
		log.Printf("error converting timer to int: %s", err)
		return
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(timeLeft) * time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	updateMessage := func() error {
		mu.Lock()
		defer mu.Unlock()

		remaining := int(time.Until(endTime).Seconds())
		if remaining < 0 {
			remaining = 0
		}

		var usernames []string
		for _, p := range participants {
			usernames = append(usernames, p.Username)
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
						Text: fmt.Sprintf("Timer: %d", remaining),
					},
				},
			},
		}

		_, err := s.ChannelMessageEditComplex(edit)
		return err
	}

	// Initial update
	if err := updateMessage(); err != nil {
		log.Printf("error updating message: %s", err)
		return
	}

	for {
		select {
		case <-stopSignal:
			NominationPhase(s, i)
			return
		case <-ticker.C:
			if err := updateMessage(); err != nil {
				log.Printf("error updating message: %s", err)
				continue
			}
			if time.Now().After(endTime) {
				NominationPhase(s, i)
				return
			}
		}
	}
}

func HandleAuctionInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("error responding to interaction: %s\n", err)
		return
	}

	mu.Lock()
	JoinAuction(i)
	mu.Unlock()
}

func HandleForceStartAuction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("error responding to interaction: %s\n", err)
		return
	}

	auctionStatesMu.Lock()
	state, exists := auctionStates[i.Message.ID]
	if exists {
		state.StopSignal <- true
		state.NominationOrder = RollNominationOrder()
		state.NominationPhase = true
	}
	auctionStatesMu.Unlock()

	err = NominationPhase(s, i)
	if err != nil {
		log.Printf("error starting nomination phase: %s\n", err)
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Error starting nomination phase: %s", err))
	}
}

func AuctionCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Reset participants list at start of new auction
	participants = make([]*types.Player, 0)

	// Get timer value directly from command options
	timerStr := i.ApplicationCommandData().Options[1].StringValue()

	gen, err := strconv.Atoi(i.ApplicationCommandData().Options[0].Value.(string))
	if err != nil {
		log.Println("error while converting string to int")
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
	auctionStatesMu.Lock()
	if oldState, exists := auctionStates[msg.ID]; exists {
		close(oldState.StopSignal)
		delete(auctionStates, msg.ID)
	}
	auctionStates[msg.ID] = &types.AuctionState{StopSignal: stopSignal, GenNumber: gen, ChannelID: msg.ChannelID}
	log.Printf("Auction state set: msgID=%s, ChannelID=%s", msg.ID, msg.ChannelID)
	auctionStatesMu.Unlock()

	// Start the timer with the original timer value with a cleanup channel
	go AuctionTimer(s, newInteraction, timerStr, stopSignal)
}

// Signal cleanup helper function
func cleanupAuction(messageID string) {
	auctionStatesMu.Lock()
	defer auctionStatesMu.Unlock()

	if state, exists := auctionStates[messageID]; exists {
		select {
		case <-state.StopSignal:

		default:
			close(state.StopSignal)
		}
		// delete(auctionStates, messageID)
	}
}
