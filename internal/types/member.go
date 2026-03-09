package types

type GuildMemberFlag = int

const (
	GuildMemberFlagDidRejoin                GuildMemberFlag = 1 << 0
	GuildMemberCompletedOnboarding          GuildMemberFlag = 1 << 1
	GuildMemberByPassesVerification         GuildMemberFlag = 1 << 2
	GuildMemberStartedOnboarding            GuildMemberFlag = 1 << 3
	GuildMemberIsGuest                      GuildMemberFlag = 1 << 4
	GuildMemberStartedHomeActions           GuildMemberFlag = 1 << 5
	GuildMemberCompletedHomeActions         GuildMemberFlag = 1 << 6
	GuildMemberAutomodQuarantinedUsername   GuildMemberFlag = 1 << 7
	GuildMemberDMSettingsUpsellAcknowledged GuildMemberFlag = 1 << 9
)

type Member struct {
	User                   User            `json:"user"`
	UnusualDmActivityUntil interface{}     `json:"unusual_dm_activity_until"`
	Nick                   string          `json:"nick,omitempty"`
	Avatar                 string          `json:"avatar,omitempty"`
	Banner                 string          `json:"banner,omitempty"`
	Roles                  []string        `json:"roles"`
	Deaf                   bool            `json:"deaf"`
	Mute                   bool            `json:"mute"`
	Flags                  GuildMemberFlag `json:"flags"`
	Pending                bool            `json:"pending,omitempty"`
	Permissions            string          `json:"permissions"`
}
