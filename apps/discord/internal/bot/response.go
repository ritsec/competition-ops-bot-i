package bot

import "github.com/bwmarrin/discordgo"

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
