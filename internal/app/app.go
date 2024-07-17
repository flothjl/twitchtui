package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flothjl/twitchtui/internal/api"
	"github.com/flothjl/twitchtui/internal/utils"
)

var docStyle = lipgloss.NewStyle().Margin(2, 2)

type (
	state int
)

var renders int = 0

const (
	stateShowStreams state = iota
	stateShowChat
)

func openTwitchStream(username string) error {
	var err error

	baseUrl := "https://twitch.tv/"
	url := baseUrl + username

	err = utils.OpenBrowser(url)
	return err
}

type item struct {
	title, desc, userLogin string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	apiClient api.Api
	state     state
	streams   streamerListModel
	chat      streamChatModel
	fatalErr  error
}

type streamerListModel struct {
	list    list.Model
	choices []api.Stream
}

type (
	errMsg           struct{ err error }
	chatConnectedMsg struct{}
)

func buildListFromTwitchStreamData(data []api.Stream) []list.Item {
	var s []list.Item
	for _, stream := range data {
		title := fmt.Sprintf("%s | %s", stream.UserDisplayName, stream.GameName)
		subtitle := fmt.Sprintf("%s | %d viewers", stream.Title, stream.ViewerCount)
		s = append(s, item{title, subtitle, stream.UserLogin})
	}
	return s
}

func InitialModel(twitchApi *api.Api) (*model, error) {
	choices, err := twitchApi.GetFollowedStreams()
	if err != nil {
		return nil, err
	}

	listItems := buildListFromTwitchStreamData(choices.Data)
	l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Who's online?"
	chatMessages := make([]api.ChatMessage, 0)
	return &model{
		streams:   streamerListModel{l, choices.Data},
		chat:      streamChatModel{&chatMessages, false},
		apiClient: *twitchApi,
		state:     stateShowStreams,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.fatalErr != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case chatConnectedMsg:
		m.chat.isConnected = true
		renders += 1
		return m, m.chat.Init()
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		if m.state == stateShowStreams {
			m.streams.list.SetSize(msg.Width-h, msg.Height-v)
		}

		// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "enter" key and the spacebar (a literal space) toggle
		case "enter", " ":
			if m.state == stateShowStreams {
				i, ok := m.streams.list.SelectedItem().(item)
				if ok {
					err := openTwitchStream(i.userLogin)
					if err != nil {
						return m, tea.Quit
					}
					m.state = stateShowChat
					if !m.chat.isConnected {
						return m, initChatCmd(m, i.userLogin)
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.state == stateShowStreams {
		m.streams.list, cmd = m.streams.list.Update(msg)
	}
	if m.state == stateShowChat {
		renders += 1
		m.chat, cmd = m.chat.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case stateShowChat:
		if m.chat.isConnected {
			out := fmt.Sprintf("RENDERS: %v\n\n", renders)
			out += m.chat.view()
			return docStyle.Render(out)
		}
		return "Connecting to chat"
	default:
		return docStyle.Render(m.streams.list.View())
	}
}
