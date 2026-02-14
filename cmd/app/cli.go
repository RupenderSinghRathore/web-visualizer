package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
)

func commonFlags(fg *flag.FlagSet, cfg *confugration) {
	fg.StringVar(
		&cfg.env,
		"env",
		"development",
		"Environment (development|staging|production)",
	)

	fg.IntVar(
		&cfg.crawl.maxGoroutine,
		"concurrency",
		20,
		"max goroutines for crawler",
	)
	fg.IntVar(&cfg.crawl.maxDepth, "max-depth", 100, "max width of the links graph")
	fg.IntVar(&cfg.crawl.maxPages, "max-pages", 1000, "max width of the links graph")
}

func serverFlags(serverCmd *flag.FlagSet, cfg *confugration) {

	serverCmd.IntVar(&cfg.port, "port", 8080, "Port to be used")
	serverCmd.Float64Var(
		&cfg.limiter.rps,
		"limiter-rps",
		2,
		"Rate limiter maximum requests per second",
	)
	serverCmd.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	serverCmd.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

}
func crawlFlags(crawlCmd *flag.FlagSet, cfg *confugration) {
	crawlCmd.StringVar(&cfg.urlStr, "url", "", "url to crawl")
}

func (app *application) printGraph(urlStr string) error {
	if urlStr == "" {
		return errors.New("empty url")
	}

	stopSpinner := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go app.spinningAnimation(stopSpinner, &wg)

	var stopOnce sync.Once
	stopAnimation := func() {
		stopOnce.Do(func() {
			close(stopSpinner)
			wg.Wait()
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

		for _, link := range edge.Links {
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
