package slash

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ritsec/competition-ops-bot-i/internal/logging"
	"github.com/ritsec/competition-ops-bot-i/internal/utils"
)

// startStatus sends the first response to an interaction. Returns false if
// the response failed, in which case later status updates will also fail.
func startStatus(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) bool {
	if err := utils.InitialMessage(s, i, message); err != nil {
		logging.Error(ctx, fmt.Sprintf("Could not respond to interaction: %v", err))
		return false
	}
	return true
}

// setStatus edits the interaction response, logging if the edit fails
func setStatus(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := utils.UpdateMessage(s, i, message); err != nil {
		logging.Error(ctx, fmt.Sprintf("Could not update interaction response: %v", err))
	}
}

// reportError logs err and shows it to the user in the interaction response
func reportError(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, message string, err error) {
	logging.Error(ctx, fmt.Sprintf("%s: %v", message, err))
	setStatus(ctx, s, i, fmt.Sprintf("%s: %v", message, err))
}

// optionsByName maps each top-level command option to its name, so handlers
// don't depend on the order Discord sends them in
func optionsByName(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, option := range i.ApplicationCommandData().Options {
		options[option.Name] = option
	}
	return options
}
