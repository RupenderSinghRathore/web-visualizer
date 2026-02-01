package main

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

type application struct {
	logger *log.Logger
}

func main() {
	logger := log.New(os.Stdout)
	app := application{
		logger: logger,
	}
	app.normalizeUrl("")
	logger.SetTimeFormat(time.DateTime)
	// logger.SetTimeFormat(time.TimeOnly)
	logger.SetReportTimestamp(true)
}
