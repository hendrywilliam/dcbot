package types

import "time"

type VoiceState struct {
	GuildID                 string    `json:"guild_id"`
	ChannelID               string    `json:"channel_id"`
	UserID                  string    `json:"user_id"`
	Member                  Member    `json:"member,omitempty"`
	SessionID               string    `json:"session_id"`
	Deaf                    bool      `json:"deaf"`
	Mute                    bool      `json:"mute"`
	SelfDeaf                bool      `json:"self_deaf"`
	SelfMute                bool      `json:"self_mute"`
	SelfStream              bool      `json:"self_stream"`
	SelfVideo               bool      `json:"self_video"`
	Suppress                bool      `json:"suppress"`
	RequestToSpeakTimestamp time.Time `json:"request_to_speak_timestamp"`
}

type VoiceStateUpdate struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	SelfMute  bool   `json:"self_mute"`
	SelfDeaf  bool   `json:"self_deaf"`
}

type VoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type VoiceHello struct {
	HeartbeatInterval uint `json:"heartbeat_interval"`
}

type VoiceHeartbeat struct {
	T      int64  `json:"t"`
	Opcode int    `json:"op"`
	Event  string `json:"e"`
}

type VoiceIdentify struct {
	ServerID  string `json:"server_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
}

type VoiceReadyStream struct {
	Active  bool   `json:"active"`
	Quality int    `json:"quality"`
	RtcSsrc uint32 `json:"rtc_ssrc"`
	Ssrc    uint32 `json:"ssrc"`
	Type    string `json:"type"`
}

type VoiceReady struct {
	Experiments []string           `json:"experiments"`
	Heartbeat   int                `json:"heartbeat_interval"`
	IP          string             `json:"ip"`
	Modes       []string           `json:"modes"`
	Port        int                `json:"port"`
	Ssrc        uint32             `json:"ssrc"`
	Streams     []VoiceReadyStream `json:"streams"`
}

type VoiceIPDiscovery struct {
	Type    uint16
	Length  uint16
	Ssrc    uint32
	Address [16]byte
	Port    uint16
}

type SelectProtocolData struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Port    int    `json:"port"`
}

type SelectProtocol struct {
	Protocol string             `json:"protocol"`
	Data     SelectProtocolData `json:"data"`
}

type SessionDescription struct {
	AudioCodec          string   `json:"audio_codec"`
	MediaSessionID      string   `json:"media_session_id"`
	Mode                string   `json:"mode"`
	SecretKey           [32]byte `json:"secret_key"`
	VideoCodec          string   `json:"video_codec"`
	RtxSsrc             uint32   `json:"rtx_ssrc"`
	AudioSsrc           uint32   `json:"audio_ssrc"`
	Streams             []any    `json:"streams"`
	DisableUdpDiscovery bool     `json:"disable_udp_discovery"`
}

type ClientsConnect struct {
	UsersID []string `json:"user_ids"`
	Video   bool     `json:"video"`
	Audio   bool     `json:"audio"`
}

type Speaking struct {
	Speaking int    `json:"speaking"`
	Delay    int    `json:"delay"`
	Ssrc     uint32 `json:"ssrc"`
	UserID   string `json:"user_id"`
}

type VoiceResume struct {
	ServerID            string `json:"server_id"`
	SessionID           string `json:"session_id"`
	Token               string `json:"token"`
	ResumeGatewayURL    string `json:"resume_gateway_url"`
	ResumeSessionID     string `json:"resume_session_id"`
	ResumeSequence      uint64 `json:"resume_sequence"`
	ResumeGatewayStatus string `json:"resume_gateway_status"`
}
