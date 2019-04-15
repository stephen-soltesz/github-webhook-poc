package local

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stephen-soltesz/github-webhook-poc/events/issues/iface"

	"github.com/stephen-soltesz/pretty"

	"github.com/stephen-soltesz/github-webhook-poc/githubx"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/events/issues"
)

var (
	// ErrNewClient is returned when a new Github client fails.
	ErrNewClient = fmt.Errorf("Failed to allocate new Github client")
)

// Config contains
type Config struct {
	// Delay is
	Delay time.Duration

	getIface func(client *github.Client) iface.Issues
}

// NewConfig creates a new config instantce.
func NewConfig(delay time.Duration) *Config {
	// Initialize the new config instance with a default getIface function.
	return &Config{Delay: delay, getIface: getIface}
}

// Client collects local data needed for these operations.
type getInstallationer interface {
	GetInstallation() *github.Installation
}

func getSafeID(event getInstallationer) int64 {
	if event.GetInstallation() != nil {
		return event.GetInstallation().GetID()
	}
	return 0
}

func getIface(client *github.Client) iface.Issues {
	return iface.NewIssues(client.Issues)
}

func genSprintLabels(now time.Time) []string {
	year, week := now.ISOWeek()
	sprintLabel := fmt.Sprintf("Sprint %d", (week+1)/2)
	yearLabel := fmt.Sprintf("%d", year)
	return []string{sprintLabel, yearLabel}
}

// IssuesEvent .
func (c *Config) IssuesEvent(event *github.IssuesEvent) error {
	client := githubx.NewClient(getSafeID(event))
	if client == nil {
		return ErrNewClient
	}
	ev := issues.NewEvent(c.getIface(client), event)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Lose the race for loading page load after "Submit new issue"
	// so that the new label is visible to user.
	// time.Sleep(c.Delay)

	var err error
	var resp *github.Response
	var issue *github.Issue
	var labels []*github.Label

	log.Println("IssuesEvent:", ev.GetAction(), ev.GetIssue().GetHTMLURL())
	switch {
	case ev.GetAction() == "opened" || ev.GetAction() == "reopened":
		pretty.Print(event)
		labels, resp, err = ev.AddIssueLabels(ctx, []string{"review/triage"})
	case ev.GetAction() == "closed":
		pretty.Print(event)
		resp, err = ev.RemoveIssueLabels(ctx, []string{"review/triage"})
	case ev.GetAction() == "labeled":
		// Note: the issue labels always include the label triggering the event.
		fmt.Println("label")
		pretty.Print(event)
		current := []string{}
		for _, label := range event.GetIssue().Labels {
			current = append(current, label.GetName())
		}
		eventLabel := event.GetLabel().GetName()
		fmt.Println("Event label:", eventLabel)
		fmt.Println("Current labels:", current)
		switch strings.ToLower(eventLabel) {
		case "backlog":
			resp, err = ev.RemoveIssueLabels(ctx, []string{"review/triage", "current", "closed"})
		case "current":
			_, resp, err = ev.AddIssueLabels(ctx, genSprintLabels(time.Now()))
			if err == nil && len(current) != 0 {
				resp, err = ev.RemoveIssueLabels(ctx, []string{"review/triage", "backlog", "closed"})
			}
			// current = append(current, genSprintLabels(time.Now())...)
			// current = slice.FilterStrings(current, []string{"review/triage", "backlog", "closed"})
			// issue, resp, err = ev.SetIssueLabels(ctx, current)
		case "closed":
			issue, resp, err = ev.CloseIssue(ctx, nil)
		}
	case ev.GetAction() == "unlabeled":
		pretty.Print(event)
		fmt.Println("unlabel")
	default:
		log.Println("IssuesEvent: ignoring unsupported action:", ev.GetAction())
	}
	if err != nil {
		log.Println("IssuesEvent: error:       ", resp, err)
	}
	if labels != nil {
		log.Println("IssuesEvent: okay: ", resp, labels)
	}
	if issue != nil {
		labels := ""
		for _, currentLabel := range issue.Labels {
			labels += currentLabel.GetName() + " "
		}
		log.Println("IssuesEvent: okay: ", resp, labels)
	}
	return nil
}

// InstallationEvent handles events when an application is installed for the
// first time.
func InstallationEvent(event *github.InstallationEvent) error {
	pretty.Print(event)
	return nil
}

// InstallationRepositoriesEvent handles events when repositories are added or
// removed from a particular application installation.
func InstallationRepositoriesEvent(event *github.InstallationRepositoriesEvent) error {
	pretty.Print(event)
	return nil
}
