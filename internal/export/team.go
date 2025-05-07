package export

import (
	"bytes"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/types"
)

func ExportTeam(s *discordgo.Session, i *discordgo.InteractionCreate, players []*types.Player, gen int) {
	var files []*discordgo.File
	for _, player := range players {
		var d []byte

		for _, pokemon := range player.Team {
			d = fmt.Appendf(d, "%s\n\n", pokemon.Name)
		}

		files = append(files, &discordgo.File{
			Name:        fmt.Sprintf("%s_team.txt", player.Username),
			ContentType: "text/plain",
			Reader:      bytes.NewReader(d),
		})
	}

	msg := &discordgo.MessageSend{
		Content: "Here are your teams!",
		Files:   files,
	}

	s.ChannelMessageSendComplex(i.ChannelID, msg)
}
