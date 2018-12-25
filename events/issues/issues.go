package issues

import (
	"context"
	"log"

	"github.com/stephen-soltesz/pretty"

	"github.com/google/go-github/github"
)

// Event encapsulates operations on a *github.IssuesEvent.
type Event struct {
	*github.Client
	*github.IssuesEvent
}

// NewEvent creates a new Event based on the given client and event.
func NewEvent(client *github.Client, event *github.IssuesEvent) *Event {
	return &Event{
		client,
		event,
	}
}

// EditIssue updates the event issue using the given request.
func (ev *Event) EditIssue(
	ctx context.Context, req *github.IssueRequest) (
	*github.Issue, *github.Response, error) {
	issue := ev.GetIssue()
	log.Println("Issues.Edit:",
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		issue.GetNumber(),
		pretty.SprintPlain(req))
	return ev.Issues.Edit(
		ctx,
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		issue.GetNumber(),
		req)
}

// AddIssueLabel adds a label to the event issue.
func (ev *Event) AddIssueLabel(label string) (*github.Issue, *github.Response, error) {
	ctx := context.Background()
	currentLabels := []string{}
	issue := ev.GetIssue()
	for _, currentLabel := range issue.Labels {
		if currentLabel.GetName() == label {
			// No need to continue, since the label is already present.
			return nil, nil, nil
		}
		currentLabels = append(currentLabels, currentLabel.GetName())
	}
	// Append the new label to the current labels.
	currentLabels = append(currentLabels, label)
	return ev.EditIssue(
		ctx,
		&github.IssueRequest{
			Labels: &currentLabels,
		},
	)
}

// RemoveIssueLabel removes the given label from the event issue. If the label is not
// found, no action is taken.
func (ev *Event) RemoveIssueLabel(label string) (*github.Issue, *github.Response, error) {
	ctx := context.Background()
	issue := ev.GetIssue()
	labels := filterLabels(issue.Labels, label)
	if len(labels) != len(issue.Labels) {
		return ev.EditIssue(
			ctx,
			&github.IssueRequest{
				Labels: &labels,
			},
		)
	}
	return nil, nil, nil
}

func filterLabels(labels []github.Label, remove string) []string {
	currentLabels := []string{}
	for _, currentLabel := range labels {
		if currentLabel.GetName() == remove {
			// Skip this label since we want it removed.
			continue
		}
		currentLabels = append(currentLabels, currentLabel.GetName())
	}
	return currentLabels
}
