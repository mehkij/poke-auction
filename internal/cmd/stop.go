package cmd

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/dispatcher"
	"github.com/mehkij/poke-auction/internal/utils"
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

			state.AuctionStateMu.Lock()

			log.Printf("Attempting to stop auction for message %s in channel %s", msgID, state.ChannelID)

			// Try to fetch the message before editing
			_, err := s.ChannelMessage(state.ChannelID, msgID)
			if err != nil {
				log.Printf("Could not fetch message %s: %v", msgID, err)
			} else {
				log.Printf("Fetched message %s successfully", msgID)
			}

			log.Println("Interrupting timers...")
			// Interrupt the auction's timer if it is running
			utils.SafeCloseChannel(state.StopSignal)
			log.Println("Timers interrupted.")

			done := gd.QueueEditMessage(s, state.ChannelID, msgID, &discordgo.MessageEdit{
				Channel: state.ChannelID,
				ID:      msgID,
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Auction closed.",
						Description: "A /stopall command was issued, this auction is now closed.",
					},
				},
				Components: &[]discordgo.MessageComponent{},
			})

			editedMsg := <-done
			if editedMsg == nil {
				log.Printf("Failed to edit message %s in channel %s", msgID, state.ChannelID)
			} else {
				log.Printf("Successfully edited message %s in channel %s", msgID, state.ChannelID)
			}

			log.Println("Deleting auction state...")
			// Delete the auction's state from the channel
			delete(auctionStates, msgID)
			log.Println("Auction state deleted.")

			state.AuctionStateMu.Unlock()
		}
	}
	mu.Unlock()
	log.Printf("All auctions in channel %s closed.", channelID)
}
