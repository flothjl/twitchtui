package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	chat "github.com/gempir/go-twitch-irc/v4"
)

type Api struct {
	httpClient   *http.Client
	apiUrl       string
	tokenPath    string
	authUrl      string
	clientToken  *AccessToken
	clientID     string
	clientSecret string
	chatClient   *chat.Client
}

type (
	Option        func(*Api)
	httpMethod    string
	RequestOption func(*http.Request)
)

const (
	baseAuthUrl       = "https://id.twitch.tv/oauth2"
	relativeTokenPath = "/.twitchtui/.token"
	baseAPIUrl        = "https://api.twitch.tv/helix"
)

func AddQueryParam(key, value string) RequestOption {
	return func(req *http.Request) {
		query := req.URL.Query()
		query.Add(key, value)
		req.URL.RawQuery = query.Encode()
	}
}

func New(clientID string, clientSecret string, options ...Option) (*Api, error) {
	// TODO: refactor so taht authorization can support tokenpath override

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fullFilepath := filepath.Join(homeDir, relativeTokenPath)
	api := &Api{
		httpClient:   &http.Client{},
		apiUrl:       baseAPIUrl,
		tokenPath:    fullFilepath,
		authUrl:      baseAuthUrl,
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	for _, opt := range options {
		opt(api)
	}

	err = api.authorizeUser()
	if err != nil {
		return api, err
	}

	api.chatClient = chat.NewClient(api.clientToken.Login, fmt.Sprintf("oauth:%s", api.clientToken.Token))

	return api, err
}

func (api *Api) doRequest(ctx context.Context, method string, path string, body io.Reader, reqOpts ...RequestOption) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", api.apiUrl, strings.TrimPrefix(path, "/"))
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-ID", api.clientToken.ClientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.clientToken.Token))

	for _, opt := range reqOpts {
		opt(req)
	}

	return api.httpClient.Do(req)
}
