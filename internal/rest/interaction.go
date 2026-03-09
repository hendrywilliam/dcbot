package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hendrywilliam/siren/internal/types"
)

type interaction struct {
	rest RESTClient
}

type Interaction interface {
	Reply(ctx context.Context, interactionID, interactionToken string, options CreateInteractionResponseOptions) (*http.Response, error)
	EditOriginal(ctx context.Context, applicationID, interactionToken string, options EditOriginalOptions) (*http.Response, error)
	GetOriginal(ctx context.Context, applicationID, interactionToken string, options GetOriginalOptions) (*http.Response, error)
	DeleteOriginal(ctx context.Context, applicationID, interactionToken string) (*http.Response, error)
}

func NewInteractionAPI(rest RESTClient) Interaction {
	return &interaction{rest: rest}
}

func (i *interaction) interactionResponseCallbackRoute(interactionID, interactionToken string, withResponse bool) (string, error) {
	u, err := url.Parse(i.rest.URL())
	if err != nil {
		return "", err
	}
	cbPath := fmt.Sprintf("/interactions/%s/%s/callback", interactionID, interactionToken)
	actualPath, err := url.JoinPath(u.Path, cbPath)
	if err != nil {
		return "", err
	}
	cbURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   actualPath,
	}
	queries := u.Query()
	if withResponse {
		queries.Add("with_response", "true")
	}
	cbURL.RawQuery = queries.Encode()
	return cbURL.String(), nil
}

func (i *interaction) originalInteractionRoute(applicationID, interactionToken, threadID string, withComponents bool) (string, error) {
	u, err := url.Parse(i.rest.URL())
	if err != nil {
		return "", err
	}
	orgPath := fmt.Sprintf("/webhooks/%s/%s/messages/@original", applicationID, interactionToken)
	actualPath, err := url.JoinPath(u.Path, orgPath)
	if err != nil {
		return "", err
	}
	orgURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   actualPath,
	}
	q := u.Query()
	if threadID != "" {
		q.Add("thread_id", threadID)
	}
	if withComponents {
		q.Add("with_components", "true")
	}
	orgURL.RawQuery = q.Encode()
	return orgURL.String(), nil
}

type CreateInteractionResponseOptions struct {
	InteractionResponse *types.InteractionResponse
	WithResponse        bool
}

func (i *interaction) Reply(ctx context.Context, interactionID, interactionToken string, options CreateInteractionResponseOptions) (*http.Response, error) {
	cbURL, err := i.interactionResponseCallbackRoute(interactionID, interactionToken, options.WithResponse)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(options.InteractionResponse); err != nil {
		return nil, err
	}
	return i.rest.Post(ctx, cbURL, buf, nil)
}

type EditOriginalData struct {
	Content         string `json:"content"`
	Flags           int32  `json:"flags"`
	PayloadJson     string `json:"payload_json"`
	Attachments     any    `json:"attachments"`
	Poll            any    `json:"poll"`
	Embed           any    `json:"embed"`
	AllowedMentions any    `json:"allowed_mentions"`
	Components      any    `json:"components"`
	Files           any    `json:"files"`
}

type EditOriginalOptions struct {
	Data           EditOriginalData
	ThreadID       string
	WithComponents bool
}

func (i *interaction) EditOriginal(ctx context.Context, applicationID, interactionToken string, options EditOriginalOptions) (*http.Response, error) {
	orgURL, err := i.originalInteractionRoute(applicationID, interactionToken, options.ThreadID, options.WithComponents)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(options.Data); err != nil {
		return nil, err
	}
	return i.rest.Patch(ctx, orgURL, buf, nil)
}

type GetOriginalOptions struct {
	ThreadID string
}

func (i *interaction) GetOriginal(ctx context.Context, applicationID, interactionToken string, options GetOriginalOptions) (*http.Response, error) {
	orgURL, err := i.originalInteractionRoute(applicationID, interactionToken, options.ThreadID, false)
	if err != nil {
		return nil, err
	}
	return i.rest.Get(ctx, orgURL, nil, nil)
}

func (i *interaction) DeleteOriginal(ctx context.Context, applicationID, interactionToken string) (*http.Response, error) {
	orgURL, err := i.originalInteractionRoute(applicationID, interactionToken, "", false)
	if err != nil {
		return nil, err
	}
	return i.rest.Delete(ctx, orgURL, nil, nil)
}
