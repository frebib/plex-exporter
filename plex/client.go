package plex

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

type PlexClient struct {
	Logger *log.Entry
	server *Server
}

func NewPlexClient(s *Server, l *log.Entry) (*PlexClient, error) {
	return &PlexClient{
		Logger: l,
		server: s,
	}, nil
}

// GetServerMetrics fetches all metrics for each server and returns them in a map
// with the servers' names as keys.
func (c *PlexClient) GetServerMetrics() (ServerMetric, error) {
	logger := c.Logger.WithFields(log.Fields{"server": c.server.Name})

	serverMetric := ServerMetric{
		Version:  c.server.Version,
		Platform: c.server.Platform,
	}

	// Get active session status
	sessionStatus, err := c.server.GetSessionStatus()
	if err != nil {
		logger.Debugf("Could not get session status: %s", err)
		return serverMetric, err
	}
	serverMetric.ActiveSessions = sessionStatus.Size

	for _, metadata := range sessionStatus.Metadata {
		sessionMetric := SessionMetric{
			Location: metadata.Session.Location,
		}

		playerMetric := PlayerMetric{
			Device:   metadata.Player.Device,
			Platform: metadata.Player.Platform,
			Profile:  metadata.Player.Profile,
			State:    metadata.Player.State,
			Local:    metadata.Player.Local,
			Relayed:  metadata.Player.Relayed,
			Secure:   metadata.Player.Secure,
		}

		serverMetric.Sessions = append(serverMetric.Sessions, sessionMetric)
		serverMetric.Players = append(serverMetric.Players, playerMetric)
	}

	// Get library metrics
	library, err := c.server.GetLibrary()
	if err != nil {
		logger.Debugf("Could not get library: %s", err)
		return serverMetric, err
	}

	for _, section := range library.Sections {
		id, err := strconv.Atoi(section.ID)
		if err != nil {
			logger.Debugf("Could not convert sections ID to int. (%s)", section.ID)
		}
		size, err := c.server.GetSectionSize(id)
		if err != nil {
			logger.Debugf("Could not get section size for \"%s\": %s", section.Name, err)
			return serverMetric, err
		}
		libraryMetric := LibraryMetric{
			Name: section.Name,
			Type: section.Type,
			Size: size,
		}

		serverMetric.Libraries = append(serverMetric.Libraries, libraryMetric)
	}

	return serverMetric, nil
}
