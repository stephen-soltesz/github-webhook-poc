// github_webhook_receiver is a proof of concept for creating a github
// webhook that automatically adds and removes labels from issues.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/kr/pretty"
	"golang.org/x/oauth2"
)

const (
	usage = `
USAGE:

  GitHub webhook receiver requires two environment variables at runtime:

   * GITHUB_AUTH_TOKEN - to authenticate with the GitHub API
   * GITHUB_WEBHOOK_SECRET - to validate the registered webhook

ALLOCATE AUTH TOKEN:

  Allocate a "Personal Access Token" by visiting github.com:

   * https://github.com/settings/tokens

REGISTER WEBHOOK:

  Register the webhook by visiting your repo on github.com. Click "Settings"
  and then "Webhooks". You should land on a URL like:

   * https://github.com/<owner>/<repo>/settings/hooks

  Click "Add Webhook".

  Use the payload URL (note the "/event_handler" path):

   * Payload URL: https://<service-url>/event_handler
   * Secret: value matching the environment variable GITHUB_WEBHOOK_SECRET
   * Select "Let me select individual events."
   * Check "Issues".
   * Uncheck "Pushes".
   * Click the green "Add Webhook" button.

  If the registration was successful, there should be a green checkmark. If
  registration failed, there will be a red "X".

FLAGS:

`
)

var (
	authToken     string
	webhookSecret string
	fListenAddr   string
)

func init() {
	authToken = os.Getenv("GITHUB_AUTH_TOKEN")
	webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
	flag.StringVar(&fListenAddr, "addr", ":8901", "The github user or organization name.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

// A Client extends operations on the Github API.
type Client struct {
	*github.Client
}

// NewClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewClient(authToken string) *Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	client := &Client{
		github.NewClient(oauth2.NewClient(ctx, tokenSource)),
	}
	return client
}

func (c *Client) updateIssueLabels(
	ctx context.Context, event *github.IssuesEvent, labels []string) (*github.Issue, *github.Response, error) {
	issue := event.GetIssue()
	return c.Issues.Edit(
		ctx,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		issue.GetNumber(),
		&github.IssueRequest{
			Labels: &labels,
		},
	)
}

func (c *Client) addIssueLabel(event *github.IssuesEvent, label string) (*github.Issue, *github.Response, error) {
	ctx := context.Background()
	currentLabels := &[]string{}
	issue := event.GetIssue()
	for _, currentLabel := range issue.Labels {
		if currentLabel.GetName() == label {
			// No need to continue, since the label is already present.
			return nil, nil, nil
		}
		*currentLabels = append(*currentLabels, currentLabel.GetName())
	}
	// Append the new label to the current labels.
	*currentLabels = append(*currentLabels, label)
	return c.updateIssueLabels(ctx, event, *currentLabels)
}

func (c *Client) removeIssueLabel(event *github.IssuesEvent, label string) (*github.Issue, *github.Response, error) {
	ctx := context.Background()
	issue := event.GetIssue()
	return c.updateIssueLabels(ctx, event, filterLabels(issue.Labels, label))
}

func (c *Client) eventHandler(w http.ResponseWriter, r *http.Request) {
	// Restrict event handling to POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported event type", http.StatusMethodNotAllowed)
		return
	}
	// webhookSecret should match the secret used when registering the webhook.
	payload, err := github.ValidatePayload(r, []byte(webhookSecret))
	if err != nil {
		http.Error(w, "payload did not validate", http.StatusInternalServerError)
		return
	}
	// Convert the payload into a specific github event type.
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, "failed to parse webhook", http.StatusInternalServerError)
		return
	}

	switch event := event.(type) {
	case *github.PingEvent:
		// Ping events occur during first registration.
		// If we return without error, the webhook is registered successfully.
		if !eventIsSupported(event.Hook.Events, "issues") {
			pretty.Print(r.Header)
			log.Println("Unsupported event type:", event.Hook.Events)
			http.Error(w, "Unsupported event type:", http.StatusNotImplemented)
		}
		return

	case *github.IssuesEvent:
		var err error
		var resp *github.Response
		var issue *github.Issue

		switch {
		case event.GetAction() == "opened" || event.GetAction() == "reopened":
			log.Println("Issue open: ", event.GetIssue().GetHTMLURL())
			issue, resp, err = c.addIssueLabel(event, "review/triage")
		case event.GetAction() == "closed":
			log.Println("Issue close:", event.GetIssue().GetHTMLURL())
			issue, resp, err = c.removeIssueLabel(event, "review/triage")
		default:
			log.Println("Ignoring unsupported issue action:", event.GetAction())
		}
		if err != nil {
			log.Println("Error:       ", resp, err)
		}
		if issue != nil {
			log.Println("Success:     ", resp, filterLabels(issue.Labels, ""))
		}

	default:
		fmt.Println("Unsupported event type")
		http.Error(w, "Unsupported event type", http.StatusNotImplemented)
	}
	return
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

func eventIsSupported(events []string, supported string) bool {
	for _, event := range events {
		if event == supported {
			return true
		}
	}
	return false
}

func setupRoutes(c *Client) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", usage)
		flag.CommandLine.SetOutput(w)
		flag.PrintDefaults()
	})
	http.HandleFunc("/event_handler", c.eventHandler)
}

func main() {
	flag.Parse()
	if authToken == "" || webhookSecret == "" {
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(authToken)
	setupRoutes(client)

	fmt.Println("Listening on ", fListenAddr)
	log.Fatal(http.ListenAndServe(fListenAddr, nil))
}
