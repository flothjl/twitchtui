package app

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flothjl/twitchtui/internal/api"
)

var docStyle = lipgloss.NewStyle().Margin(2, 2)

func openTwitchStream(username string) error {
	var err error

	baseUrl := "https://twitch.tv/"
	url := baseUrl + username

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

type item struct {
	title, desc, userLogin string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list      list.Model
	choices   []api.Stream
	cursor    int
	apiClient api.Api
}

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
	return &model{
		list:      l,
		choices:   choices.Data,
		apiClient: *twitchApi,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

		// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c":
			return m, tea.Quit

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				openTwitchStream(i.userLogin)
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}
