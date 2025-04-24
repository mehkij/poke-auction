package cmd

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Callback    func(s *discordgo.Session, i *discordgo.InteractionCreate)
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

func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		for _, cmd := range AllCommands {
			if i.ApplicationCommandData().Name == cmd.Name {
				cmd.Callback(s, i)
				return
			}
		}

	case discordgo.InteractionMessageComponent:
		switch i.MessageComponentData().CustomID {
		case "join_auction":
			HandleAuctionInteraction(s, i)
		case "force_start":
			HandleForceStartAuction(s, i)
		}
	}
}

var AllCommands = []*Command{
	AuctionCommand,
	NominateCommand,
}
