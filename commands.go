package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/cmd"
	"github.com/mehkij/poke-auction/internal/types"
)

type Command struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Callback    func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *types.GlobalConfig)
}

func RegisterAll(s *discordgo.Session, appID, guildID string) []*discordgo.ApplicationCommand {
	var cmds []*discordgo.ApplicationCommand

	for _, cmd := range AllCommands {
		ac, err := s.ApplicationCommandCreate(appID, guildID, &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
		if err != nil {
			log.Printf("Failed to register command: '%s': %v", cmd.Name, err)
		} else {
			log.Printf("Registered command: %s", cmd.Name)
		}

		cmds = append(cmds, ac)
	}

	return cmds
}

func NewInteractionHandler(cfg *types.GlobalConfig) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			for _, cmd := range AllCommands {
				if i.ApplicationCommandData().Name == cmd.Name {
					cmd.Callback(s, i, cfg)
					return
				}
			}
		case discordgo.InteractionMessageComponent:
			switch i.MessageComponentData().CustomID {
			case "join_auction":
				cmd.HandleAuctionInteraction(s, i, cfg.GlobalDispatcher)
			case "force_start":
				cmd.HandleForceStartAuction(s, i, cfg.GlobalDispatcher)
			}
		}
	}
}

var AllCommands = []*Command{
	AuctionCommand,
	NominateCommand,
	BidCommand,
	StopAllCommand,
}

// Command definitions

// "/auction"
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
	Callback: cmd.AuctionCallback,
}

// "/nominate"
var NominateCommand = &Command{
	Name:        "nominate",
	Description: "Nominate a Pokemon.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "The name of the Pokemon.",
			Required:    true,
		},
	},
	Callback: cmd.NominateCallback,
}

// "/bid"
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
	Callback: cmd.BidCallback,
}

// "/stopall"
var StopAllCommand = &Command{
	Name:        "stopall",
	Description: "Stop all running Pokemon auctions in the current channel.",
	Callback:    cmd.StopAllCallback,
}

// "/config"
var ConfigCommand = &Command{
	Name:        "config",
	Description: "Edit your server's bot configuration.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "field",
			Description: "The name of the config field that you want to change.",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "value",
			Description: "The value you want to insert into the target field.",
			Required:    false,
		},
	},
	Callback: cmd.ConfigCallback,
}
