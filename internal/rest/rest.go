package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type REST struct {
	httpBaseURL string
	httpClient  *http.Client
	botToken    string
}

type RESTClient interface {
	URL() string
	Get(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error)
	Put(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error)
	Patch(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error)
	Delete(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error)
	Post(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error)
}

type RESTOptions struct {
	Headers map[string]string
}

func NewREST(baseURL, botToken string) *REST {
	return &REST{
		httpBaseURL: baseURL,
		httpClient:  http.DefaultClient,
		botToken:    botToken,
	}
}

func (r *REST) applyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func (r *REST) makeRequest(ctx context.Context, method, url string, body io.Reader, options *RESTOptions) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", r.botToken))
	if options != nil {
		r.applyHeaders(req, options.Headers)
	}
	return req, nil
}

func (r *REST) URL() string {
	return r.httpBaseURL
}

func (r *REST) Get(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error) {
	req, err := r.makeRequest(ctx, http.MethodGet, url, body, options)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(req)
}

func (r *REST) Put(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error) {
	req, err := r.makeRequest(ctx, http.MethodPut, url, body, options)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(req)
}

func (r *REST) Patch(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error) {
	req, err := r.makeRequest(ctx, http.MethodPatch, url, body, options)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(req)
}

func (r *REST) Delete(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error) {
	req, err := r.makeRequest(ctx, http.MethodDelete, url, body, options)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(req)
}

func (r *REST) Post(ctx context.Context, url string, body io.Reader, options *RESTOptions) (*http.Response, error) {
	req, err := r.makeRequest(ctx, http.MethodPost, url, body, options)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(req)
}
