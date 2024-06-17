package bot

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	gamePack "github.com/https-whoyan/MafiaBot/core/game"
	botFMTPack "github.com/https-whoyan/MafiaBot/internal/fmt"
	botGamePack "github.com/https-whoyan/MafiaBot/internal/game"
	handlerPack "github.com/https-whoyan/MafiaBot/internal/handlers"
	userPack "github.com/https-whoyan/MafiaBot/internal/user"

	"github.com/bwmarrin/discordgo"
)

// ____________
// BotConfig
// ____________

type BotConfig struct {
	token string
}

func LoadBotConfig() *BotConfig {
	token := os.Getenv("BOT_TOKEN")
	return &BotConfig{
		token: token,
	}
}

// ________
// Bot
// ________

var (
	botOnce     sync.Once
	botInstance *Bot
)

type Bot struct {
	// DiscordGo token
	token string
	// DiscordGo Session
	Session *discordgo.Session
	// Seen from https://github.com/bwmarrin/discordgo/tree/master/examples/slash_commands
	// The key is the name of the command.
	Commands map[string]handlerPack.Command
	// Games this a map,
	// the key in which is the State of the server where the bot is running,
	// and the value is the game.
	Games map[string]*gamePack.Game
	// To save DiscordGo.ApplicationCommand's for closing deleting.
	registeredCommands []*discordgo.ApplicationCommand
	// To format messages.
	//
	// Implement of FmtInterface.
	FMTer *botFMTPack.DiscordFMTer
}

func InitBot(cnf *BotConfig) {
	botOnce.Do(func() {
		token := cnf.token
		botStr := "Bot " + token
		s, err := discordgo.New(botStr)
		if err != nil {
			log.Fatal(err)
		}

		bot := &Bot{
			token:    token,
			Session:  s,
			Commands: make(map[string]handlerPack.Command),
			Games:    make(map[string]*gamePack.Game),
			FMTer:    botFMTPack.DiscordFMTInstance,
		}

		bot.initBotCommands()
		bot.registerHandlers()
		botInstance = bot
	})
}

func Run() {
	if botInstance == nil {
		log.Fatal("Bot isn't instance!")
	}
	err := botInstance.Open()
	if err != nil {
		log.Fatal(err)
	}
	botInstance.loginAs()
	botInstance.registerCommands()

	// If you need delete all registered commands, use here: bot.DeleteAllGloballyRegisteredCommands()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func DisconnectBot() error {
	if botInstance == nil {
		return errors.New("bot isn't initialized")
	}
	return botInstance.Close()

}

func (b *Bot) loginAs() {
	b.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
}

func (b *Bot) Open() error {
	err := b.Session.Open()
	if err != nil {
		return err
	}
	b.loginAs()
	return nil
}

func (b *Bot) Close() error {
	b.DeleteHandlers()
	b.removeRegisteredCommands()
	err := b.Session.Close()
	if err != nil {
		return err
	}
	return nil
}

// ____________________________________________________________
// All below functions initialize or delete the required variables.
// ____________________________________________________________

func (b *Bot) initCommand(c handlerPack.Command) {
	commandName := c.GetName()
	b.Commands[commandName] = c
}

func (b *Bot) initBotCommands() {
	// Channels
	b.initCommand(handlerPack.NewAddMainChannelCommand())
	b.initCommand(handlerPack.NewAddChannelRoleCommand())

	// Game
	b.initCommand(handlerPack.NewRegisterGameCommand())
	b.initCommand(handlerPack.NewChoiceGameConfigCommand())
	b.initCommand(handlerPack.NewStartGameCommand())

	// Other
	b.initCommand(handlerPack.NewYanLohCommand())
	b.initCommand(handlerPack.NewAboutRolesCommand())

}

func (b *Bot) registerHandlers() {
	log.Print("Register handlers")
	for _, cmd := range b.Commands {
		// To avoid closing the loop
		newCmd := cmd
		cmdName := newCmd.GetName()
		log.Printf("Register handler, command name: %v", cmdName)
		// Lock
		newHandler := b.getSIHandler(newCmd, cmdName)
		b.Session.AddHandler(newHandler)
	}
}

// getSIHandler Is bot command handler get Function.
// All comments in function.
func (b *Bot) getSIHandler(cmd handlerPack.Command, cmdName string) func(
	s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// I recognize the name of the team
		executedCommandName := i.ApplicationCommandData().Name

		// If it not equals as iterable cmd.Name
		if executedCommandName != cmdName {
			return
		}

		// If it executed in private chat
		if handlerPack.IsPrivateMessage(i) {
			// Reply "it is a private chat"
			handlerPack.NoticePrivateChat(s, i, b.FMTer)
			return
		}
		log.Printf("Executed guild ID: %v", i.GuildID)

		// If command use not use for game interaction
		if !cmd.IsUsedForGame() {
			// Just execute a Execute()
			log.Printf("Execute %v command.", cmdName)
			cmd.Execute(s, i.Interaction, nil, b.FMTer)
			return
		}

		// I know the guildID
		executedGuildID := i.GuildID
		// And is there a game on this server
		_, containsGame := b.Games[executedGuildID]

		// If yes
		if containsGame {
			log.Printf("Execute %v command.", cmdName)
			currGame := b.Games[executedGuildID]
			// Validate Is correct command by game state
			content, isOk := handlerPack.ValidateCommandByGameState(executedCommandName, currGame, b.FMTer)
			if !isOk {
				handlerPack.Response(s, i.Interaction, content)
				return
			}
			// If ok, I call the Execute method of the command
			cmd.Execute(s, i.Interaction, currGame, b.FMTer)
			return
		}

		// Otherwise I know the game isn't registered.
		// I check to see if the command name is register_game. If not, it means that the
		// person uses the game command without registering it.
		if executedCommandName != handlerPack.RegisterGameCommandName {
			handlerPack.NoticeIsEmptyGame(s, i, b.FMTer)
			return
		}

		// I use the register_game command
		log.Printf("Must be register_game: Execute %v command.", cmdName)

		// (Get UserRename provider)
		userRenameProvider := userPack.NewBotUserRenameProvider(s, executedGuildID)
		gameConfig := botGamePack.GetNewGameConfig(userRenameProvider)

		b.Games[executedGuildID] = gamePack.GetNewGame(executedGuildID, gameConfig...)
		content, isOk := handlerPack.ValidateCommandByGameState(executedCommandName, b.Games[executedGuildID], b.FMTer)
		if !isOk {
			handlerPack.Response(s, i.Interaction, content)
			return
		}
		log.Printf("Registered new game by %v Guild ID", executedGuildID)
		cmd.Execute(s, i.Interaction, b.Games[executedGuildID], b.FMTer)

		return
	}
}

func (b *Bot) registerCommands() {
	log.Println("Register commands")
	stateId := b.Session.State.User.ID
	for _, cmd := range b.Commands {
		newCmd := cmd
		registeredCmd, err := b.Session.ApplicationCommandCreate(stateId, "", newCmd.GetCmd())
		if err != nil {
			log.Print(err)
		}
		b.registeredCommands = append(b.registeredCommands, registeredCmd)
		log.Printf("Registered command, name %v", registeredCmd.Name)
	}
}

func (b *Bot) removeRegisteredCommands() {
	b.DeleteAllGloballyRegisteredCommands()
}

// DeleteAllGloballyRegisteredCommands Delete all registered to bot functions. Globally Registered
func (b *Bot) DeleteAllGloballyRegisteredCommands() {
	log.Println("Init Delete all globally registered commands.")
	userId := b.Session.State.User.ID
	globallyRegisteredCommands, err := b.Session.ApplicationCommands(userId, "")
	if err != nil {
		log.Fatal(err)
	}
	for _, command := range globallyRegisteredCommands {
		log.Printf("Removed command ID: %v", command.ID)
		err = b.Session.ApplicationCommandDelete(userId, "", command.ID)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("All global commands deleted.")
}

func (b *Bot) DeleteHandlers() {
	err := b.Session.AddHandler(nil)
	if err != nil { // :))))
		log.Println("Delete all handlers")
		return
	}
}