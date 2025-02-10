package collector

import (
	"github.com/frebib/plex-exporter/plex"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type PlexCollector struct {
	Logger *log.Entry
	client *plex.PlexClient

	serverInfo         *prometheus.GaugeVec
	activeSessionCount *prometheus.GaugeVec
	libraryMetric      *prometheus.GaugeVec
	playerMetric       *prometheus.GaugeVec
}

func NewPlexCollector(c *plex.PlexClient, l *log.Entry) *PlexCollector {
	return &PlexCollector{
		Logger: l,
		client: c,

		serverInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "plex",
				Subsystem: "server",
				Name:      "info",
				Help:      "Information about Plex server",
			},
			[]string{"version", "platform"},
		),
		activeSessionCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "plex",
				Subsystem: "sessions",
				Name:      "active_count",
				Help:      "Number of active Plex sessions",
			},
			[]string{},
		),
		libraryMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "plex",
				Subsystem: "library",
				Name:      "section_size_count",
				Help:      "Number of items in a library section",
			},
			[]string{"name", "type"},
		),
		playerMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "plex",
				Subsystem: "player",
				Name:      "count",
				Help:      "Details about current players connected to Plex",
			},
			[]string{"device", "platform", "profile", "state", "local", "relayed", "secure"},
		),
	}
}

func (c *PlexCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serverInfo.Describe(ch)
	c.activeSessionCount.Describe(ch)
	c.libraryMetric.Describe(ch)
	c.playerMetric.Describe(ch)
}

func (c *PlexCollector) Collect(ch chan<- prometheus.Metric) {
	v, err := c.client.GetServerMetrics()
	if err != nil {
		c.Logger.Errorf("Could not retrieve server metrics: %s", err)
		return
	}

	c.Logger.Trace(v)
	c.serverInfo.WithLabelValues(v.Version, v.Platform).Set(1)
	c.activeSessionCount.WithLabelValues().Set(float64(v.ActiveSessions))

	c.playerMetric.Reset()
	for _, p := range v.Players {
		c.playerMetric.WithLabelValues(p.Device, p.Platform, p.Profile, p.State, p.Local, p.Relayed, p.Secure).Inc()
	}

	for _, l := range v.Libraries {
		c.libraryMetric.WithLabelValues(l.Name, l.Type).Set(float64(l.Size))
	}

	c.serverInfo.Collect(ch)
	c.activeSessionCount.Collect(ch)
	c.libraryMetric.Collect(ch)
	c.playerMetric.Collect(ch)
}
