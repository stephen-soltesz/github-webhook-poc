package issues

import (
	"context"
	"log"

	"github.com/stephen-soltesz/github-webhook-poc/events/issues/iface"
	"github.com/stephen-soltesz/pretty"

	"github.com/google/go-github/github"
)

// Event encapsulates operations on a *github.IssuesEvent.
type Event struct {
	//	*github.Client
	*github.IssuesEvent
	iface.Issues
}

// NewEvent creates a new Event based on the given client and event.
// func NewEvent(client *github.Client, event *github.IssuesEvent) *Event {
func NewEvent(issues iface.Issues, event *github.IssuesEvent) *Event {
	return &Event{
		IssuesEvent: event,
		Issues:      issues,
	}
}

// EditIssue updates the event issue using the given request.
func (ev *Event) EditIssue(
	ctx context.Context,
	req *github.IssueRequest) (*github.Issue, *github.Response, error) {

	log.Println("Issues.Edit:",
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		ev.GetIssue().GetNumber(),
		pretty.SprintPlain(req))

	return ev.Issues.Edit(
		ctx,
		ev.GetRepo().GetOwner().GetLogin(),
		ev.GetRepo().GetName(),
		ev.GetIssue().GetNumber(),
		req)
}

// AddIssueLabel adds a label to the event issue.
func (ev *Event) AddIssueLabel(
	ctx context.Context, label string) (*github.Issue, *github.Response, error) {

	// Check curent labels for the new label and collect the list.
	currentLabels := []string{}
	for _, currentLabel := range ev.GetIssue().Labels {
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

// RemoveIssueLabel removes the given label from the event issue. If the label
// is not found, no action is taken.
func (ev *Event) RemoveIssueLabel(
	ctx context.Context, label string) (*github.Issue, *github.Response, error) {

	labels := filterLabels(ev.GetIssue().Labels, label)
	if len(labels) != len(ev.GetIssue().Labels) {
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
