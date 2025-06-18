package cmd

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/dispatcher"
)

func StopAllCallback(s *discordgo.Session, i *discordgo.InteractionCreate, gd *dispatcher.Dispatcher) {
	log.Println("Stopping all auctions...")
	gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Stopping all auctions...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	mu.Lock()

	var channelID string

	for msgID, state := range auctionStates {
		if state.ChannelID == i.ChannelID {
			channelID = state.ChannelID

			log.Println("Interrupting timers...")
			// Interrupt the auction's timer if it is running
			state.StopSignal <- true
			close(state.StopSignal)
			log.Println("Timers interrupted.")

			gd.QueueEditMessage(s, state.ChannelID, msgID, &discordgo.MessageEdit{
				Channel: state.ChannelID,
				ID:      msgID,
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Auction closed.",
						Description: "A /stopall command was issued, therefore we will be closing all auctions.",
					},
				},
				Components: &[]discordgo.MessageComponent{},
			})

			log.Println("Deleting auction state...")
			// Delete the auction's state from the channel
			delete(auctionStates, msgID)
			log.Println("Auction state deleted.")
		}
	}
	mu.Unlock()
	log.Printf("All auctions in channel %s closed.", channelID)
}
