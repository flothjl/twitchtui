package app

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flothjl/twitchtui/internal/api"
)

type chatTickMsg time.Time

type streamChatModel struct {
	messages    *[]api.ChatMessage
	isConnected bool
}

func doChatTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return chatTickMsg(t)
	})
}

func (m streamChatModel) view() string {
	s := ""
	for _, m := range *m.messages {
		s += fmt.Sprintf("%s  %s\n", m.Username, m.Message)
	}
	return s
}

func (m streamChatModel) Init() tea.Cmd {
	return doChatTick()
}

func (m streamChatModel) Update(msg tea.Msg) (streamChatModel, tea.Cmd) {
	switch msg.(type) {
	case chatTickMsg:
		return m, doChatTick()
	}
	return m, nil
}

func initChatCmd(m model, userLogin string) tea.Cmd {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	wg.Add(1)
	go m.apiClient.InitChat(userLogin, m.chat.messages, &wg, errChan)
	go func() {
		wg.Wait()
		close(errChan)
	}()
	return handleChatErrors(errChan)
}

func handleChatErrors(errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		for err := range errChan {
			if err != nil {
				return errMsg{err}
			}
		}
		return chatConnectedMsg{}
	}
}
