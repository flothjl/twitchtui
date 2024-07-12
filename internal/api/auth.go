package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type AccessToken struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	UserID    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
	Scopes    []string `json:"scopes"`
	Token     string   `json:"token,omitempty"`
}

func openBrowser(url string) error {
	var err error

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

func callbackHandler(code chan<- string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callbackValue := r.URL.Query().Get("code")
		w.Write([]byte("Success! Please return to your terminal"))
		code <- callbackValue
	}
}

func getAccessToken(clientID, clientSecret, code string, tokenURL string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "http://localhost:7394")

	req, err := http.NewRequest("POST", tokenURL+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token, status code: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func validateAccessToken(accessToken string, authUrl string) (*AccessToken, error) {
	req, err := http.NewRequest("GET", authUrl, nil)
	var tokenDetails *AccessToken
	if err != nil {
		return tokenDetails, err
	}

	req.Header.Set("Authorization", "OAuth "+accessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return tokenDetails, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(resp.Body).Decode(&tokenDetails); err != nil {
			return tokenDetails, fmt.Errorf("error decoding response: %v", err)
		}
		tokenDetails.Token = accessToken
		return tokenDetails, nil
	case http.StatusUnauthorized:
		return tokenDetails, fmt.Errorf("token invalid")
	default:
		return tokenDetails, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func getTokenFromFile(filePath string) (*AccessToken, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	token := &AccessToken{}
	err = json.NewDecoder(file).Decode(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func writeAccessTokenToFile(accessToken *AccessToken, tokenPath string) error {
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(tokenPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(*accessToken)
}

func (api *Api) authorizeUser() error {
	// check to see if token exists and if it's valid
	validationUrl := fmt.Sprintf("%s/%s", api.authUrl, "validate")
	accessToken, err := getTokenFromFile(api.tokenPath)
	if err == nil {
		accessToken, err = validateAccessToken(accessToken.Token, validationUrl)
		if err == nil {
			api.clientToken = accessToken
			return nil
		}
	}

	params := url.Values{
		"client_id":     {api.clientID},
		"redirect_uri":  {"http://localhost:7394"},
		"response_type": {"code"},
		"scope":         {"user:read:follows"},
	}

	authorizationUrl := fmt.Sprintf("%s/%s?%s", api.authUrl, "authorize", params.Encode())
	server := &http.Server{
		Addr: ":7394",
	}
	openBrowser(authorizationUrl)

	codeChan := make(chan string)
	go func() {
		http.HandleFunc("/", callbackHandler(codeChan))
		fmt.Println("Listening on http://localhost:7394/")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()

	code := <-codeChan

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
		return err
	}

	fmt.Println("Server stopped")
	// get access token using code
	tokenString, err := getAccessToken(api.clientID, api.clientSecret, code, api.authUrl)
	if err != nil {
		return err
	}

	accessToken, err = validateAccessToken(tokenString, validationUrl)
	if err != nil {
		return err
	}
	// write token to file
	err = writeAccessTokenToFile(accessToken, api.tokenPath)
	if err != nil {
		fmt.Printf("error writing token: %v", err)
	}
	api.clientToken = accessToken
	return nil
}
