package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/hendrywilliam/siren/internal/rest"
	"github.com/hendrywilliam/siren/internal/types"
	"github.com/hendrywilliam/siren/internal/voice"
	"github.com/hendrywilliam/siren/internal/voicemanager"
)

type GatewayIntent int

const (
	GuildsIntent GatewayIntent = 1 << (0 + iota)
	GuildMembersIntent
	GuildModerationIntent
	GuildExpressionIntent
	GuildIntegrationsIntent
	GuildWebhooksIntent
	GuildInvitesIntent
	GuildVoiceStatesIntent
	GuildPresencesIntent
	GuildMessagesIntent
	GuildMessageReactionIntent
	GuildMessageTypingIntent
	DirectMessageIntent
	DirectMessageReactionIntent
	DirectMessageTypingIntent
	MessageContentIntent
	GuildScheduledEventsIntent
	AutoModerationConfigurationIntent
	AutoModerationExecutionIntent
	GuildMessagePollsIntent
	DirectMessagePollsIntent
)

type GatewayStatus int

const (
	StatusReady GatewayStatus = iota + 1
	StatusDisconnected
)

const (
	OpcodeDispatch types.EventOpcode = iota
	OpcodeHeartbeat
	OpcodeIdentify
	OpcodePresenceUpdate
	OpcodeVoiceStateUpdate
	OpcodeResume
	OpcodeReconnect
	OpcodeRequestGuildMember
	OpcodeInvalidSession
	OpcodeHello
	OpcodeHeartbeatAck
	OpcodeRequestSoundboardSounds
)

type GatewayCloseEventCode int

const (
	CloseUnknownError GatewayCloseEventCode = 4000 + iota
	CloseUnknownOpcode
	CloseDecodeError
	CloseNotAuthenticated
	CloseAuthenticationFailed
	CloseAlreadyAuthenticated
	CloseInvalidSeq
	CloseRateLimited
	CloseSessionTimedOut
	CloseDisallowedIntents
)

var (
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")
	ErrNotAuthenticated     = fmt.Errorf("not authenticated")
	ErrDecode               = fmt.Errorf("decode error")
	ErrGatewayIsAlreadyOpen = fmt.Errorf("gateway is already open")
	ErrUnknown              = fmt.Errorf("unknown error")
	ErrDisallowedIntents    = fmt.Errorf("disallowed intents")
	ErrUnrecognizedEvent    = fmt.Errorf("unrecognized event")
)

// Using EventOpcode from types package instead
type HandlerFunc[T any] func(ctx context.Context, event T) error

type Discord struct {
	rwlock           sync.RWMutex
	wsURL            string
	resumeGatewayURL string
	sessionID        string
	wsConn           *websocket.Conn
	wsDialer         *websocket.Dialer
	sequence         atomic.Uint64
	ctx              context.Context
	heartbeatTicker  *time.Ticker
	status           GatewayStatus

	botToken           string
	botIntents         int
	botVersion         uint
	clientID           string
	discordHTTPBaseURL string

	voiceManager voicemanager.VoiceManager
	log          *slog.Logger

	message     *rest.MessageAPI
	interaction rest.Interaction
	voice       *rest.VoiceAPI

	handlers map[string][]any
}

const (
	EventMessageCreate     = "MESSAGE_CREATE"
	EventInteractionCreate = "INTERACTION_CREATE"
	EventVoiceStateUpdate  = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate = "VOICE_SERVER_UPDATE"
	EventReady             = "READY"
	EventGuildCreate       = "GUILD_CREATE"
	EventGuildMemberAdd    = "GUILD_MEMBER_ADD"
	EventGuildMemberRemove = "GUILD_MEMBER_REMOVE"
	EventMessageUpdate     = "MESSAGE_UPDATE"
	EventMessageDelete     = "MESSAGE_DELETE"
)

type DiscordArgs struct {
	BotToken   string
	BotIntent  []int
	BotVersion uint
	ClientID   string
	Logger     *slog.Logger
	Handlers   map[string][]any
}

func New(a DiscordArgs) *Discord {
	wsBaseURL := url.URL{
		Scheme:   "wss",
		Host:     "gateway.discord.gg",
		RawQuery: fmt.Sprintf("v=%d&encoding=json", a.BotVersion),
	}
	httpBaseURL := url.URL{
		Scheme: "https",
		Host:   "discord.com",
		Path:   fmt.Sprintf("api/v%d", a.BotVersion),
	}
	intents := 0
	for _, v := range a.BotIntent {
		intents |= v
	}
	restClient := rest.NewREST(httpBaseURL.String(), a.BotToken)

	return &Discord{
		clientID:           a.ClientID,
		wsDialer:           websocket.DefaultDialer,
		wsURL:              wsBaseURL.String(),
		botToken:           a.BotToken,
		botIntents:         intents,
		botVersion:         a.BotVersion,
		status:             StatusDisconnected,
		discordHTTPBaseURL: httpBaseURL.String(),
		voiceManager:       voicemanager.NewVoiceManager(),
		log:                a.Logger,
		handlers:           a.Handlers,
		message:            rest.NewMessageAPI(restClient),
		interaction:        rest.NewInteractionAPI(restClient),
		voice:              rest.NewVoiceAPI(restClient),
	}
}

func (g *Discord) Open(ctx context.Context) error {
	if g.status == StatusReady {
		return ErrGatewayIsAlreadyOpen
	}
	g.ctx = ctx
	return g.open()
}

func (g *Discord) open() error {
	conn, _, err := g.wsDialer.DialContext(g.ctx, g.wsURL, nil)
	if err != nil {
		return err
	}
	g.wsConn = conn

	messageType, rawMessage, err := g.wsConn.ReadMessage()
	if err != nil {
		return err
	}
	if messageType != websocket.TextMessage {
		return ErrDecode
	}
	helloEvent := &types.HelloEvent{}
	if err := json.NewDecoder(bytes.NewBuffer(rawMessage)).Decode(helloEvent); err != nil {
		return err
	}
	g.heartbeating(time.Duration(helloEvent.HeartbeatInterval) * time.Millisecond)

	if g.sessionID != "" {
		resumeEvent := &types.Event{
			Op: OpcodeResume,
			D: &types.ResumeEvent{
				Token:     g.botToken,
				SessionID: g.sessionID,
				Seq:       g.sequence.Load(),
			},
		}
		data, err := json.Marshal(resumeEvent)
		if err != nil {
			return err
		}
		if err := g.sendEvent(websocket.TextMessage, data); err != nil {
			return err
		}
	} else {
		identifyEvent := &types.Event{
			Op: OpcodeIdentify,
			D: &types.IdentifyEvent{
				Token: g.botToken,
				Properties: types.IdentifyEventProperties{
					Os:      "linux",
					Browser: "dcbot",
					Device:  "dcbot",
				},
				Intents: g.botIntents,
			},
		}
		data, err := json.Marshal(identifyEvent)
		if err != nil {
			return err
		}
		if err := g.sendEvent(websocket.TextMessage, data); err != nil {
			return err
		}
	}

	// Start listening for events
	go g.listen()

	return nil
}

func (g *Discord) retry() {
	if g.ctx.Err() != nil {
		return
	}

	backoff := 1 * time.Second
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-time.After(backoff):
			if err := g.open(); err != nil {
				g.log.Error("failed to reconnect", "error", err, "backoff", backoff)
				backoff *= 2
				if backoff > 5*time.Minute {
					backoff = 5 * time.Minute
				}
				continue
			}
			return
		}
	}
}

func (g *Discord) isSelf(id string) bool {
	return id == g.clientID
}

func (g *Discord) acceptEvent(messageType int, rawMessage []byte) (*types.RawEvent, error) {
	e := &types.RawEvent{}
	if err := json.NewDecoder(bytes.NewBuffer(rawMessage)).Decode(e); err != nil {
		return e, err
	}

	switch e.Op {
	case OpcodeHeartbeat:
		sequence := g.sequence.Load()
		data, _ := json.Marshal(types.Event{Op: OpcodeHeartbeat, D: sequence})
		g.sendEvent(websocket.BinaryMessage, data)
		return e, nil

	case OpcodeHeartbeatAck:
		g.log.Info("heartbeat acknowledged")
		return e, nil

	case OpcodeReconnect:
		g.status = StatusDisconnected
		if err := g.reconnect(); err != nil {
			return nil, err
		}
		return e, nil

	case OpcodeDispatch:
		if err := g.dispatch(e.T, e.D); err != nil {
			return nil, err
		}
		return e, nil

	default:
		return nil, ErrUnrecognizedEvent
	}
}

func (g *Discord) dispatch(eventName string, rawData []byte) error {
	g.sequence.Store(g.sequence.Load() + 1)

	// First call custom handlers for all events
	if err := g.callCustomHandlers(eventName, rawData); err != nil {
		return err
	}

	// Then handle certain events internally
	switch eventName {
	case EventMessageCreate:
		return g.handleMessageCreate(rawData)
	case EventInteractionCreate:
		return g.handleInteractionCreate(rawData)
	case EventVoiceStateUpdate:
		return g.handleVoiceStateUpdate(rawData)
	case EventVoiceServerUpdate:
		return g.handleVoiceServerUpdate(rawData)
	case EventReady:
		return g.handleReady(rawData)
	}

	return nil
}

func (g *Discord) callCustomHandlers(eventName string, rawData []byte) error {
	handlers, ok := g.handlers[eventName]
	if !ok || len(handlers) == 0 {
		g.log.Debug("no custom handlers registered for event", "event", eventName)
		return nil
	}

	g.log.Debug("calling custom handlers for event", "event", eventName, "handler_count", len(handlers))

	for _, handler := range handlers {
		handlerValue := reflect.ValueOf(handler)
		if handlerValue.Kind() != reflect.Func {
			g.log.Error("handler is not a function", "event", eventName)
			continue
		}

		ctx := context.Background()

		handlerType := handlerValue.Type()
		if handlerType.NumIn() != 2 {
			g.log.Error("handler must have exactly 2 parameters (context.Context and event data)",
				"event", eventName, "num_params", handlerType.NumIn())
			continue
		}

		eventType := handlerType.In(1)
		eventPtr := reflect.New(eventType)
		eventInterface := eventPtr.Interface()

		if err := json.Unmarshal(rawData, eventInterface); err != nil {
			g.log.Error("failed to unmarshal event data for handler",
				"event", eventName, "error", err)
			continue
		}

		eventValue := reflect.ValueOf(eventInterface).Elem()
		results := handlerValue.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			eventValue,
		})

		if len(results) > 0 && !results[0].IsNil() {
			if err, ok := results[0].Interface().(error); ok {
				g.log.Error("handler returned error",
					"event", eventName, "error", err)
				return err
			}
		}
	}

	return nil
}

func (g *Discord) RegisterHandler(eventName string, handler any) {
	g.rwlock.Lock()
	defer g.rwlock.Unlock()

	if g.handlers == nil {
		g.handlers = make(map[string][]any)
	}

	g.handlers[eventName] = append(g.handlers[eventName], handler)
	g.log.Info("registered handler for event", "event", eventName)
}

func (g *Discord) handleMessageCreate(rawData []byte) error {
	messageEvent := types.Message{}
	if err := json.Unmarshal(rawData, &messageEvent); err != nil {
		return err
	}
	if g.isSelf(messageEvent.Author.ID) {
		return nil
	}

	// Only send a response if there are no custom handlers for MESSAGE_CREATE
	// or if custom handlers exist but didn't return an error
	if handlers, exists := g.handlers[EventMessageCreate]; !exists || len(handlers) == 0 {
		_, err := g.message.CreateMessage(g.ctx, messageEvent.ChannelID, rest.CreateMessageOptions{
			Data: rest.CreateMessageData{
				Content: fmt.Sprintf("hello, %s", messageEvent.Author.Mention()),
			},
		})
		return err
	}

	return nil
}

func (g *Discord) handleInteractionCreate(rawData []byte) error {
	interactionEvent := types.Interaction{}
	if err := json.Unmarshal(rawData, &interactionEvent); err != nil {
		return err
	}

	userVoiceState, err := g.voice.GetUserVoiceState(g.ctx, interactionEvent.GuildID, interactionEvent.Member.User.ID)
	if err != nil {
		return err
	}

	if userVoiceState.SessionID == "" {
		_, err := g.interaction.Reply(g.ctx, interactionEvent.ID, interactionEvent.Token, rest.CreateInteractionResponseOptions{
			InteractionResponse: &types.InteractionResponse{
				Type: types.InteractionResponseTypeChannelMessageWithSource,
				Data: types.InteractionResponseDataMessage{
					Content: fmt.Sprintf("%s, please join a voice channel first.", interactionEvent.Member.User.Mention()),
				},
			},
		})
		return err
	}

	if _, err = g.interaction.Reply(g.ctx, interactionEvent.ID, interactionEvent.Token, rest.CreateInteractionResponseOptions{
		InteractionResponse: &types.InteractionResponse{
			Type: types.InteractionResponseTypeChannelMessageWithSource,
			Data: types.InteractionResponseDataMessage{
				Content: fmt.Sprintf("Playing 'Sirens' for %s", interactionEvent.Member.User.Mention()),
			},
		},
		WithResponse: false,
	}); err != nil {
		return err
	}

	voiceStateUpdate := &types.Event{
		Op: OpcodeVoiceStateUpdate,
		D: &types.VoiceStateUpdate{
			GuildID:   userVoiceState.GuildID,
			ChannelID: userVoiceState.ChannelID,
			SelfMute:  userVoiceState.SelfMute,
			SelfDeaf:  userVoiceState.SelfDeaf,
		},
	}
	data, err := json.Marshal(voiceStateUpdate)
	if err != nil {
		return err
	}
	return g.sendEvent(websocket.BinaryMessage, data)
}

func (g *Discord) handleVoiceStateUpdate(rawData []byte) error {
	voiceStateEvent := types.VoiceState{}
	if err := json.Unmarshal(rawData, &voiceStateEvent); err != nil {
		return err
	}
	g.log.Info("voice state update received", "guild_id", voiceStateEvent.GuildID)

	if !g.isSelf(voiceStateEvent.UserID) {
		return nil
	}

	newVoice := voice.NewVoice(voice.NewVoiceArguments{
		SessionID:  voiceStateEvent.SessionID,
		ServerID:   voiceStateEvent.GuildID,
		BotVersion: g.botVersion,
		UserID:     voiceStateEvent.UserID,
		Log:        g.log,
	})
	g.voiceManager.Add(voiceStateEvent.GuildID, newVoice)

	return nil
}

func (g *Discord) handleVoiceServerUpdate(rawData []byte) error {
	voiceServerEvent := types.VoiceServerUpdate{}
	if err := json.Unmarshal(rawData, &voiceServerEvent); err != nil {
		return err
	}
	g.log.Info("voice server update received", "guild_id", voiceServerEvent.GuildID)

	v := g.voiceManager.Get(voiceServerEvent.GuildID)
	if v == nil {
		g.log.Error("no voice instance found for guild", "guild_id", voiceServerEvent.GuildID)
		return nil
	}
	v.VoiceGatewayURL = voiceServerEvent.Endpoint
	v.Token = voiceServerEvent.Token
	v.Open(g.ctx)

	return nil
}

func (g *Discord) handleReady(rawData []byte) error {
	readyEvent := types.ReadyEvent{}
	if err := json.Unmarshal(rawData, &readyEvent); err != nil {
		return err
	}

	g.status = StatusReady
	g.resumeGatewayURL = readyEvent.ResumeGatewayURL
	g.sessionID = readyEvent.SessionID
	g.log.Info("connected to discord gateway", "session_id", g.sessionID)

	return nil
}

func (g *Discord) reconnect() error {
	if g.wsConn != nil {
		g.wsConn.Close()
		g.wsConn = nil
	}

	if g.sessionID != "" {
		return g.open()
	}

	// If we don't have a sessionID, connect from scratch
	g.sessionID = ""
	return g.open()
}

func (g *Discord) listen() {
	for {
		select {
		case <-g.ctx.Done():
			return
		default:
			messageType, rawMessage, err := g.wsConn.ReadMessage()
			if err != nil {
				g.log.Error("error reading message", "error", err)
				g.status = StatusDisconnected
				go g.retry()
				return
			}

			if _, err := g.acceptEvent(messageType, rawMessage); err != nil {
				g.log.Error("error processing event", "error", err)
			}
		}
	}
}

func (g *Discord) heartbeating(interval time.Duration) {
	g.heartbeatTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-g.ctx.Done():
				g.heartbeatTicker.Stop()
				return
			case <-g.heartbeatTicker.C:
				sequence := g.sequence.Load()
				data, _ := json.Marshal(types.Event{Op: OpcodeHeartbeat, D: sequence})
				if err := g.sendEvent(websocket.TextMessage, data); err != nil {
					g.log.Error("failed to send heartbeat", "error", err)
				}
			}
		}
	}()
}

func (g *Discord) Close() error {
	if g.wsConn != nil {
		g.wsConn.Close()
		g.wsConn = nil
	}
	g.status = StatusDisconnected
	return nil
}

func (g *Discord) sendEvent(messageType int, data []byte) error {
	if g.wsConn == nil {
		return ErrNotAuthenticated
	}
	g.rwlock.RLock()
	defer g.rwlock.RUnlock()
	return g.wsConn.WriteMessage(messageType, data)
}
