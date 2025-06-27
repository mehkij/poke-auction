package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/database"
	"github.com/mehkij/poke-auction/internal/types"
)

func UpdateConfig(guildID string, cfg *types.GlobalConfig, key, val string) {
	err := cfg.Queries.UpsertConfig(context.Background(), database.UpsertConfigParams{
		ServerID: guildID,
		Key:      key,
		Value:    val,
	})
	if err != nil {
		log.Printf("error inserting record into configs table: %s\n", err)
		return
	}
}

func ConfigCallback(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *types.GlobalConfig) {
	gd := cfg.GlobalDispatcher

	gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Getting your server's config...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	// Check if bot is configured for server. If not, configure it with default values.
	// Servers in the database are guaranteed to have configuration records.

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Printf("guild not found: %s", err)
		return
	}
	server, err := cfg.Queries.GetServer(context.Background(), i.GuildID)

	if err != nil {
		log.Printf("server not found in DB, creating new record...")

		err = cfg.Queries.UpsertServer(context.Background(), database.UpsertServerParams{
			ID: guild.ID,
			Name: sql.NullString{
				String: guild.Name,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("error inserting server record: %s", err)
			return
		}

		UpdateConfig(guild.ID, cfg, "BidTimerDuration", "30")
		UpdateConfig(guild.ID, cfg, "StartingAmount", "10000")
	}

	if i.ApplicationCommandData().Options != nil {
		if len(i.ApplicationCommandData().Options) < 2 {
			log.Println("missing arguments to command.")
			return
		}

		var key string
		var val string

		for _, option := range i.ApplicationCommandData().Options {
			if option.Name == "field" {
				key = option.StringValue()
			}

			if option.Name == "value" {
				val = option.StringValue()
			}
		}

		UpdateConfig(i.GuildID, cfg, key, val)
	}

	config, err := cfg.Queries.GetServerConfig(context.Background(), server.ID)
	if err != nil {
		log.Printf("error fetching config from DB: %s", err)
		return
	}

	var fields []*discordgo.MessageEmbedField
	for _, row := range config {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  row.Key,
			Value: row.Value,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("%s's Configuration", guild.ID),
		Fields: fields,
	}

	gd.QueueSendMessage(s, i.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			embed,
		},
	})
}
