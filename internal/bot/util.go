package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/https-whoyan/MafiaBot/internal/core/config"
	"log"
)

func Bold(s string) string {
	return "**" + s + "**"
}

func Italic(s string) string {
	return "_" + s + "_"
}

func Emphasized(s string) string {
	return "__" + s + "__"
}

func CodeBlock(language, text string) string {
	return "```" + language + text + "```"
}

func noticeChat(s *discordgo.Session, chatID string, content ...string) (map[string]*discordgo.Message, error) {
	// Represent Message by their content.
	messages := make(map[string]*discordgo.Message)
	for _, onceContent := range content {
		message, err := s.ChannelMessageSend(chatID, onceContent)
		if err != nil {
			return nil, err
		}
		messages[onceContent] = message
	}
	return messages, nil
}

func isPrivateMessage(i *discordgo.InteractionCreate) bool {
	return i.GuildID == ""
}

func noticePrivateChat(s *discordgo.Session, i *discordgo.InteractionCreate) {
	content := Bold("All commands are used on the server.\n") + "If you have difficulties in using the bot, " +
		"please refer to the repository documentation: https://github.com/https-whoyan/MafiaBot"
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Println(errors.Join(
			errors.New("there was an error when sending a private message, err: "), err),
		)
	}
}

func noticeIsEmptyGame(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "You can't interact with the game because you haven't registered it\n" +
				Bold("Write the "+Emphasized("/register_game")+" command") + " to start the game.",
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func CreateConfigMessage(cfg *config.RolesConfig) string {
	return ""
}
