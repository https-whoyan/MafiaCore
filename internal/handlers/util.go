package bot

import (
	"errors"
	"github.com/https-whoyan/MafiaBot/core/converter"
	"log"
	"strings"

	coreGamePack "github.com/https-whoyan/MafiaBot/core/game"
	coreRolePack "github.com/https-whoyan/MafiaBot/core/roles"
	botChannelPack "github.com/https-whoyan/MafiaBot/internal/channel"
	botConvertedPack "github.com/https-whoyan/MafiaBot/internal/converter"
	botFMT "github.com/https-whoyan/MafiaBot/internal/fmt"

	"github.com/https-whoyan/MafiaBot/pkg/repository/mongo"

	"github.com/bwmarrin/discordgo"
)

func isCorrectChatID(s *discordgo.Session, chatID string) bool {
	if s.TryLock() {
		defer s.Unlock()
	}

	ch, err := s.Channel(chatID)
	return err == nil && ch != nil
}

// Send message to chatID
func sendMessages(s *discordgo.Session, chatID string, content ...string) (map[string]*discordgo.Message, error) {
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

// SendToUser Send to userID a message
func SendToUser(s *discordgo.Session, userID string, msg string) error {
	// Create a channel
	channel, err := s.UserChannelCreate(userID)
	if err != nil || channel == nil {
		if channel == nil {
			return errors.New("channel Create Failed, empty channel")
		}
		return err
	}
	channelID := channel.ID
	_, err = s.ChannelMessageSend(channelID, msg)
	return err
}

// ___________________
// Response func
// ___________________

// Response reply to interaction by provided content (s.InteractionResponse)
func Response(s *discordgo.Session, i *discordgo.Interaction, content string) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Print(err)
	}
}

// ____________________
// Error responses
// ____________________

// IsPrivateMessage finds out if a message has been sent to private messages
func IsPrivateMessage(i *discordgo.InteractionCreate) bool {
	return i.GuildID == ""
}

func NoticePrivateChat(s *discordgo.Session, i *discordgo.InteractionCreate, fMTer *botFMT.DiscordFMTer) {
	content := fMTer.Bold("All commands are used on the server.") + fMTer.NL() +
		"If you have difficulties in using the bot, " +
		"please refer to the repository documentation: https://github.com/https-whoyan/MafiaBot"
	Response(s, i.Interaction, content)
}

// NoticeIsEmptyGame If game not exists
func NoticeIsEmptyGame(s *discordgo.Session, i *discordgo.InteractionCreate, fMTer *botFMT.DiscordFMTer) {
	content := "You can't interact with the game because you haven't registered it" + fMTer.NL() +
		fMTer.Bold("Write the "+fMTer.Underline(RegisterGameCommandName)+" command") + " to start the game."
	Response(s, i.Interaction, content)
}

// __________________
// Channels
// ___________________

// SetRolesChannels to game.
func setRolesChannels(s *discordgo.Session, guildID string, g *coreGamePack.Game) ([]string, error) {
	// Get night interaction roles names
	allRolesNames := coreRolePack.GetAllNightInteractionRolesNames()
	// Get curr MongoDB struct
	currDB, isContains := mongo.GetCurrMongoDB()
	if !isContains {
		return []string{}, errors.New("MongoDB doesn't initialized")
	}
	// emptyRolesMp: save not contains channel roles
	emptyRolesMp := make(map[string]bool)
	// mappedRoles: save contains channels roles
	mappedRoles := make(map[string]*botChannelPack.BotRoleChannel)

	addNewChannelIID := func(roleName, channelName string) {
		channelIID, err := currDB.GetChannelIIDByRole(guildID, channelName)
		if channelIID == "" {
			emptyRolesMp[channelName] = true
			return
		}
		newRoleChannel, err := botChannelPack.NewBotRoleChannel(s, channelIID, roleName)
		if err != nil {
			emptyRolesMp[channelName] = true
			return
		}
		mappedRoles[roleName] = newRoleChannel
	}

	for _, roleName := range allRolesNames {
		if strings.ToLower(roleName) == strings.ToLower(coreRolePack.Don.Name) {
			addNewChannelIID(roleName, coreRolePack.Mafia.Name)
			continue
		}
		addNewChannelIID(roleName, roleName)
	}
	// If a have all roles
	if len(emptyRolesMp) == 0 {
		// Convert
		sliceMappedRoles := botConvertedPack.GetMapValues(mappedRoles)
		InterfaceRoleChannelSlice := botConvertedPack.ConvertRoleChannelsSliceToIChannelSlice(sliceMappedRoles)

		// Save it to g.RoleChannels.
		err := g.SetRoleChannels(InterfaceRoleChannelSlice)

		return []string{}, err
	}

	return converter.GetMapKeys(emptyRolesMp), nil
}

// Check, if main channel exists or not
func existsMainChannel(guildID string) bool {
	currMongo, exists := mongo.GetCurrMongoDB()
	if !exists {
		return false
	}
	channelIID, err := currMongo.GetMainChannelIID(guildID)
	if err != nil {
		return false
	}
	return channelIID != ""
}

func setMainChannel(s *discordgo.Session, guildID string, g *coreGamePack.Game) {
	currMongo, _ := mongo.GetCurrMongoDB()
	channelIID, _ := currMongo.GetMainChannelIID(guildID)
	mainChannel, _ := botChannelPack.NewBotMainChannel(s, channelIID)
	_ = g.SetMainChannel(mainChannel)
}