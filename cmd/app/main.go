package main

import (
	"flag"
	"fmt"
	"mime"
	"net/http"
	"net/url"
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
	port  int
	crawl struct {
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

func isHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html"
}

func main() {
	var cfg confugration

	urlc := flag.String("url", "", "url to crawl")

	flag.IntVar(&cfg.port, "port", 8080, "Port to be used")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.IntVar(
		&cfg.crawl.maxGoroutine,
		"concurrency",
		20,
		"max goroutines for crawler",
	)
	flag.IntVar(&cfg.crawl.maxDepth, "max-depth", 100, "max width of the links graph")
	flag.IntVar(&cfg.crawl.maxPages, "max-pages", 1000, "max width of the links graph")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	logger := log.New(os.Stdout)
	logger.SetTimeFormat(time.DateTime)
	logger.SetReportTimestamp(true)

	app := application{
		logger: logger,
		config: &cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	if *urlc != "" {
		go func() {
			for {
				t := 100 * time.Millisecond
				for _, r := range `-\|/` {
					fmt.Printf("\r%c", r)
					time.Sleep(t)
				}
			}
		}()

		urlB, err := url.ParseRequestURI(*urlc)
		if err != nil {
			logger.Error(err)
		}
		graph, err := app.crawlPage(urlB)
		if err != nil {
			logger.Error(err)
		}
		for endPoint, edge := range graph {
			if endPoint == "" {
				endPoint = "/"
			}
			fmt.Printf("%s(%d, %d) -> [ ", endPoint, edge.Visited, edge.Status)
			for link := range edge.Links {
				if link == "" {
					fmt.Printf("%s ", "/")
				} else {
					fmt.Printf("%s ", link)
				}
			}
			fmt.Printf("]\n\n")
		}
		os.Exit(0)
	}

	if err := app.serve(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
