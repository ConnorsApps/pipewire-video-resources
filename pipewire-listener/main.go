package main

import (
	"context"
	"os"
	"time"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"
	"github.com/ConnorsApps/pipewire-video-resources/pipewire-listener/config"
	"github.com/ConnorsApps/pipewire-video-resources/pipewire-listener/reconciler"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var localConfig *config.Config

// Only watch for nodes
func eventFilter(e *pwmonitor.Event) bool {
	return e.Type == pwmonitor.EventNode || e.IsRemovalEvent()
}

// Goal: turn speakers on if Volt is online, and off otherwise
// Uses pw-dump ---monitor --no-colors

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		reconciler      = reconciler.New(localConfig)
		eventsChan      = make(chan []*pwmonitor.Event)
		gotInitialState = make(chan interface{})
	)

	go func() {
		panic(pwmonitor.Monitor(ctx, eventsChan, eventFilter))
	}()

	go func() {
		reconciler.GetInitialState()
		gotInitialState <- nil
	}()

	// Wait for initial state to be set
	<-gotInitialState

	for {
		var (
			changeMade bool
			events     = <-eventsChan
		)
		for _, e := range events {
			if reconciler.UpdateState(e) {
				changeMade = true
			}
		}

		if changeMade {
			reconciler.Reconcile()
		}
	}
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.Kitchen,
	})

	config.MustRead(os.Getenv("CONFIG_PATH"), &localConfig)
}
