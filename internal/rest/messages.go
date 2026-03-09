package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/hendrywilliam/siren/internal/types"
)

type MessageAPI struct {
	rest RESTClient
}

func NewMessageAPI(rest RESTClient) *MessageAPI {
	return &MessageAPI{rest: rest}
}

func (m *MessageAPI) createMessageRoute(channelID string) (string, error) {
	u, err := url.Parse(m.rest.URL())
	if err != nil {
		return "", err
	}
	cmPath := fmt.Sprintf("/channels/%s/messages", channelID)
	actualPath, err := url.JoinPath(u.Path, cmPath)
	if err != nil {
		return "", err
	}
	cmURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   actualPath,
	}
	return cmURL.String(), nil
}

type CreateMessageData struct {
	Content          string `json:"content"`
	Tts              bool   `json:"tts"`
	Nonce            any    `json:"nonce,omitempty"`
	Embeds           any    `json:"embeds,omitempty"`
	AllowedMentions  any    `json:"allowed_mentions,omitempty"`
	MessageReference any    `json:"message_reference,omitempty"`
	Components       any    `json:"components,omitempty"`
	StickerIDS       any    `json:"sticker_ids,omitempty"`
}

type CreateMessageOptions struct {
	Data CreateMessageData
}

func (m *MessageAPI) CreateMessage(ctx context.Context, channelID string, options CreateMessageOptions) (*types.Message, error) {
	cmURL, err := m.createMessageRoute(channelID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(options.Data); err != nil {
		return nil, err
	}
	res, err := m.rest.Post(ctx, cmURL, buf, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	msg := &types.Message{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
