package config

import (
	"fmt"
	"log/slog"
	"os"
)

type AppConfig struct {
	DiscordClientID            string
	DiscordBotToken            string
	DiscordPublicKey           string
	DiscordOauth2Token         string
	DiscordGatewayVersion      string
	DiscordVoiceGatewayVersion string
	DiscordHTTPBaseURL         string
	DiscordGatewayAddress      string
	AppEnv                     string
}

func LoadConfiguration() AppConfig {
	cfg := AppConfig{}
	requiredEnv := map[string]*string{
		"DC_APPLICATION_ID":        &cfg.DiscordClientID,
		"DC_BOT_TOKEN":             &cfg.DiscordBotToken,
		"DC_PUBLIC_KEY":            &cfg.DiscordPublicKey,
		"DC_OAUTH2_TOKEN":          &cfg.DiscordOauth2Token,
		"DC_GATEWAY_VERSION":       &cfg.DiscordGatewayVersion,
		"DC_VOICE_GATEWAY_VERSION": &cfg.DiscordVoiceGatewayVersion,
		"DC_HTTP_BASE_URL":         &cfg.DiscordHTTPBaseURL,
		"DC_GATEWAY_ADDRESS":       &cfg.DiscordGatewayAddress,
		"APP_ENV":                  &cfg.AppEnv,
	}
	for k, v := range requiredEnv {
		val, ok := os.LookupEnv(k)
		if !ok {
			slog.Error(fmt.Sprintf("missing required environment variable: %s", k))
			os.Exit(1)
		}
		*v = val
	}
	return cfg
}
