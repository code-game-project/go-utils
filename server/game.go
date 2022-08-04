package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/code-game-project/go-utils/external"
)

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
		return nil, fmt.Errorf("Couldn't decode /info data.")
	}

	return data.Players, nil
}
