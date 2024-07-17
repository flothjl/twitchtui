package api

import (
	"sync"

	chat "github.com/gempir/go-twitch-irc/v4"
)

type ChatMessage struct {
	Username string
	Message  string
}

func (api *Api) InitChat(streamer string, messages *[]ChatMessage, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	api.chatClient.OnPrivateMessage(func(message chat.PrivateMessage) {
		*messages = append(*messages, ChatMessage{message.User.DisplayName, message.Message})
	})

	api.chatClient.OnSelfJoinMessage(func(message chat.UserJoinMessage) {
		*messages = append(*messages, ChatMessage{"JOINED", message.Raw})
	})

	api.chatClient.Join(streamer)

	go func() {
		err := api.chatClient.Connect()
		if err != nil {
			errChan <- err
		}
	}()
}

func (api *Api) DisconnectChat() error {
	return api.chatClient.Disconnect()
}
