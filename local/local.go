package local

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stephen-soltesz/github-webhook-poc/events/issues/iface"
	"github.com/stephen-soltesz/github-webhook-poc/slice"

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

// func containsAtMostOneOf(labels []github.Label, sequence []string) map[string]bool {
func findSequenceLabels(labels, sequence []string) map[string]bool {
	extra := map[string]bool{}
	for _, label := range labels {
		if slice.ContainsString(sequence, label) {
			extra[label] = true
		}
	}
	return extra
}

func apply2(name string, current, sequence []string) {
}

func apply(applyLabel bool, name string, current, sequence []string) {
	if applyLabel {
		// Add label.
		if slice.ContainsString(sequence, name) {
			// Adding a sequence label.
			extra := findSequenceLabels(current, sequence)
			switch len(extra) {
			case 0:
				// 'name' is a sequence label and there are currently no other sequence labels.
				// TODO: add 'name' label.
			case 1:
				// 'name' is a sequence label and there is currently one other sequence labels.
				// TODO: is 'name' or 'extra' greater.
			default:
			}
		}
	} else {
		// Remove label.

	}
}

func findMaxRank(sequence []string, found map[string]bool) int {
	maxRank := -1
	for key := range found {
		rank := slice.IndexString(sequence, key)
		if rank > maxRank {
			maxRank = rank
		}
	}
	return maxRank
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

	// sequence := []string{"review/triage", "backlog", "current", "close"}

	var err error
	var resp *github.Response
	var issue *github.Issue

	log.Println("IssuesEvent:", ev.GetAction(), ev.GetIssue().GetHTMLURL())
	switch {
	case ev.GetAction() == "opened" || ev.GetAction() == "reopened":
		pretty.Print(event)
		issue, resp, err = ev.AddIssueLabels(ctx, []string{"review/triage"})
	case ev.GetAction() == "closed":
		pretty.Print(event)
		issue, resp, err = ev.RemoveIssueLabels(ctx, []string{"review/triage"})
	case ev.GetAction() == "labeled":
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
			issue, resp, err = ev.RemoveIssueLabels(ctx, []string{"review/triage", "current", "closed"})
		case "current":
			current = append(current, genSprintLabels(time.Now())...)
			current = slice.FilterStrings(current, []string{"review/triage", "backlog", "closed"})
			issue, resp, err = ev.SetIssueLabels(ctx, current)
		case "closed":
			issue, resp, err = ev.CloseIssue(ctx, nil)
		}
		// --------------------------------------------------------------------
		// Note: the issue labels always include the label triggering the event.
		/*
			found := findSequenceLabels(current, sequence)
			switch len(found) {
			case 0:
				// No sequence labels, add the first.
				issue, resp, err = ev.AddIssueLabels(ctx, sequence[:1])
			case 1:
				// One sequence label.
			default:
				// Multiple sequence labels.
				newRank := slice.IndexString(sequence, eventLabel)
				prevRank := findMaxRank(sequence, found)
				fmt.Printf("Ranks: %d < %d\n", prevRank, newRank)
				if prevRank <= newRank {
					// Remove previous labels.
					del := []string{}
					for key := range found {
						if key == eventLabel {
							continue
						}
						del = append(del, key)
					}
					fmt.Println("Remove:", del)
					issue, resp, err = ev.RemoveIssueLabels(ctx, del)
				} else {
					// Remove new label.
					fmt.Println("Remove:", eventLabel)
					issue, resp, err = ev.RemoveIssueLabels(ctx, []string{eventLabel})
				}
			}
		*/

		/*
			issue opened -> ApplyOneOf('r/t', 'backlog', 'current', 'closed')
			add label 'backlog' -> ApplyOneOf('r/t', 'backlog', 'current', 'closed')
			add label 'current' -> ApplyOneOf('r/t', 'backlog', 'current', 'closed') &&
								   ApplySprintLabels(time.Now())
			add label 'close' -> ApplyOneOf('r/t', 'backlog', 'current', 'closed') &&
			                     CloseIssue()

			remove 'review/traige' -> ApplyOneOf('backlog', 'current', 'closed')
			remove 'current' -> ApplyOneOf('backlog', 'review/triage', 'closed')
			remove 'backlog' -> ApplyOneOf('review/triage', 'current', 'closed')
			remove 'close ' -> do nothing

			removed := event.GetLabel().GetName()
			if ApplyOneOf(removed, invariants) {
			}
			if removing 'review/triage' label, then
			   * add labels for 'backlog', unless 'current' is being added.

			if removing 'current' label, then
			   * if 'closed' label is present, then do nothing.
			   * else add 'review/triage' label, remove sprint label.

			if removing 'backlog' label, then
			   * if 'review/triage', 'current', or 'closed' present, do nothing.

			if removing 'closed' label, then
			   * if issue is closed, do nothing.
			   * if issue is open, add review/triage label.
		*/
	case ev.GetAction() == "unlabeled":
		pretty.Print(event)
		fmt.Println("unlabel")
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
