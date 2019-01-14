package local

import (
	"context"
	"fmt"
	"log"
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
	time.Sleep(c.Delay)

	var err error
	var resp *github.Response
	var issue *github.Issue

	log.Println("IssuesEvent:", ev.GetAction(), ev.GetIssue().GetHTMLURL())
	switch {
	case ev.GetAction() == "opened" || ev.GetAction() == "reopened":
		issue, resp, err = ev.AddIssueLabel(ctx, "review/triage")
	case ev.GetAction() == "closed":
		issue, resp, err = ev.RemoveIssueLabel(ctx, "review/triage")
	default:
		log.Println("IssuesEvent: ignoring unsupported action:", ev.GetAction())
	}
	if err != nil {
		log.Println("IssuesEvent: error:       ", resp, err)
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
