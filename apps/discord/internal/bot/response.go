package bot

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func initialMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) (error error) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
	return err
}

func updateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) (error error) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
	return err
}

func validateAction(s *discordgo.Session, i *discordgo.InteractionCreate) (verified bool) {
	var valid bool

	id := uuid.New().String()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000
	prompt := fmt.Sprintf("Please type %d to proceed with sending credentials.", code)

	closeChan := make(chan bool)
	defer close(closeChan)

	(ComponentHandlers)[id] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		input, err := strconv.Atoi(i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value)
		if err != nil || input != code {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Value not entered properly, cancelling request.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Fatal(err)
			}

			valid = false
		} else if input == code {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Verified",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Fatal(err)
			}

			valid = true
		}

		closeChan <- true
	}

	// defer delete(ComponentHandlers, id)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: id,
			Title:    "Verify",
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: prompt,
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "verify",
							Label:    "Verify",
							Style:    discordgo.TextInputShort,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	<-closeChan

	return valid
}
