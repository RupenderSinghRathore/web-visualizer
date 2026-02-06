package main

import (
	"flag"
	"fmt"
	"mime"
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
	concurrencyLimit,
	maxDepth,
	maxWidth int
}

func isHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html"
}

func main() {
	var cfg confugration

	urlc := flag.String("url", "", "url to crawl")
	flag.IntVar(&cfg.concurrencyLimit, "concurrency_limit", 20, "max goroutines to be spawned")
	flag.IntVar(&cfg.maxDepth, "max_depth", 100, "max depth of the links graph")
	flag.IntVar(&cfg.maxWidth, "max_width", 100, "max width of the links graph")

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
		graph, err := app.crawlPage(*urlc)
		if err != nil {
			logger.Error(err)
		}
		for endPoint, edge := range graph {
			fmt.Printf("%s(%d, %d) -> %v\n", endPoint, edge.Visited, edge.Status, edge.Links)
		}
	}
}
