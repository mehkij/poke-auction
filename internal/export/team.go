package export

import (
	"bytes"
	"fmt"
	"log"
	"os"

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

		filename, data, err := WriteToFile(d, player.Username)
		if err != nil {
			log.Printf("an error occurred while trying to export team: %v", err)
			return
		}

		files = append(files, &discordgo.File{
			Name:        filename,
			ContentType: "text/plain",
			Reader:      bytes.NewReader(data),
		})
	}

	msg := &discordgo.MessageSend{
		Content: "Here are your teams!",
		Files:   files,
	}

	s.ChannelMessageSendComplex(i.ChannelID, msg)
}

func WriteToFile(data []byte, username string) (string, []byte, error) {
	teamsDir := "teams"
	if err := os.MkdirAll(teamsDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create teams directory: %w", err)
	}

	filename := fmt.Sprintf("%s_team.txt", username)
	filepath := fmt.Sprintf("%s/%s", teamsDir, filename)

	err := os.WriteFile(filepath, data, 0644)
	if err != nil {
		return "", nil, err
	}

	log.Printf("file successfully created: %s_team.txt", username)
	return filename, data, nil
}
