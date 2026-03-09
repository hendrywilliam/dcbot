package voice

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hendrywilliam/siren/internal/audio"
	"github.com/hendrywilliam/siren/internal/audiosender"
	"github.com/hendrywilliam/siren/internal/types"
)

// VoiceGatewayStatus represents the current state of the voice WebSocket connection.
type VoiceGatewayStatus = string

const (
	StatusReady        VoiceGatewayStatus = "READY"
	StatusDisconnected VoiceGatewayStatus = "DISCONNECTED"
)

// Speaking mode flags.
var (
	SpeakingModeMicrophone = 1 << 0
	SpeakingModeSoundshare = 1 << 1
	SpeakingModePriority   = 1 << 2
)

// VoiceOpcode represents the op codes used by the Discord Voice Gateway.
type VoiceOpcode = int

const (
	OpcodeIdentify           VoiceOpcode = 0
	OpcodeSelectProtocol     VoiceOpcode = 1
	OpcodeReady              VoiceOpcode = 2
	OpcodeHeartbeat          VoiceOpcode = 3
	OpcodeSessionDescription VoiceOpcode = 4
	OpcodeSpeaking           VoiceOpcode = 5
	OpcodeHeartbeatAck       VoiceOpcode = 6
	OpcodeResume             VoiceOpcode = 7
	OpcodeHello              VoiceOpcode = 8
	OpcodeResumed            VoiceOpcode = 9
	OpcodeClientsConnect     VoiceOpcode = 11
	OpcodeClientDisconnect   VoiceOpcode = 13

	// DAVE (Discord's E2EE protocol) opcodes.
	DAVEPrepareTransition        VoiceOpcode = 21
	DAVEExecuteTransition        VoiceOpcode = 22
	DAVETransitionReady          VoiceOpcode = 23
	DAVEPrepareEpoch             VoiceOpcode = 24
	DAVEMLSExternalSender        VoiceOpcode = 25
	DAVEMLSKeyPackage            VoiceOpcode = 26
	DAVEMLSProposals             VoiceOpcode = 27
	DAVECommitWelcome            VoiceOpcode = 28
	DAVEAnnounceCommitTransition VoiceOpcode = 29
	DAVEMLSWelcome               VoiceOpcode = 30
	DAVEMLSInvalidCommitWelcome  VoiceOpcode = 31
)

// VoiceCloseCode represents close event codes for the voice WebSocket connection.
type VoiceCloseCode = int

const (
	VoiceCloseUnknownOpcode        VoiceCloseCode = 4001
	VoiceCloseFailedToDecode       VoiceCloseCode = 4002
	VoiceCloseNotAuthenticated     VoiceCloseCode = 4003
	VoiceCloseAuthenticationFailed VoiceCloseCode = 4004
	VoiceCloseAlreadyAuthenticated VoiceCloseCode = 4005
	VoiceCloseSessionInvalid       VoiceCloseCode = 4006
	VoiceCloseSessionTimeout       VoiceCloseCode = 4009
	VoiceCloseServerNotFound       VoiceCloseCode = 4011
	VoiceCloseUnknownProtocol      VoiceCloseCode = 4012
	VoiceCloseDisconnected         VoiceCloseCode = 4014
	VoiceCloseServerCrashed        VoiceCloseCode = 4015
	VoiceCloseUnknownEncryption    VoiceCloseCode = 4016
)

var ErrUnrecognizedEvent = errors.New("unrecognized event")

// Voice manages a single voice gateway connection for one guild.
type Voice struct {
	rwlock     sync.RWMutex
	wsDialer   *websocket.Dialer
	wsConn     *websocket.Conn
	log        *slog.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc

	status     VoiceGatewayStatus
	botVersion uint
	sequence   atomic.Uint64

	heartbeatTicker *time.Ticker

	udpConn        *net.UDPConn
	port           uint16
	ssrc           uint32
	ip             string
	encryptionMode string

	// Voice session identifiers.
	SessionID       string
	ServerID        string // Guild ID
	UserID          string
	VoiceGatewayURL string
	Token           string

	// Audio pipeline.
	secretKeys      [32]byte
	audio           *audio.Audio
	audioSender     *audiosender.AudioSender
	audioCtx        context.Context
	audioCancelFunc context.CancelFunc
	audioDataChan   chan []byte
	audioIsFinished chan bool
}

// NewVoiceArguments holds the parameters required to create a new Voice instance.
type NewVoiceArguments struct {
	SessionID  string
	BotVersion uint
	ServerID   string
	UserID     string
	Log        *slog.Logger
}

// NewVoice creates and returns an initialised Voice instance.
func NewVoice(args NewVoiceArguments) *Voice {
	return &Voice{
		wsDialer:        websocket.DefaultDialer,
		status:          StatusDisconnected,
		log:             args.Log.With("voice_id", fmt.Sprintf("voice_%s", args.SessionID)),
		botVersion:      args.BotVersion,
		SessionID:       args.SessionID,
		UserID:          args.UserID,
		ServerID:        args.ServerID,
		audio:           &audio.Audio{},
		audioSender:     &audiosender.AudioSender{},
		audioDataChan:   make(chan []byte),
		audioIsFinished: make(chan bool, 1),
	}
}

// Open starts the voice gateway connection.
func (v *Voice) Open(ctx context.Context) error {
	return v.open(ctx)
}

func (v *Voice) open(ctx context.Context) error {
	v.ctx, v.cancelFunc = context.WithCancel(ctx)

	wsURL := url.URL{
		Scheme:   "wss",
		Host:     v.VoiceGatewayURL,
		RawQuery: fmt.Sprintf("v=%d", v.botVersion),
	}

	var err error
	v.wsConn, _, err = v.wsDialer.DialContext(v.ctx, wsURL.String(), nil)
	if err != nil {
		v.log.Error("failed to dial voice gateway", "error", err)
		return err
	}

	identifyEvent := &types.Event{
		Op: OpcodeIdentify,
		D: types.VoiceIdentify{
			ServerID:  v.ServerID,
			UserID:    v.UserID,
			SessionID: v.SessionID,
			Token:     v.Token,
		},
	}
	data, err := json.Marshal(identifyEvent)
	if err != nil {
		return err
	}
	if err = v.sendEvent(websocket.TextMessage, data); err != nil {
		v.log.Error("failed to send identify event", "error", err)
		return err
	}
	v.log.Info("identify event sent")

	// First response should be Hello.
	e := &types.RawEvent{}
	if err = v.wsConn.ReadJSON(e); err != nil {
		v.log.Error("failed to read hello event", "error", err)
		return err
	}
	v.log.Info("event received", "op", e.Op)

	if e.Op == OpcodeHello {
		d := &types.VoiceHello{}
		if err := json.Unmarshal(e.D, d); err != nil {
			return err
		}
		go v.heartbeating(time.Duration(d.HeartbeatInterval))
	}

	// Second response should be Ready.
	e = &types.RawEvent{}
	if err = v.wsConn.ReadJSON(e); err != nil {
		return err
	}
	v.log.Info("event received", "op", e.Op)

	if e.Op == OpcodeReady {
		readyEvent := &types.VoiceReady{}
		if err := json.Unmarshal(e.D, readyEvent); err != nil {
			return err
		}
		v.status = StatusReady
		v.ip = readyEvent.IP
		v.port = readyEvent.Port
		v.encryptionMode = "aead_xchacha20_poly1305_rtpsize"
		v.ssrc = readyEvent.SSRC

		go v.listen(v.wsConn)

		if err = v.dialUDP(readyEvent.IP, readyEvent.Port); err != nil {
			v.log.Error("failed to dial UDP", "error", err)
			return err
		}
	}
	return nil
}

func (v *Voice) listen(conn *websocket.Conn) {
	for {
		select {
		case <-v.ctx.Done():
			return
		default:
			v.rwlock.Lock()
			same := v.wsConn == conn
			v.rwlock.Unlock()
			if !same {
				// A newer connection has been opened; this goroutine should exit.
				return
			}
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				v.log.Error("voice listen error", "error", err)
				// @todo: implement reconnect instead of panic
				panic(err)
			}
			if _, err = v.acceptEvent(messageType, message); err != nil {
				v.log.Error("failed to handle voice event", "error", err)
			}
		}
	}
}

func (v *Voice) acceptEvent(messageType int, rawMessage []byte) (*types.RawEvent, error) {
	e := &types.RawEvent{}
	if err := json.NewDecoder(bytes.NewBuffer(rawMessage)).Decode(e); err != nil {
		return e, err
	}

	switch e.Op {
	case OpcodeHeartbeatAck:
		v.log.Info("heartbeat acknowledged")
		return e, nil

	case OpcodeSessionDescription:
		v.log.Info("session description received")

		sessionDesc := &types.SessionDescription{}
		if err := json.Unmarshal(e.D, sessionDesc); err != nil {
			return nil, err
		}
		v.secretKeys = sessionDesc.SecretKey
		v.encryptionMode = sessionDesc.Mode

		speakingEvent := &types.Event{
			Op: OpcodeSpeaking,
			D: &types.Speaking{
				Speaking: SpeakingModeMicrophone,
				Delay:    0,
				SSRC:     v.ssrc,
			},
		}
		data, err := json.Marshal(speakingEvent)
		if err != nil {
			return nil, err
		}
		if err = v.sendEvent(websocket.BinaryMessage, data); err != nil {
			return nil, err
		}
		v.log.Info("speaking event sent")

		v.audioCtx, v.audioCancelFunc = context.WithCancel(v.ctx)
		go v.audio.Encode(v.audioCtx, "sirens.mp3", v.audioDataChan, v.audioIsFinished)
		go v.audioSender.Send(v.audioCtx, v.udpConn, v.ssrc, v.secretKeys, v.audioDataChan, v.audioIsFinished)

		return e, nil

	default:
		v.log.Info("unhandled voice event", "op", e.Op)
		return e, nil
	}
}

// Close stops heartbeating, cancels the context, and closes the WebSocket.
func (v *Voice) Close() {
	if v.heartbeatTicker != nil {
		v.heartbeatTicker.Stop()
		v.heartbeatTicker = nil
	}
	if v.audioCancelFunc != nil {
		v.audioCancelFunc()
	}
	v.status = StatusDisconnected
	v.cancelFunc()
	v.wsConn.Close()
	v.log.Info("voice connection closed")
}

func (v *Voice) heartbeating(dur time.Duration) error {
	v.heartbeatTicker = time.NewTicker(dur * time.Millisecond)
	for {
		select {
		case <-v.ctx.Done():
			v.heartbeatTicker.Stop()
			v.log.Info("heartbeating stopped")
			return nil
		case <-v.heartbeatTicker.C:
			seq := v.sequence.Load()
			heartbeatEvent := &types.Event{
				Op: OpcodeHeartbeat,
				D: &types.VoiceHeartbeat{
					T:      v.nonce(),
					SeqAck: seq,
				},
			}
			data, err := json.Marshal(heartbeatEvent)
			if err != nil {
				return err
			}
			if err = v.sendEvent(websocket.BinaryMessage, data); err != nil {
				v.log.Error("failed to send heartbeat", "error", err)
				return err
			}
			v.log.Info("heartbeat event sent")
		}
	}
}

func (v *Voice) nonce() int64 {
	return time.Now().UnixMilli()
}

func (v *Voice) sendEvent(messageType int, data []byte) error {
	v.rwlock.Lock()
	defer v.rwlock.Unlock()
	return v.wsConn.WriteMessage(messageType, data)
}

func (v *Voice) sendIPDiscovery() error {
	var packet []byte
	packet = binary.BigEndian.AppendUint16(packet, 0x1)
	packet = binary.BigEndian.AppendUint16(packet, 70)
	packet = binary.BigEndian.AppendUint32(packet, v.ssrc)
	var address [64]byte
	copy(address[:], v.ip)
	packet = append(packet, address[:]...)
	packet = binary.BigEndian.AppendUint16(packet, v.port)

	if _, err := v.udpConn.Write(packet); err != nil {
		return err
	}

	b := make([]byte, 100)
	if _, err := v.udpConn.Read(b); err != nil {
		return err
	}

	// Extract the external IP address from the response (bytes 8–71).
	var ipBuilder strings.Builder
	for i := 8; i < 72; i++ {
		if b[i] == 0 {
			break
		}
		ipBuilder.WriteByte(b[i])
	}
	externalIP := ipBuilder.String()
	externalPort := binary.BigEndian.Uint16(b[72:74])

	v.log.Info("IP discovery complete", "ip", externalIP, "port", externalPort)
	return v.sendSelectProtocol(externalIP, externalPort)
}

func (v *Voice) sendSelectProtocol(ipAddr string, port uint16) error {
	e := &types.Event{
		Op: OpcodeSelectProtocol,
		D: &types.SelectProtocol{
			Protocol: "udp",
			Data: types.SelectProtocolData{
				Address: ipAddr,
				Port:    port,
				Mode:    v.encryptionMode,
			},
		},
	}
	data, err := json.Marshal(e)
	if err != nil {
		v.log.Error("failed to marshal select protocol event", "error", err)
		return err
	}
	if err = v.sendEvent(websocket.BinaryMessage, data); err != nil {
		v.log.Error("failed to send select protocol event", "error", err)
		return err
	}
	v.log.Info("select protocol event sent")
	return nil
}

func (v *Voice) dialUDP(ip string, port uint16) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		v.log.Error("failed to resolve UDP address", "error", err)
		return err
	}
	v.udpConn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		v.log.Error("failed to dial UDP", "error", err)
		return err
	}
	if err = v.sendIPDiscovery(); err != nil {
		v.log.Error("IP discovery failed", "error", err)
		return err
	}
	return nil
}
