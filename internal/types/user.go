package types

import (
	"fmt"
	"log/slog"
)

type User struct {
	ID                   string      `json:"id"`
	Username             string      `json:"username"`
	Discriminator        string      `json:"discriminator"`
	GlobalName           string      `json:"global_name,omitempty"`
	Avatar               string      `json:"avatar"`
	Bot                  bool        `json:"bot,omitempty"`
	System               bool        `json:"system,omitempty"`
	PublicFlags          uint8       `json:"public_flags"`
	Email                string      `json:"email,omitempty"`
	Clan                 interface{} `json:"clan,omitempty"`
	AvatarDecorationData interface{} `json:"avatar_decoration_data,omitempty"`
}

func (u *User) Mention() string {
	return fmt.Sprintf("<@%s>", u.ID)
}

func (u *User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", u.ID),
		slog.String("username", u.Username),
		slog.String("discriminator", u.Discriminator),
		slog.String("global_name", u.GlobalName),
	)
}
