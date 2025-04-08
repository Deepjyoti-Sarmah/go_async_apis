package reports

import "net/http"

const baseUrl = "https://botw-compendium.herokuapp.com/api/v3/compendium"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl    string
	httpClient HttpClient
}

func NewClient(baseUrl string, httpClient HttpClient) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

type Monster struct {
	Name           string   `json:"name"`
	Id             int      `json:"id"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	Image          string   `json:"image"`
	CommonLocation []string `json:"common_location"`
	Drops          []string `json:"drops"`
	Dlc            bool     `json:"dlc"`
}

type GetMonstersResponse struct {
	Date []Monster `json:"data"`
}
