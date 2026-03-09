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
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Failed to load configuration file.")
		os.Exit(1)
	}

	env := config.LoadConfiguration()

	var logHandler slog.Handler
	logOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	if env.AppEnv != "production" {
		logHandler = logger.NewCustomHandler(os.Stdout, logger.CustomHandlerOpts{
			SlogOpts: logOpts,
		})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &logOpts)
	}

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
		log.Info("Custom handler for MESSAGE_CREATE called")
		return nil
	})

	bot.RegisterHandler(discord.EventInteractionCreate, func(ctx context.Context, event interface{}) error {
		log.Info("Custom handler for INTERACTION_CREATE called")
		return nil
	})

	bot.RegisterHandler(discord.EventReady, func(ctx context.Context, event interface{}) error {
		log.Info("Custom handler for READY called - Bot is ready!")
		return nil
	})

	bot.RegisterHandler(discord.EventVoiceStateUpdate, func(ctx context.Context, event interface{}) error {
		log.Info("Custom handler for VOICE_STATE_UPDATE called")
		return nil
	})

	bot.RegisterHandler(discord.EventVoiceServerUpdate, func(ctx context.Context, event interface{}) error {
		log.Info("Custom handler for VOICE_SERVER_UPDATE called")
		return nil
	})

	if err := bot.Open(ctx); err != nil {
		log.Error("Failed to open Discord connection", "error", err)
		os.Exit(1)
	}

	log.Info("Discord bot started successfully")

	<-ctx.Done()
	log.Info("Shutting down Discord bot...")
	bot.Close()
	log.Info("Discord bot shutdown complete")
}
