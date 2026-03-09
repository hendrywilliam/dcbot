package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/hendrywilliam/siren/internal/types"
)

type VoiceAPI struct {
	rest RESTClient
}

func NewVoiceAPI(rest RESTClient) *VoiceAPI {
	return &VoiceAPI{rest: rest}
}

func (v *VoiceAPI) listVoiceRegionsRoute() (string, error) {
	return url.JoinPath(v.rest.URL(), "/voice/regions")
}

func (v *VoiceAPI) getCurrentUserVoiceStateRoute(guildID string) (string, error) {
	return url.JoinPath(v.rest.URL(), fmt.Sprintf("/guilds/%s/voice-states/@me", guildID))
}

func (v *VoiceAPI) getUserVoiceStateRoute(guildID, userID string) (string, error) {
	return url.JoinPath(v.rest.URL(), fmt.Sprintf("/guilds/%s/voice-states/%s", guildID, userID))
}

func (v *VoiceAPI) GetCurrentUserVoiceState(ctx context.Context, guildID string) (*types.VoiceState, error) {
	voiceStateURL, err := v.getCurrentUserVoiceStateRoute(guildID)
	if err != nil {
		return nil, err
	}
	res, err := v.rest.Get(ctx, voiceStateURL, nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	userVoiceState := &types.VoiceState{}
	if err := json.Unmarshal(data, userVoiceState); err != nil {
		return nil, err
	}
	return userVoiceState, nil
}

func (v *VoiceAPI) GetUserVoiceState(ctx context.Context, guildID, userID string) (*types.VoiceState, error) {
	voiceStateURL, err := v.getUserVoiceStateRoute(guildID, userID)
	if err != nil {
		return nil, err
	}
	res, err := v.rest.Get(ctx, voiceStateURL, nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	userVoiceState := &types.VoiceState{}
	if err := json.Unmarshal(data, userVoiceState); err != nil {
		return nil, err
	}
	return userVoiceState, nil
}
