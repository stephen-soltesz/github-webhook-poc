package local

import (
	"sync"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/pretty"
)

var mux = sync.Mutex{}

// ProjectCardEvent prints a card event.
func ProjectCardEvent(event *github.ProjectCardEvent) error {
	mux.Lock()
	defer mux.Unlock()
	pretty.Print(event)
	return nil
}

// ProjectColumnEvent prints a column event.
func ProjectColumnEvent(event *github.ProjectColumnEvent) error {
	mux.Lock()
	defer mux.Unlock()
	pretty.Print(event)
	return nil
}

// ProjectEvent prints a generic project event.
func ProjectEvent(event *github.ProjectEvent) error {
	mux.Lock()
	defer mux.Unlock()
	pretty.Print(event)
	return nil
}
