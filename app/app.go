package app

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flothjl/twitchnerds/api"
)

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

type model struct {
	choices   []api.Stream
	cursor    int
	apiClient api.Api
}

func InitialModel(twitchApi *api.Api) (*model, error) {
	choices, err := twitchApi.GetFollowedStreams()
	if err != nil {
		return nil, err
	}
	return &model{
		choices:   choices.Data,
		apiClient: *twitchApi,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		//"r" key reloads the list of streams
		case "r":
			reloadedList, err := m.apiClient.GetFollowedStreams()
			if err == nil {
				m.choices = reloadedList.Data
			}
		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			openTwitchStream(m.choices[m.cursor].UserLogin)
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Who's live?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s, %s\n", cursor, choice.UserDisplayName, choice.GameName)
	}
	s += "\nPress r to reload the list.\n"
	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
