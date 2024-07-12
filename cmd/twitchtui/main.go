package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flothjl/twitchtui/internal/api"
	"github.com/flothjl/twitchtui/internal/app"
)

func main() {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	apiClient, err := api.New(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Error Creating API: %v", err)
	}

	if err != nil {
		log.Fatalf("Error getting user info: %v", err)
	}
	initialModel, err := app.InitialModel(apiClient)
	if err != nil {
		os.Exit(1)
	}
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
