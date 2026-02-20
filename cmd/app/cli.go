package main

import (
	"RupenderSinghRathore/web-visualizer/internal/data"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/list"
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

func (app *application) fetchGraph(urlStr string) error {
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

	urlStruct, err := validateUrl(urlStr)
	if err != nil {
		return err
	}
	graph := app.crawlUrl(urlStruct)

	stopAnimation()

	if err := app.printGraph(graph, urlStruct, linkedText); err != nil {
		return err
	}

	return nil
}

func (app *application) printGraph(
	graph data.Graph,
	urlStruct *url.URL,
	formatter func(u string, status, visited int, base string) string,
) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	base := urlStruct.Scheme + "://" + urlStruct.Host
	base = strings.TrimSuffix(base, "/")

	var dfs func(u string)
	visited := make(map[string]bool)
	dfs = func(u string) {
		e := graph[u]
		if e == nil {
			return
		}

		l.AppendItem(formatter(u, e.Status, e.Visited, base))
		l.Indent()

		if !visited[u] {
			visited[u] = true
			for _, next := range e.Links {
				dfs(next)
			}
		}
		l.UnIndent()
	}
	origin := app.getPath(urlStruct)
	dfs(origin)

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	_, err := writer.WriteString(l.Render())
	// _, err := writer.WriteString(l.RenderHTML())
	if err != nil {
		return err
	}
	writer.WriteByte('\n')
	return nil
}
