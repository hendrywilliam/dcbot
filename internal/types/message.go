package types

import "fmt"

type Message struct {
	ID                   string   `json:"id"`
	ChannelID            string   `json:"channel_id"`
	Author               User     `json:"author"`
	Content              string   `json:"content"`
	Timestamp            string   `json:"timestamp"`
	EditedTimestamp      string   `json:"edited_timestamp,omitempty"`
	Tts                  bool     `json:"tts"`
	MentionEveryone      bool     `json:"mention_everyone"`
	Mentions             []User   `json:"mentions"`
	MentionRoles         []string `json:"mention_roles"`
	MentionChannels      []any    `json:"mention_channels,omitempty"`
	Attachments          []any    `json:"attachments"`
	Embeds               []any    `json:"embeds"`
	Reactions            []any    `json:"reactions,omitempty"`
	Nonce                any      `json:"nonce,omitempty"`
	Pinned               bool     `json:"pinned"`
	WebhookID            string   `json:"webhook_id,omitempty"`
	Type                 int      `json:"type"`
	Activity             any      `json:"activity,omitempty"`
	Application          any      `json:"application,omitempty"`
	ApplicationID        string   `json:"application_id,omitempty"`
	MessageReference     any      `json:"message_reference,omitempty"`
	Flags                int      `json:"flags,omitempty"`
	ReferencedMessage    *Message `json:"referenced_message,omitempty"`
	Interaction          any      `json:"interaction,omitempty"`
	Thread               any      `json:"thread,omitempty"`
	Components           []any    `json:"components,omitempty"`
	StickerItems         []any    `json:"sticker_items,omitempty"`
	Position             int      `json:"position,omitempty"`
	RoleSubscriptionData any      `json:"role_subscription_data,omitempty"`
	Resolved             any      `json:"resolved,omitempty"`
	Poll                 any      `json:"poll,omitempty"`
}

func (m Message) Mention() string {
	return fmt.Sprintf("<@%s>", m.Author.ID)
}
