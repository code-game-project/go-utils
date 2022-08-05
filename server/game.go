package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/code-game-project/go-utils/external"
)

type GameListEntry struct {
	Id      string `json:"id"`
	Players int    `json:"players"`
}

func (a *API) ListGames() (private int, public []GameListEntry, err error) {
	type response struct {
		Private int             `json:"private"`
		Public  []GameListEntry `json:"public"`
	}
	url := a.baseURL + "/games"
	res, err := a.http.Get(url)
	if err != nil || res.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("Couldn't access %s.", url)
	}
	if !external.HasContentType(res.Header, "application/json") {
		return 0, nil, fmt.Errorf("%s doesn't return JSON.", url)
	}
	defer res.Body.Close()

	var data response
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return 0, nil, fmt.Errorf("Couldn't decode response data: %s", err)
	}

	return data.Private, data.Public, nil
}

func (a *API) GetPlayers(gameId string) (map[string]string, error) {
	type response struct {
		Players map[string]string `json:"players"`
	}
	url := a.baseURL + "/games/" + gameId + "/players"
	res, err := a.http.Get(url)
	if err != nil || res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Couldn't access %s.", url)
	}
	if !external.HasContentType(res.Header, "application/json") {
		return nil, fmt.Errorf("%s doesn't return JSON.", url)
	}
	defer res.Body.Close()

	var data response
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode response data: %s", err)
	}

	return data.Players, nil
}
