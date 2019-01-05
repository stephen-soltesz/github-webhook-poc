package local

import (
	"log"
	"time"

	"github.com/stephen-soltesz/github-webhook-poc/githubx"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/events/issues"
)

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

// IssuesEvent .
func IssuesEvent(event *github.IssuesEvent) error {
	client := githubx.NewClient(getSafeID(event))
	ev := issues.NewEvent(client, event)
	source := ev.GetRepo().GetHTMLURL()

	log.Println("Issues:", source)
	// Lose the race for loading page load after "Submit new issue"
	// so that the new label is visible to user.
	time.Sleep(time.Second)
	var err error
	var resp *github.Response
	var issue *github.Issue

	switch {
	case event.GetAction() == "opened" || event.GetAction() == "reopened":
		log.Println("Issues: open: ", event.GetIssue().GetHTMLURL())
		issue, resp, err = ev.AddIssueLabel("review/triage")
	case event.GetAction() == "closed":
		log.Println("Issues: close:", event.GetIssue().GetHTMLURL())
		issue, resp, err = ev.RemoveIssueLabel("review/triage")
	default:
		log.Println("Issues: ignoring unsupported action:", event.GetAction())
	}
	if err != nil {
		log.Println("Issues: error:       ", resp, err)
	}
	if issue != nil {
		labels := ""
		for _, currentLabel := range issue.Labels {
			labels += currentLabel.GetName() + " "
		}
		log.Println("Issues: okay: ", resp, labels)
	}
	return nil
}
