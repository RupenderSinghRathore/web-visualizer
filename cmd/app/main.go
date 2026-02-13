package main

import (
	"bufio"
	"flag"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type application struct {
	logger *log.Logger
	config *confugration
	client *http.Client
	wg     sync.WaitGroup
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
		if err := app.printGraph(*urlc); err != nil {
			app.crashErr(err)
		}
		os.Exit(0)
	}

	if err := app.serve(); err != nil {
		app.crashErr(err)
	}
}

func (app *application) printGraph(urlStr string) error {
	stopSpinner := make(chan struct{})

	app.wg.Add(1)
	go app.spinningAnimation(stopSpinner)

	var stopOnce sync.Once
	stopAnimation := func() {
		stopOnce.Do(func() {
			close(stopSpinner)
			app.wg.Wait()
			fmt.Fprintf(os.Stderr, EraseLineANSI)
		})
	}
	defer stopAnimation()

	urlB, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return err
	}
	graph, err := app.crawlPage(urlB)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	stopAnimation()

	for endPoint, edge := range graph {
		_, err = fmt.Fprintf(writer, "%s(%d, %d) -> [ ", endPoint, edge.Visited, edge.Status)
		if err != nil {
			return err
		}

		for link := range edge.Links {
			_, err = fmt.Fprintf(writer, "%s ", link)
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintf(writer, "]\n\n")
		if err != nil {
			return err
		}
	}

	return nil
}
