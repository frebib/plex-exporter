package plex

import (
	"crypto/tls"
	"fmt"
	"maps"
	"net/http"
	"runtime"
	"time"

	"github.com/frebib/plex-exporter/config"
	"github.com/frebib/plex-exporter/plex/api"
	"github.com/frebib/plex-exporter/version"
)

type Server struct {
	ID         string
	Name       string
	BaseURL    string
	token      string
	httpClient *http.Client
	headers    map[string]string
}

const TestURI = "%s/identity"
const ServerInfoURI = "%s/media/providers"
const StatusURI = "%s/status/sessions"
const LibraryURI = "%s/library/sections"
const SectionURI = "%s/library/sections/%d/all"

var DefaultHeaders = map[string]string{
	"User-Agent":               fmt.Sprintf("plex_exporter/%s", version.Version),
	"Accept":                   "application/json",
	"X-Plex-Platform":          runtime.GOOS,
	"X-Plex-Version":           version.Version,
	"X-Plex-Client-Identifier": fmt.Sprintf("plex-exporter-v%s", version.Version),
	"X-Plex-Device-Name":       "Plex Exporter",
	"X-Plex-Product":           "Plex Exporter",
	"X-Plex-Device":            runtime.GOOS,
}

func NewServer(c config.PlexServerConfig) (*Server, error) {
	headers := maps.Clone(DefaultHeaders)
	headers["X-Plex-Token"] = c.Token

	server := &Server{
		BaseURL: c.BaseURL,
		token:   c.Token,
		headers: headers,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.Insecure},
			},
		},
	}

	// Check the server, and pre-cache server id/name
	_, err := server.GetServerInfo()
	return server, err
}

func (s *Server) GetServerInfo() (*api.ServerInfoResponse, error) {
	info, err := httpRequest[api.ServerInfoResponse](s.httpClient, http.MethodGet, fmt.Sprintf(ServerInfoURI, s.BaseURL), s.headers)
	if err != nil {
		return nil, err
	}
	// Cache last-known ID (shouldn't ever change) and name
	s.ID = info.ID
	s.Name = info.Name
	return info, nil
}

func (s *Server) GetSessionStatus() (*api.SessionList, error) {
	return httpRequest[api.SessionList](s.httpClient, http.MethodGet, fmt.Sprintf(StatusURI, s.BaseURL), s.headers)
}

func (s *Server) GetLibrary() (*api.LibraryResponse, error) {
	return httpRequest[api.LibraryResponse](s.httpClient, http.MethodGet, fmt.Sprintf(LibraryURI, s.BaseURL), s.headers)
}

func (s *Server) GetSectionSize(id int) (int, error) {
	// We don't want to get every item in the library section
	// these headers make sure we only get metadata
	headers := map[string]string{
		"X-Plex-Container-Start": "0",
		"X-Plex-Container-Size":  "0",
	}
	maps.Copy(headers, s.headers)

	url := fmt.Sprintf(SectionURI, s.BaseURL, id)
	resp, err := httpRequest[api.SectionResponse](s.httpClient, http.MethodGet, url, headers)
	if err != nil {
		return -1, err
	}
	return resp.TotalSize, nil
}
