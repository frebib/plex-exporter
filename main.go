package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/frebib/plex-exporter/collector"
	"github.com/frebib/plex-exporter/config"
	"github.com/frebib/plex-exporter/plex"
	"github.com/frebib/plex-exporter/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Token(c *cli.Context) error {
	fmt.Printf("Attempting to authenticate with Plex\n")

	pinRequest, err := plex.GetPinRequest()
	if err != nil {
		return fmt.Errorf("could not make a pin request: %w", err)
	}

	fmt.Printf("\n\tGot PIN Code: %s\n\tGo to https://plex.tv/pin and enter pin to authenticate.\n\n", pinRequest.Code)

	// Repeatedly check pin request
	ticker := time.NewTicker(time.Second * 5)
	for t := range ticker.C {
		if pinRequest.Expiry.Before(t) {
			ticker.Stop()
			return fmt.Errorf("PIN expired, exiting")
		}

		token, err := plex.GetTokenFromPinRequest(pinRequest)
		if err != nil {
			if !errors.Is(err, plex.ErrPinNotAuthorised) {
				ticker.Stop()
				return fmt.Errorf("could not check PIN request: %w", err)
			}
		} else {
			fmt.Printf("Authenticated successfully!\nYour token is: %s\n", token)
			ticker.Stop()
			break
		}
	}
	return nil
}

func Run(c *cli.Context) error {
	// Loading configuration
	conf, err := config.Load(c)
	if err != nil {
		return err
	}

	var serverList []*plex.Server

	l := log.New()

	for _, serverConf := range conf.Servers {
		plexServer, err := plex.NewServer(serverConf)
		if err != nil {
			log.WithFields(log.Fields{"BaseURL": serverConf.BaseURL}).Errorf("Could not add server: %s", err)
		} else {
			serverList = append(serverList, plexServer)
		}
	}

	if conf.AutoDiscover {
		discoveryList, err := plex.DiscoverServers(conf.Token)
		if err == nil {
			serverList = append(serverList, discoveryList...)
		}
	}

	l.Infof("Found %d working servers", len(serverList))

	reg := prometheus.NewPedanticRegistry()

	for _, server := range serverList {
		// Create a Plex client
		clientLogger := log.WithFields(log.Fields{"context": "client", "server": server.Name})
		client, err := plex.NewPlexClient(server, clientLogger)

		// Create the Prometheus collector
		collectorLogger := log.WithFields(log.Fields{"context": "collector", "server": server.Name})
		pc := collector.NewPlexCollector(client, collectorLogger)
		prometheus.WrapRegistererWith(
			prometheus.Labels{"server_name": server.Name, "server_id": server.ID}, reg,
		).MustRegister(pc)

		if err != nil {
			return err
		}
	}

	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	// Start HTTP server
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Infof("Beginning to serve on port %s", conf.ListenAddress)
	log.Fatal(http.ListenAndServe(conf.ListenAddress, nil))
	return nil
}

func Init(c *cli.Context) error {
	verbose := c.String("log-level")
	format := c.String("format")

	// Set verbosity level
	switch verbose {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "err":
		log.SetLevel(log.ErrorLevel)
	default:
		return fmt.Errorf("Available log levels are trace, debug, info, warn, err")
	}

	// Set log format
	switch format {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("Available log formats are text, json")
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "plex_exporter"
	app.Usage = "A Prometheus exporter that exports metrics on Plex Media Server."
	app.Version = version.Version

	flags := []cli.Flag{
		cli.StringFlag{
			Name:   "config-path, c",
			Value:  "/etc/plex_exporter/config.yaml",
			Usage:  "Path config file",
			EnvVar: "PLEX_CONFIG_PATH,CONFIG_PATH",
		},
		cli.StringFlag{
			Name:   "listen-address, l",
			Value:  ":9594",
			Usage:  "Port for server",
			EnvVar: "PLEX_LISTEN_ADDR,LISTEN_ADDR,ADDR",
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "info",
			Usage:  "Verbosity level of logs",
			EnvVar: "PLEX_LOG_LEVEL,LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "format, f",
			Value:  "text",
			Usage:  "Output format of logs",
			EnvVar: "PLEX_LOG_FORMAT,LOG_FORMAT",
		},
		cli.BoolFlag{
			Name:   "auto-discover, a",
			Usage:  "Auto discover Plex servers from plex.tv",
			EnvVar: "PLEX_AUTO_DISCOVER,AUTO_DISCOVER",
		},
		cli.StringFlag{
			Name:   "plex-server, p",
			Usage:  "Address of Plex Media Server",
			EnvVar: "PLEX_SERVER",
		},
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "Authentication token for Plex Media Server",
			EnvVar: "PLEX_TOKEN,TOKEN",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "Get authentication token from plex.tv",
			Action:  Token,
		},
	}

	app.Action = Run
	app.Before = Init
	app.Flags = flags

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
