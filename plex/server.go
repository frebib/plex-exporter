package plex

import (
	"crypto/tls"
	"encoding/json"
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
	Version    string
	Platform   string
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

	serverInfo, err := server.getServerInfo()
	if err != nil {
		return nil, err
	}

	server.ID = serverInfo.ID
	server.Name = serverInfo.Name
	server.Version = serverInfo.Version
	server.Platform = serverInfo.Platform

	return server, nil
}

func (s *Server) getServerInfo() (*api.ServerInfoResponse, error) {
	serverInfoResponse := api.ServerInfoResponse{}

	body, err := s.get(fmt.Sprintf(ServerInfoURI, s.BaseURL))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &serverInfoResponse)
	if err != nil {
		return nil, err
	}

	return &serverInfoResponse, nil
}

func (s *Server) GetSessionStatus() (*api.SessionList, error) {
	sessionList := api.SessionList{}

	body, err := s.get(fmt.Sprintf(StatusURI, s.BaseURL))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &sessionList)
	if err != nil {
		return nil, err
	}

	return &sessionList, nil
}

func (s *Server) GetLibrary() (*api.LibraryResponse, error) {
	libraryResponse := api.LibraryResponse{}

	body, err := s.get(fmt.Sprintf(LibraryURI, s.BaseURL))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &libraryResponse)
	if err != nil {
		return nil, err
	}

	return &libraryResponse, nil
}

func (s *Server) GetSectionSize(id int) (int, error) {
	// We don't want to get every item in the library section
	// these headers make sure we only get metadata
	eh := map[string]string{
		"X-Plex-Container-Start": "0",
		"X-Plex-Container-Size":  "0",
	}
	maps.Copy(eh, s.headers)

	sectionResponse := api.SectionResponse{}

	_, body, err := sendRequest("GET", fmt.Sprintf(SectionURI, s.BaseURL, id), eh, s.httpClient)
	if err != nil {
		return -1, err
	}

	err = json.Unmarshal(body, &sectionResponse)
	if err != nil {
		return -1, err
	}

	return sectionResponse.TotalSize, nil
}

func (s *Server) get(url string) ([]byte, error) {
	_, body, err := sendRequest("GET", url, s.headers, s.httpClient)
	return body, err
}

func (s *Server) head(url string) (*http.Response, error) {
	resp, _, err := sendRequest("HEAD", url, s.headers, s.httpClient)
	return resp, err
}
