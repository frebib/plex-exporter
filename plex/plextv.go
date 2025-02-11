package plex

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/frebib/plex-exporter/config"
	"github.com/frebib/plex-exporter/plex/api"
)

var ErrPinNotAuthorised = errors.New("pin not authorised")

type PinRequest struct {
	Pin `json:"pin"`
}

type Pin struct {
	Id        int       `json:"id"`
	Code      string    `json:"code"`
	Expiry    time.Time `json:"expires_at"`
	Trusted   bool      `json:"trusted"`
	AuthToken string    `json:"auth_token"`
}

func DiscoverServers(token string) ([]*Server, error) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	// This endpoint only supports XML.
	// I want to specify the "Accept: application/xml" header
	// to make sure that if the endpoint does support JSON in
	// the future it won't break the application.
	headers := map[string]string{
		"Accept":       "application/xml",
		"X-Plex-Token": token,
	}
	maps.Copy(headers, DefaultHeaders)

	resp, err := httpRequest[api.DeviceList](httpClient, http.MethodGet, "https://plex.tv/api/resources?includeHttps=1", headers)
	if err != nil {
		return nil, err
	}

	var servers []*Server
	for _, device := range resp.Devices {
		// Device is a server and is owned by user
		if strings.Contains(device.Roles, "server") && device.Owned {
			// Loop over the server's connections and use the first one to work
			// If none of the connections work, the server is skipped
			for _, conn := range device.Connections {
				s, err := NewServer(config.PlexServerConfig{
					BaseURL:  conn.URI,
					Token:    device.AccessToken,
					Insecure: false,
				})
				if err != nil {
					fmt.Println(err)
					continue
				}
				servers = append(servers, s)
				break
			}
		}
	}

	return servers, nil
}

// GetPinRequest creates a PinRequest using the Plex API and returns it.
func GetPinRequest() (*PinRequest, error) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	return httpRequest[PinRequest](httpClient, http.MethodPost, "https://plex.tv/pins", DefaultHeaders)
}

// GetTokenFromPinRequest takes in a PinRequest and checks if it has been authenticated.
// If it has been authenticated it returns the token.
// If it has not been authenticated it returns an empty string.
func GetTokenFromPinRequest(p *PinRequest) (string, error) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	resp, err := httpRequest[PinRequest](httpClient, http.MethodGet, fmt.Sprintf("https://plex.tv/pins/%d", p.Id), DefaultHeaders)
	if err != nil {
		return "", err
	}

	if resp.AuthToken == "" {
		return "", ErrPinNotAuthorised
	}
	return resp.AuthToken, nil
}
