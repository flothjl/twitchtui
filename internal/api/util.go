package api

import (
	"encoding/json"
	"net/http"
)

type ResponseData[T any] struct {
	Total  int `json:"total,omitempty"`  // Only present in some endpoints.
	Points int `json:"points,omitempty"` // Only present in some endpoints.

	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination,omitempty"`

	Status  int    `json:"status"`            // If not provided by Twitch, defaults to HTTP status code.
	Code    string `json:"error"`             // If not provided by Twitch, defaults to HTTP status text.
	Message string `json:"message,omitempty"` // Only present if status is non-200
}

type Pagination struct {
	Cursor string `json:"cursor,omitempty"`
}

func decodeResponse[T any](res *http.Response) (*ResponseData[T], error) {
	var data ResponseData[T]
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
