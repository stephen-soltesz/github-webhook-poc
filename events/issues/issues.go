package issues

import (
	"context"
	"log"

	"github.com/stephen-soltesz/github-webhook-poc/slice"

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

func (ev *Event) CloseIssue(ctx context.Context, labels []string) (*github.Issue, *github.Response, error) {
	var req *github.IssueRequest
	closed := "closed"
	if labels == nil {
		req = &github.IssueRequest{
			State: &closed,
		}
	} else {
		req = &github.IssueRequest{
			State:  &closed,
			Labels: &labels,
		}
	}
	return ev.EditIssue(ctx, req)
}

func (ev *Event) SetIssueLabels(ctx context.Context, labels []string) (
	*github.Issue, *github.Response, error) {
	return ev.EditIssue(
		ctx,
		&github.IssueRequest{
			Labels: &labels,
		},
	)
}

// AddIssueLabel adds a label to the event issue.
func (ev *Event) AddIssueLabels(
	ctx context.Context, labels []string) (*github.Issue, *github.Response, error) {

	// Check curent labels for the new label and collect the list.
	currentLabels := []string{}
	for _, current := range ev.GetIssue().Labels {
		currentLabels = append(currentLabels, current.GetName())
	}
	// Append the new label to the current labels.
	for _, label := range labels {
		if !slice.ContainsString(currentLabels, label) {
			currentLabels = append(currentLabels, label)
		}
	}
	if len(ev.GetIssue().Labels) == len(currentLabels) {
		return nil, nil, nil
	}
	return ev.EditIssue(
		ctx,
		&github.IssueRequest{
			Labels: &currentLabels,
		},
	)
}

// RemoveIssueLabel removes the given label from the event issue. If the label
// is not found, no action is taken.
func (ev *Event) RemoveIssueLabels(
	ctx context.Context, labels []string) (*github.Issue, *github.Response, error) {

	currentLabels := filterLabels(ev.GetIssue().Labels, labels)
	if len(currentLabels) != len(ev.GetIssue().Labels) {
		return ev.EditIssue(
			ctx,
			&github.IssueRequest{
				Labels: &currentLabels,
			},
		)
	}
	return nil, nil, nil
}

func filterLabels(labels []github.Label, remove []string) []string {
	currentLabels := []string{}
	for _, currentLabel := range labels {
		if slice.ContainsString(remove, currentLabel.GetName()) {
			// Skip this label since we want it removed.
			continue
		}
		currentLabels = append(currentLabels, currentLabel.GetName())
	}
	return currentLabels
}
