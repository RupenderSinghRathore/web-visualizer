package main

import (
	"fmt"
	"os"
	"time"
)

const (
	EraseLineANSI = "\r\033[K"
)

func (app *application) crashErr(err error) {
	app.logger.Error(err)
	os.Exit(1)
}
func (app *application) spinningAnimation(ch <-chan struct{}) {
	defer app.wg.Done()
	spinner := `-\|/`
	n := len(spinner)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ch:
			return
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r%c", spinner[i])
			i = (i + 1) % n
		}
	}
}
