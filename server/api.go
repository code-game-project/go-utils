package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/code-game-project/go-utils/external"
)

type API struct {
	baseURL string
	http    *http.Client
}

func NewAPI(url string) (*API, error) {
	url = external.TrimURL(url)
	tls := external.IsTLS(url)

	api := &API{
		baseURL: external.BaseURL("http", tls, url),
		http:    http.DefaultClient,
	}
	api.http.Timeout = 10 * time.Second

	resp, err := api.http.Get(api.baseURL + "/api/info")
	if err != nil {
		return nil, fmt.Errorf("Cannot reach %s.", api.baseURL)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK && external.HasContentType(resp.Header, "application/json") {
		api.baseURL += "/api"
	}

	return api, nil
}

func (a *API) BaseURL() string {
	return a.baseURL
}
