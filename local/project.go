package local

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/stephen-soltesz/github-webhook-poc/events/projects"

	"github.com/stephen-soltesz/github-webhook-poc/githubx"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/pretty"
)

var mux = sync.Mutex{}

// ProjectCardEvent prints a card event.
func ProjectCardEvent(event *github.ProjectCardEvent) error {
	mux.Lock()
	defer mux.Unlock()
	switch event.GetAction() {
	case "moved":
		event.GetProjectCard().GetColumnID()
		client := githubx.NewClient(event.GetInstallation().GetID())
		if client == nil {
			return ErrNewClient
		}
		ev := projects.NewEvent(client, event)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		col := ev.GetColumn(ctx)
		if col != nil {
			pretty.Print(col)
		}
		// ev := issues.NewEvent(c.getIface(client), event)
	}
	log.Print("project card event:")
	pretty.Print(event)
	return nil
}

// ProjectColumnEvent prints a column event.
func ProjectColumnEvent(event *github.ProjectColumnEvent) error {
	mux.Lock()
	defer mux.Unlock()
	log.Print("project column event:")
	pretty.Print(event)
	return nil
}

// ProjectEvent prints a generic project event.
func ProjectEvent(event *github.ProjectEvent) error {
	mux.Lock()
	defer mux.Unlock()
	log.Print("project event:")
	pretty.Print(event)
	return nil
}
