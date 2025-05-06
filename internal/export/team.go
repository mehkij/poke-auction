package export

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mehkij/poke-auction/internal/types"
)

func ExportTeam(s *discordgo.Session, i *discordgo.InteractionCreate, players []*types.Player, gen int) {
	var d []byte

	var files []*discordgo.File
	for _, player := range players {
		for _, pokemon := range player.Team {
			switch gen {
			case 1:
				var moves string
				m := pokemon.RollMoves()
				for _, move := range m {
					moves += fmt.Sprintf("- %s\n", move)
				}

				var evs string
				var evals string
				for stat, ev := range pokemon.EVs {
					if len(pokemon.EVs) == 0 {
						break
					}

					evals += fmt.Sprintf("%d %s / ", ev, stat)
				}
				evs = fmt.Sprintf("EVs: %s", evals)
				evs = strings.Trim(evs, "/")

				var ivs string
				var ivals string
				for stat, iv := range pokemon.IVs {
					if len(pokemon.IVs) == 0 {
						break
					}

					ivals += fmt.Sprintf("%d %s / ", iv, stat)
				}
				ivs = fmt.Sprintf("IVs: %s", ivals)
				ivs = strings.Trim(ivs, "/")

				d = fmt.Appendf(d, "%s\n%s\n%s\n%s\nLevel: %d", pokemon.Name, evs, ivs, moves, pokemon.Level)

			case 8:
				var moves string
				m := pokemon.RollMoves()
				for _, move := range m {
					moves += fmt.Sprintf("- %s\n", move)
				}

				ability := pokemon.RollAbility()
				item := pokemon.RollItem()

				var evs string
				var evals string
				for stat, ev := range pokemon.EVs {
					if len(pokemon.EVs) == 0 {
						break
					}

					evals += fmt.Sprintf("%d %s / ", ev, stat)
				}
				evs = fmt.Sprintf("EVs: %s", evals)
				evs = strings.Trim(evs, "/")

				var ivs string
				var ivals string
				for stat, iv := range pokemon.IVs {
					if len(pokemon.IVs) == 0 {
						break
					}

					ivals += fmt.Sprintf("%d %s / ", iv, stat)
				}
				ivs = fmt.Sprintf("IVs: %s", ivals)
				ivs = strings.Trim(ivs, "/")

				d = fmt.Appendf(d, "%s @ %s\nAbility: %s\nLevel: %d\n%s\n%s\n%s\n\n", pokemon.Name, item, ability, pokemon.Level, evs, ivs, moves)

			case 9:
				role := pokemon.RollRandomRole()

				var moves string
				m := role.RollMoves()
				for _, move := range m {
					moves += fmt.Sprintf("- %s\n", move)
				}

				ability := role.RollAbility()
				item := role.RollItem()
				teratype := role.RollTeratype()

				var evs string
				var evals string
				for stat, ev := range role.EVs {
					if len(role.EVs) == 0 {
						break
					}

					evals += fmt.Sprintf("%d %s / ", ev, stat)
				}
				evs = fmt.Sprintf("EVs: %s", evals)
				evs = strings.Trim(evs, "/")

				var ivs string
				var ivals string
				for stat, iv := range role.IVs {
					if len(role.IVs) == 0 {
						break
					}

					ivals += fmt.Sprintf("%d %s / ", iv, stat)
				}
				ivs = fmt.Sprintf("IVs: %s", ivals)
				ivs = strings.Trim(ivs, "/")

				d = fmt.Appendf(d, "%s @ %s\n%s\n%s\n%s\n%s\n%s\nLevel: %d\n\n", pokemon.Name, item, ability, teratype, evs, ivs, moves, pokemon.Level)

			default:
				role := pokemon.RollRandomRole()

				var moves string
				m := role.RollMoves()
				for _, move := range m {
					moves += fmt.Sprintf("- %s\n", move)
				}

				ability := role.RollAbility()
				item := role.RollItem()

				var evs string
				var evals string
				for stat, ev := range role.EVs {
					if len(role.EVs) == 0 {
						break
					}

					evals += fmt.Sprintf("%d %s / ", ev, stat)
				}
				evs = fmt.Sprintf("EVs: %s", evals)
				evs = strings.Trim(evs, "/")

				var ivs string
				var ivals string
				for stat, iv := range role.IVs {
					if len(role.IVs) == 0 {
						break
					}

					ivals += fmt.Sprintf("%d %s / ", iv, stat)
				}
				ivs = fmt.Sprintf("IVs: %s", ivals)
				ivs = strings.Trim(ivs, "/")

				d = fmt.Appendf(d, "%s @ %s\nAbility: %s\nLevel: %d\n%s\n%s\n%s\n\n", pokemon.Name, item, ability, pokemon.Level, evs, ivs, moves)
			}
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
