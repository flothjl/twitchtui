package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Stream struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	UserLogin       string    `json:"user_login"`
	UserDisplayName string    `json:"user_name"`
	GameID          string    `json:"game_id"`
	GameName        string    `json:"game_name"`
	Type            string    `json:"type"`
	Title           string    `json:"title"`
	Tags            []string  `json:"tags"`
	ViewerCount     int       `json:"viewer_count"`
	Language        string    `json:"language"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	IsMature        bool      `json:"is_mature"`
	StartedAt       time.Time `json:"started_at"`
}

func (api *Api) GetFollowedStreams() (*ResponseData[Stream], error) {
	opt := AddQueryParam("user_id", api.clientToken.UserID)
	recordCount := AddQueryParam("first", "40")

	resp, err := api.doRequest(context.TODO(), http.MethodGet, "streams/followed", nil, opt, recordCount)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status code: %d", resp.StatusCode)
	}
	data, err := decodeResponse[Stream](resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}
