package local

import (
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/config"
	"github.com/stephen-soltesz/github-webhook-poc/events/issues"
	"github.com/stephen-soltesz/github-webhook-poc/slice"
	"github.com/stephen-soltesz/pretty"
)

var (
	ConfigFilename   = "config.json"
	DefaultConfigURL = "https://raw.githubusercontent.com/stephen-soltesz/public-issue-test/master/config.json"
)

type Client struct {
	*github.Client
	config.Repos
}

func (c *Client) PushEvent(event *github.PushEvent) error {
	refs := strings.Split(event.GetRef(), "/")
	log.Println(refs)
	log.Println(refs[len(refs)-1])
	branch := refs[len(refs)-1]

	// TODO: parse full name.
	log.Println(event.GetRepo().GetFullName())
	for _, commit := range event.Commits {
		log.Println(commit.GetID(), commit.Added, commit.Removed, commit.Modified)
		if slice.ContainsString(commit.Added, ConfigFilename) ||
			slice.ContainsString(commit.Removed, ConfigFilename) ||
			slice.ContainsString(commit.Modified, ConfigFilename) {
			// TODO: we should always use 'master' branch.
			// Construct raw content url.
			configURL := strings.Join(
				[]string{"https://raw.githubusercontent.com",
					event.GetRepo().GetFullName(),
					branch,
					ConfigFilename}, "/")
			c.Repos = config.Load(configURL)
			pretty.Print(c.Repos)
			break
		}
	}
	return nil
}

func (c *Client) IssuesEvent(event *github.IssuesEvent) error {
	ev := issues.NewEvent(c.Client, event)
	source := ev.GetRepo().GetHTMLURL()
	if _, ok := c.Repos[source]; !ok {
		log.Println("Ignoring:", source)
		return nil
	}

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
