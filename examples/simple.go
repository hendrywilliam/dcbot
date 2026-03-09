package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hendrywilliam/siren/internal/config"
	"github.com/hendrywilliam/siren/internal/discord"
	"github.com/hendrywilliam/siren/internal/logger"
	"github.com/hendrywilliam/siren/internal/types"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Failed to load .env file")
		os.Exit(1)
	}

	env := config.LoadConfiguration()

	logOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	logHandler := logger.NewCustomHandler(os.Stdout, logger.CustomHandlerOpts{
		SlogOpts: logOpts,
	})
	log := slog.New(logHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	bot := discord.New(discord.DiscordArgs{
		BotToken:   env.DiscordBotToken,
		ClientID:   env.DiscordClientID,
		BotVersion: 10,
		BotIntent: []int{
			discord.GuildsIntent,
			discord.GuildMessagesIntent,
			discord.GuildVoiceStatesIntent,
			discord.MessageContentIntent,
		},
		Logger:   log,
		Handlers: make(map[string][]any),
	})

	bot.RegisterHandler(discord.EventMessageCreate, func(ctx context.Context, event interface{}) error {
		msg, ok := event.(types.Message)
		if !ok {
			log.Error("Failed to cast event to Message type")
			return nil
		}

		if msg.Author.ID == env.DiscordClientID {
			return nil
		}

		if msg.Content == "!ping" {
			log.Info("Received ping command", "user", msg.Author.Username, "channel", msg.ChannelID)
		} else if msg.Content == "!help" {
			log.Info("Received help command", "user", msg.Author.Username)
		}

		return nil
	})

	bot.RegisterHandler(discord.EventInteractionCreate, func(ctx context.Context, event interface{}) error {
		interaction, ok := event.(types.Interaction)
		if !ok {
			log.Error("Failed to cast event to Interaction type")
			return nil
		}

		log.Info("Interaction received",
			"type", interaction.Type,
			"user", interaction.Member.User.Username,
			"guild", interaction.GuildID)

		switch interaction.Type {
		case types.InteractionTypePing:
			log.Info("Ping interaction received")
		case types.InteractionTypeApplicationCommand:
			log.Info("Application command received", "command", interaction.Data.Name)
			if interaction.Data.Name == "play" {
				log.Info("Play command executed")
			}
		}

		return nil
	})

	bot.RegisterHandler(discord.EventReady, func(ctx context.Context, event interface{}) error {
		ready, ok := event.(types.ReadyEvent)
		if !ok {
			log.Error("Failed to cast event to ReadyEvent type")
			return nil
		}

		log.Info("Bot is ready!",
			"session_id", ready.SessionID,
			"gateway_url", ready.ResumeGatewayURL)

		return nil
	})

	bot.RegisterHandler(discord.EventVoiceStateUpdate, func(ctx context.Context, event interface{}) error {
		voiceState, ok := event.(types.VoiceState)
		if !ok {
			log.Error("Failed to cast event to VoiceState type")
			return nil
		}

		log.Info("Voice state updated",
			"user", voiceState.UserID,
			"guild", voiceState.GuildID,
			"channel", voiceState.ChannelID,
			"mute", voiceState.SelfMute,
			"deaf", voiceState.SelfDeaf)

		if voiceState.ChannelID == "" {
			log.Info("User left voice channel", "user", voiceState.UserID)
		} else {
			log.Info("User joined voice channel",
				"user", voiceState.UserID,
				"channel", voiceState.ChannelID)
		}

		return nil
	})

	bot.RegisterHandler(discord.EventVoiceServerUpdate, func(ctx context.Context, event interface{}) error {
		voiceServer, ok := event.(types.VoiceServerUpdate)
		if !ok {
			log.Error("Failed to cast event to VoiceServerUpdate type")
			return nil
		}

		log.Info("Voice server updated",
			"guild", voiceServer.GuildID,
			"endpoint", voiceServer.Endpoint,
			"token", "[REDACTED]")

		return nil
	})

	bot.RegisterHandler(discord.EventGuildCreate, func(ctx context.Context, event interface{}) error {
		log.Info("Joined new guild or guild data received")
		return nil
	})

	log.Info("Connecting to Discord...")
	if err := bot.Open(ctx); err != nil {
		log.Error("Failed to connect to Discord", "error", err)
		os.Exit(1)
	}

	log.Info("Bot is running. Press Ctrl+C to stop.")

	<-ctx.Done()
	log.Info("Shutting down...")
	bot.Close()
	log.Info("Bot shutdown complete")
}

func createMessageFilterHandler(filterWords []string) func(context.Context, interface{}) error {
	return func(ctx context.Context, event interface{}) error {
		msg, ok := event.(types.Message)
		if !ok {
			return nil
		}

		for _, word := range filterWords {
			if containsWord(msg.Content, word) {
				slog.Info("Message filtered",
					"user", msg.Author.Username,
					"word", word,
					"channel", msg.ChannelID)
			}
		}

		return nil
	}
}

func containsWord(content, word string) bool {
	return len(word) > 0 && len(content) >= len(word)
}
