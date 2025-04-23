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
	participants []*types.Player
	mu           sync.Mutex
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

func AuctionTimer(s *discordgo.Session, i *discordgo.InteractionCreate, timerStr string) {
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
							Value:  strings.Join(usernames, "\n"),
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

	for range ticker.C {
		if time.Now().After(endTime) {
			updateMessage()
			return
		}
		if err := updateMessage(); err != nil {
			log.Printf("error updating message: %s", err)
			return
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
	return
}

func AuctionCallback(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Reset participants list at start of new auction
	participants = make([]*types.Player, 0)

	// Get timer value directly from command options
	timerStr := i.ApplicationCommandData().Options[1].StringValue()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("The Gen %s auction is beginning!", i.ApplicationCommandData().Options[0].StringValue()),
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

	// Start the timer with the original timer value
	go AuctionTimer(s, newInteraction, timerStr)
}
