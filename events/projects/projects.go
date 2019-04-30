package projects

import (
	"context"
	"log"

	"github.com/google/go-github/github"
)

// Event encapsulates operations on a *github.IssuesEvent.
type Event struct {
	*github.Client
	*github.ProjectCardEvent
	// iface.Issues
}

// NewEvent creates a new Event based on the given client and event.
func NewEvent(client *github.Client, event *github.ProjectCardEvent) *Event {
	return &Event{
		Client:           client,
		ProjectCardEvent: event,
	}
}

func (ev *Event) GetColumn(ctx context.Context) *github.ProjectColumn {
	col, resp, err := ev.Client.Projects.GetProjectColumn(
		ctx, ev.ProjectCardEvent.GetProjectCard().GetColumnID())
	log.Printf("%#v", resp.Rate)
	if err != nil {
		log.Println(err)
		return nil
	}
	return col
}
