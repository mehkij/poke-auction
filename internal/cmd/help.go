package cmd

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/types"
)

var CommandDescriptions = map[string]types.CommandDescription{
	"help": {
		BasicDescription:    "Get help on what commands are available and how to use them! Use '/help help' for more details.",
		DetailedDescription: "Using '/help [command name]' displays a detailed description (with examples) of how to use a command.",
	},

	"auction": {
		BasicDescription:    "Start a Pokemon auction.",
		DetailedDescription: "/auction [generation] [timer]\n\n '/auction 6 120' starts a Gen 6 auction with a 120 second timer before the auction begins.",
	},

	"nominate": {
		BasicDescription:    "Nominate a Pokemon.",
		DetailedDescription: "/nominate [pokemon]\n\n '/nominate pikachu' nominates Pikachu for bidding. Note that this is case-insensitive! So '/nominate PiKaChU' works completely fine!",
	},

	"bid": {
		BasicDescription:    "Bid on a Pokemon.",
		DetailedDescription: "/bid [amount]\n\n '/bid 500' bids 500 PokeDollars on a currently nominated Pokemon. Please note that your bid must be higher than the currently active bid! So if the current highest bid is 500, you must bid above 500 for your bid to go through.",
	},

	"pick": {
		BasicDescription:    "You're the last one standing, pick your Pokemon!",
		DetailedDescription: "/pick [pokemon]\n\n '/pick Pikachu' picks Pikachu as the Pokemon to add to your team. This command is used for when you're the last person left without a full team of Pokemon, and cannot be used otherwise.",
	},

	"stopall": {
		BasicDescription:    "Stop all running Pokemon auctions in the current channel.",
		DetailedDescription: "/stopall.\n\n This command terminates all currently active auctions in the channel the command is sent in.",
	},

	"config": {
		BasicDescription:    "Edit your server's bot configuration.",
		DetailedDescription: "/config [field] [value]\n\n This command comes with two optional parameters: field and value. Using '/config' by itself displays a list of all available configuration options for the bot. To edit the value of a config option, you can do something like this (let's use BidTimerDuration as an example): '/config BidTimerDuration 20' (this sets the amount of time allowed before a bidding phase is over to 20 seconds).",
	},
}

func HelpCallback(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *types.GlobalConfig) {
	gd := cfg.GlobalDispatcher

	if i.ApplicationCommandData().Options == nil {
		var fields []*discordgo.MessageEmbedField

		for key, cd := range CommandDescriptions {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  fmt.Sprintf("/%s", key),
				Value: cd.BasicDescription,
			})
		}

		embed := &discordgo.MessageEmbed{
			Title:  "Commands List",
			Fields: fields,
		}

		gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		})
	} else {
		commandName := i.ApplicationCommandData().Options[0].Value.(string)

		desc, exists := CommandDescriptions[commandName]
		if !exists {
			gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Command not found!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("/%s", commandName),
			Description: desc.DetailedDescription,
		}

		gd.QueueInteractionResponse(s, i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		})
	}
}
