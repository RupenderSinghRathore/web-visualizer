package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

type application struct {
	logger *log.Logger
	config *confugration
	client *http.Client
}
type confugration struct {
	port   int
	urlStr string
	crawl  struct {
		maxGoroutine,
		maxDepth int
		maxPages int
	}
	env     string
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

func main() {
	logger := log.New(os.Stdout)
	logger.SetTimeFormat(time.DateTime)
	logger.SetReportTimestamp(true)

	if len(os.Args) < 2 {
		logger.Error("not enough args: use 'server' or 'crawl'")
		os.Exit(1)
	}

	var cfg confugration

	app := application{
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		config: &cfg,
	}

	port, err := app.getPort()
	if err != nil {
		app.crashErr(err)
	}
	cfg.port = port

	switch os.Args[1] {
	case "crawl":
		crawlCmd := flag.NewFlagSet("crawl", flag.ContinueOnError)

		crawlFlags(crawlCmd, &cfg)
		commonFlags(crawlCmd, &cfg)
		crawlCmd.Parse(os.Args[2:])

		if err := app.fetchGraph(app.config.urlStr); err != nil {
			app.crashErr(err)
		}

	case "server":
		serverCmd := flag.NewFlagSet("server", flag.ContinueOnError)

		serverFlags(serverCmd, &cfg)
		commonFlags(serverCmd, &cfg)
		serverCmd.Parse(os.Args[2:])

		if err := app.serve(); err != nil {
			app.crashErr(err)
		}

	default:
		logger.Errorf("unknown arg: %s", os.Args[1])
		os.Exit(1)
	}
}
