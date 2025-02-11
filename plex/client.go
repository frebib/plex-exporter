package plex

import (
	"strconv"
	"sync"

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

	var (
		data   ServerMetric
		wg     sync.WaitGroup
		errors = make(chan error)
	)

	wg.Add(3)

	call := func(f func() error) {
		err := f()
		if err != nil {
			errors <- err
		}
		wg.Done()
	}

	// Get server info
	go call(func() error {
		info, err := c.server.GetServerInfo()
		if err != nil {
			logger.WithError(err).Debug("Failed to get server info")
			return err
		}
		data.Platform = info.Platform
		data.Version = info.Version
		return nil
	})

	// Get active session status
	go call(func() error {
		sessionStatus, err := c.server.GetSessionStatus()
		if err != nil {
			logger.WithError(err).Debug("Could not get session status")
			return err
		}
		data.ActiveSessions = sessionStatus.Size

		for _, metadata := range sessionStatus.Metadata {
			data.Players = append(data.Players, PlayerMetric{
				Device:   metadata.Player.Device,
				Platform: metadata.Player.Platform,
				Profile:  metadata.Player.Profile,
				State:    metadata.Player.State,
				Local:    strconv.FormatBool(metadata.Player.Local),
				Relayed:  strconv.FormatBool(metadata.Player.Relayed),
				Secure:   strconv.FormatBool(metadata.Player.Secure),
			})
		}
		return nil
	})

	// Get library metrics
	go call(func() error {
		library, err := c.server.GetLibrary()
		if err != nil {
			logger.WithError(err).Debug("Could not get library")
			return err
		}

		n := len(library.Sections)
		data.Libraries = make([]LibraryMetric, n)
		wg.Add(n)

		for i, section := range library.Sections {
			go func(i int) {
				defer wg.Done()
				id, err := strconv.Atoi(section.ID)
				if err != nil {
					logger.WithError(err).Debugf("Could not convert sections ID to int. (%s)", section.ID)
					errors <- err
					return
				}
				size, err := c.server.GetSectionSize(id)
				if err != nil {
					logger.WithError(err).Debugf("Could not get section size for \"%s\"", section.Name)
					errors <- err
					return
				}
				data.Libraries[i] = LibraryMetric{
					Name: section.Name,
					Type: section.Type,
					Size: size,
				}
			}(i)
		}
		return nil
	})

	go func() {
		wg.Wait()

		// Signify no errors and that all jobs finished
		select {
		case errors <- nil:
		default: // if the channel is full
		}
	}()

	// Wait for an error (or nil), then return
	return data, <-errors
}
