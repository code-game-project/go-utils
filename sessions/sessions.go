package sessions

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

type Session struct {
	GameURL      string `json:"-"`
	Username     string `json:"-"`
	GameId       string `json:"game_id"`
	PlayerId     string `json:"player_id"`
	PlayerSecret string `json:"player_secret"`
	Path         string `json:"-"`
}

var sessionsPath = filepath.Join(xdg.DataHome, "codegame", "games")

func NewSession(gameURL, username, gameId, playerId, playerSecret string) Session {
	return Session{
		GameURL:      gameURL,
		Username:     username,
		GameId:       gameId,
		PlayerId:     playerId,
		PlayerSecret: playerSecret,
	}
}

// ListSessions returns a map of all game URLs in the session store mapped to a list of usernames.
func ListSessions() (map[string][]string, error) {
	gameDirs, err := os.ReadDir(filepath.Join(sessionsPath))
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(gameDirs))
	for _, dir := range gameDirs {
		if !dir.IsDir() {
			continue
		}
		userFiles, err := os.ReadDir(filepath.Join(sessionsPath, dir.Name()))
		if err != nil {
			return nil, err
		}
		users := make([]string, 0, len(userFiles))
		for _, dir := range userFiles {
			if !dir.IsDir() && strings.HasSuffix(dir.Name(), ".json") {
				users = append(users, string(dir.Name()[:len(dir.Name())-5]))
			}
		}
		unescapedGameDir, err := url.PathUnescape(dir.Name())
		if err != nil {
			return nil, err
		}
		result[unescapedGameDir] = users
	}

	return result, nil
}

// ListGames returns a list of all game URLs in the session store.
func ListGames() ([]string, error) {
	gameDirs, err := os.ReadDir(filepath.Join(sessionsPath))
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(gameDirs))
	for _, dir := range gameDirs {
		if !dir.IsDir() {
			continue
		}
		unescaped, err := url.PathUnescape(dir.Name())
		if err == nil {
			result = append(result, unescaped)
		}
	}

	return result, nil
}

// ListUsernames returns a list of usernames for the game.
func ListUsernames(gameURL string) ([]string, error) {
	userFiles, err := os.ReadDir(filepath.Join(sessionsPath, url.PathEscape(gameURL)))
	if err != nil {
		return nil, err
	}
	users := make([]string, 0, len(userFiles))
	for _, dir := range userFiles {
		if !dir.IsDir() && strings.HasSuffix(dir.Name(), ".json") {
			users = append(users, string(dir.Name()[:len(dir.Name())-5]))
		}
	}
	return users, nil
}

// Load a session from the session store.
func LoadSession(gameURL, username string) (Session, error) {
	data, err := os.ReadFile(filepath.Join(sessionsPath, url.PathEscape(gameURL), username+".json"))
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal(data, &session)

	session.GameURL = gameURL
	session.Username = username

	return session, err
}

// Save the session to the session store.
func (s Session) Save() error {
	if s.GameURL == "" {
		return errors.New("empty game url")
	}
	dir := filepath.Join(sessionsPath, url.PathEscape(s.GameURL))
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, s.Username+".json"), data, 0644)
}

// Remove the session from the session store.
func (s Session) Remove() error {
	if s.GameURL == "" {
		return nil
	}
	dir := filepath.Join(sessionsPath, url.PathEscape(s.GameURL))
	err := os.Remove(filepath.Join(dir, s.Username+".json"))
	if err != nil {
		return err
	}

	dirs, err := os.ReadDir(dir)
	if err == nil && len(dirs) == 0 {
		os.Remove(dir)
	}
	return nil
}
