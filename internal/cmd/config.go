package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"

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

func ValidateValueOption(val string, field reflect.StructField) bool {
	switch field.Type.Kind() {
	case reflect.Int:
		if _, err := strconv.Atoi(val); err == nil {
			log.Println("Valid int!")
			return true
		} else {
			log.Println("Invalid int")
			return false
		}

	case reflect.String:
		log.Println("Valid string!")
		return true

	default:
		log.Println("Unsupported field type")
		return false
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
	_, err = cfg.Queries.GetServer(context.Background(), i.GuildID)

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

		botConfigType := reflect.TypeOf(types.BotConfig{})

		var key string
		var val string

		for _, option := range i.ApplicationCommandData().Options {
			if option.Name == "field" {
				// Validate if the user input matches the config field name
				found := false
				for i := 0; i < botConfigType.NumField(); i++ {
					if botConfigType.Field(i).Name == option.StringValue() {
						found = true
					}
				}

				if !found {
					log.Println("invalid field name passed to config command.")
					// utils.CreateFollowupEphemeralError(s, i, "Invalid field name!")
					gd.QueueFollowupMessage(s, i, "Invalid field name!", discordgo.MessageFlagsEphemeral)
					return
				} else {
					key = option.StringValue()
				}

			}

			if option.Name == "value" {
				// Validate if the user input matches the config field name
				field, ok := botConfigType.FieldByName(key)
				if !ok {
					log.Println("field not found")
					return
				}

				ok = ValidateValueOption(option.StringValue(), field)
				if ok {
					val = option.StringValue()
				} else {
					log.Println("invalid value passed to config field.")
					// utils.CreateFollowupEphemeralError(s, i, "Invalid value!")
					gd.QueueFollowupMessage(s, i, "Invalid value!", discordgo.MessageFlagsEphemeral)
					return
				}
			}
		}

		if key != "" && val != "" {
			UpdateConfig(i.GuildID, cfg, key, val)
		}
	}

	config, err := cfg.Queries.GetServerConfig(context.Background(), i.GuildID)
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
		Title:  fmt.Sprintf("%s's Configuration", guild.Name),
		Fields: fields,
	}

	gd.QueueSendMessage(s, i.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			embed,
		},
	})
}
