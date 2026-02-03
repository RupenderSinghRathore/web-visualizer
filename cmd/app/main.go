package main

import (
	"flag"
	"fmt"
	"mime"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type application struct {
	logger *log.Logger
	currencyLimit int
	mutex sync.Mutex
}

func isHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html"
}

func main() {
	urlc := flag.String("url", "", "url to crawl")
	flag.Parse()
	logger := log.New(os.Stdout)
	logger.SetTimeFormat(time.DateTime)
	logger.SetReportTimestamp(true)

	app := application{
		logger: logger,
		currencyLimit: 20,
		mutex: sync.Mutex{},
	}
	if *urlc != "" {
		pages, err := app.crawlPage(*urlc)
		if err != nil {
			logger.Error(err)
		}
		// logger.Info(pages)
		for page := range pages {
			// logger.Info(page)
			fmt.Println(page)
		}
	}
}
