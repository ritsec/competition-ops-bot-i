package utils

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

// InitialMessage responds to an interaction with an ephemeral message
func InitialMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: message,
		},
	})
	return err
}

// UpdateMessage edits the response previously sent with InitialMessage
func UpdateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
	return err
}

// ValidateAction shows a modal asking the user to type a random code and
// blocks until they submit it. Returns true only if the code matches.
func ValidateAction(s *discordgo.Session, i *discordgo.InteractionCreate) (verified bool) {
	id := uuid.New().String()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000
	prompt := fmt.Sprintf("Please type %d to proceed with sending credentials.", code)

	// Buffered so a late submit never blocks after we've stopped waiting
	result := make(chan bool, 1)

	RegisterComponent(id, func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		input, err := strconv.Atoi(modalValue(i.ModalSubmitData().Components, "verify"))
		valid := err == nil && input == code

		message := "Value not entered properly, cancelling request."
		if valid {
			message = "Verified"
		}

		if err := InitialMessage(s, i, message); err != nil {
			log.Println(err)
		}

		result <- valid
	})
	defer UnregisterComponent(id)

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
		log.Println(err)
		return false
	}

	// Modals don't notify on dismiss, so stop waiting eventually
	select {
	case valid := <-result:
		return valid
	case <-time.After(5 * time.Minute):
		return false
	}
}
