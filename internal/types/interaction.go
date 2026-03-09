package types

import "time"

type InteractionType = uint8

const (
	InteractionTypePing                           InteractionType = 1
	InteractionTypeApplicationCommand             InteractionType = 2
	InteractionTypeMessageComponent               InteractionType = 3
	InteractionTypeApplicationCommandAutocomplete InteractionType = 4
	InteractionTypeModalSubmit                    InteractionType = 5
)

type InteractionContextType = uint8

const (
	InteractionContextTypeGuild InteractionContextType = 0
	InteractionContextTypeBotDM InteractionContextType = 1
	InteractionPrivateChannel   InteractionContextType = 2
)

type InteractionApplicationCommandData struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     uint        `json:"type"`
	Resolved interface{} `json:"resolved,omitempty"`
	Options  interface{} `json:"options,omitempty"`
	GuildID  string      `json:"guild_id,omitempty"`
	TargetID string      `json:"target_id,omitempty"`
}

type ChannelType = uint8

const (
	ChannelTypeGuildText          ChannelType = 0
	ChannelTypeDM                 ChannelType = 1
	ChannelTypeGuildVoice         ChannelType = 2
	ChannelTypeGroupDM            ChannelType = 3
	ChannelTypeGuildCategory      ChannelType = 4
	ChannelTypeGuildAnnouncement  ChannelType = 5
	ChannelTypeAnnouncementThread ChannelType = 10
	ChannelTypePublicThread       ChannelType = 11
	ChannelTypePrivateThread      ChannelType = 12
	ChannelTypeGuildStageVoice    ChannelType = 13
	ChannelTypeGuildDirectory     ChannelType = 14
	ChannelTypeGuildForum         ChannelType = 15
	ChannelTypeGuildMedia         ChannelType = 16
)

type Channel struct {
	ID                            string        `json:"id"`
	Type                          ChannelType   `json:"type"`
	GuildID                       string        `json:"guild_id"`
	Position                      uint          `json:"position,omitempty"`
	PermissionOverwrites          []interface{} `json:"permission_overwrites,omitempty"`
	Name                          string        `json:"name,omitempty"`
	Topic                         string        `json:"topic,omitempty"`
	Nsfw                          bool          `json:"nsfw,omitempty"`
	LastMessageID                 string        `json:"last_message_id,omitempty"`
	Bitrate                       uint          `json:"bitrate,omitempty"`
	UserLimit                     uint          `json:"user_limit,omitempty"`
	RateLimitPerUser              uint          `json:"rate_limit_per_user,omitempty"`
	Recipients                    []interface{} `json:"recipients,omitempty"`
	Icon                          string        `json:"icon,omitempty"`
	OwnerID                       string        `json:"owner_id,omitempty"`
	ApplicationID                 string        `json:"application_id,omitempty"`
	Managed                       bool          `json:"managed,omitempty"`
	ParentID                      string        `json:"parent_id,omitempty"`
	LastPinTimestamp              time.Time     `json:"last_pin_timestamp,omitempty"`
	RTCRegion                     string        `json:"rtc_region,omitempty"`
	VideoQualityMode              uint          `json:"video_quality_mode,omitempty"`
	MessageCount                  uint          `json:"message_count,omitempty"`
	MemberCount                   uint          `json:"member_count,omitempty"`
	ThreadMetadata                interface{}   `json:"thread_metadata,omitempty"`
	Member                        interface{}   `json:"member,omitempty"`
	DefaultAutoArchiveDuration    uint          `json:"default_auto_archive_duration,omitempty"`
	Permissions                   string        `json:"permissions,omitempty"`
	Flags                         uint          `json:"flags,omitempty"`
	TotalMessageSent              uint          `json:"total_message_sent,omitempty"`
	AvailableTags                 []interface{} `json:"available_tags,omitempty"`
	AppliedTags                   []string      `json:"applied_tags,omitempty"`
	DefaultReactionEmoji          interface{}   `json:"default_reaction_emoji,omitempty"`
	DefaultThreadRateLimitPerUser uint          `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              uint          `json:"default_sort_order,omitempty"`
	DefaultForumLayout            uint          `json:"default_forum_layout,omitempty"`
}

type Interaction struct {
	ID                           string                            `json:"id"`
	ApplicationID                string                            `json:"application_id"`
	Type                         InteractionType                   `json:"type"`
	Data                         InteractionApplicationCommandData `json:"data,omitempty"`
	GuildID                      string                            `json:"guild_id,omitempty"`
	Token                        string                            `json:"token"`
	Version                      uint                              `json:"version"`
	Context                      InteractionContextType            `json:"context,omitempty"`
	Guild                        interface{}                       `json:"guild,omitempty"`
	Channel                      Channel                           `json:"channel,omitempty"`
	Member                       Member                            `json:"member,omitempty"`
	User                         User                              `json:"user,omitempty"`
	Message                      interface{}                       `json:"message,omitempty"`
	AppPermissions               string                            `json:"app_permissions,omitempty"`
	Locale                       string                            `json:"locale,omitempty"`
	GuildLocale                  string                            `json:"guild_locale,omitempty"`
	Entitlements                 []interface{}                     `json:"entitlements,omitempty"`
	AuthorizingIntegrationOwners interface{}                       `json:"authorizing_integration_owners,omitempty"`
}

type InteractionResponseType = uint

const (
	InteractionResponseTypePong                                 InteractionResponseType = 1
	InteractionResponseTypeChannelMessageWithSource             InteractionResponseType = 4
	InteractionResponseTypeDeferredChannelMessageWithSource     InteractionResponseType = 5
	InteractionResponseTypeDeferredUpdateMessage                InteractionResponseType = 6
	InteractionResponseTypeUpdateMessage                        InteractionResponseType = 7
	InteractionResponseTypeApplicationCommandAutoCompleteResult InteractionResponseType = 8
	InteractionResponseTypeModal                                InteractionResponseType = 9
	InteractionResponseTypePremiumRequired                      InteractionResponseType = 10
	InteractionResponseTypeLaunchActivity                       InteractionResponseType = 12
)

type InteractionResponseDataMessage struct {
	Tts             bool        `json:"tts,omitempty"`
	Content         string      `json:"content,omitempty"`
	Flags           uint        `json:"flags,omitempty"`
	Embeds          interface{} `json:"embeds,omitempty"`
	AllowedMentions interface{} `json:"allowed_mentions,omitempty"`
	Components      interface{} `json:"components,omitempty"`
	Attachments     interface{} `json:"attachments,omitempty"`
	Poll            interface{} `json:"poll,omitempty"`
}

type InteractionResponse struct {
	Type InteractionResponseType        `json:"type"`
	Data InteractionResponseDataMessage `json:"data,omitempty"`
}
