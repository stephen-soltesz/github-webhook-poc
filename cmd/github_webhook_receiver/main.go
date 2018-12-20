// github_webhook_receiver is a proof of concept for creating a github
// webhook using the go-github package. github_webhook_receiver only supports
// issue events.
package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/go-github/github"
	"github.com/kr/pretty"
	"golang.org/x/oauth2"
)

var (
	slackURL     string
	authToken    string
	fGithubOwner string
	fGithubRepo  string
)

const (
	usage = `
USAGE:

Github receiver requires a github GITHUB_AUTH_TOKEN and target github --owner
and --repo names.

`
)

func init() {
	slackURL = os.Getenv("SLACK_DEST_URL")
	authToken = os.Getenv("GITHUB_AUTH_TOKEN")
	flag.StringVar(&fGithubOwner, "owner", "", "The github user or organization name.")
	flag.StringVar(&fGithubRepo, "repo", "", "The repository where issues are created.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

// A Client manages communication with the Github API.
type Client struct {
	// githubClient is an authenticated client for accessing the github API.
	GithubClient *github.Client
	// owner is the github project (e.g. github.com/<owner>/<repo>).
	owner string
	// repo is the github repository under the above owner.
	repo string
}

// NewClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewClient(owner, repo, authToken string) *Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	client := &Client{
		GithubClient: github.NewClient(oauth2.NewClient(ctx, tokenSource)),
		owner:        owner,
		repo:         repo,
	}
	return client
}

type localClient Client

func (l *localClient) addLabel(event *github.IssuesEvent, label string) error {
	ctx := context.Background()
	currentLabels := &[]string{}
	issue := event.GetIssue()
	for _, currentLabel := range issue.Labels {
		log.Println(currentLabel)
		if currentLabel.GetName() == label {
			// No need to continue, since the label is already present.
			return nil
		}
		*currentLabels = append(*currentLabels, currentLabel.GetName())
	}
	// Append the new label to the current labels.
	*currentLabels = append(*currentLabels, label)
	newIssue, resp, err := l.GithubClient.Issues.Edit(
		ctx,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		issue.GetNumber(),
		&github.IssueRequest{
			Labels: currentLabels,
		},
	)
	if err != nil {
		log.Println("Error: ", resp)
		return err
	}
	log.Printf("Success: %#v", newIssue.Labels)
	return nil
}
func (l *localClient) rmLabel(event *github.IssuesEvent, label string) error {
	ctx := context.Background()
	currentLabels := &[]string{}
	issue := event.GetIssue()
	for _, currentLabel := range issue.Labels {
		log.Println(currentLabel)
		if currentLabel.GetName() == label {
			// Skip this label since we want it removed.
			continue
		}
		*currentLabels = append(*currentLabels, currentLabel.GetName())
	}
	newIssue, resp, err := l.GithubClient.Issues.Edit(
		ctx,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		issue.GetNumber(),
		&github.IssueRequest{
			Labels: currentLabels,
		},
	)
	if err != nil {
		log.Println("Error: ", resp)
		return err
	}
	log.Printf("Success: %#v", newIssue.Labels)
	return nil
}

func setupRoutes(l *localClient) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	http.HandleFunc("/event_handler", l.eventHandler)
}

func supportedEvent(events []string, supported string) bool {
	for _, event := range events {
		if event == supported {
			return true
		}
	}
	return false
}

func getZenMessage(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(b)
}

func (l *localClient) eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported event type", http.StatusMethodNotAllowed)
		return
	}
	// Key must match the key used when registering the webhook at github.com
	payload, err := github.ValidatePayload(r, []byte("test"))
	if err != nil {
		http.Error(w, "payload did not validate", http.StatusInternalServerError)
		return
	}
	// fmt.Println(string(payload))
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, "failed to parse webhook", http.StatusInternalServerError)
		return
	}

	switch event := event.(type) {
	case *github.PingEvent:
		// Ping events occur during first registration.
		// If we return without error, the webhook is registered successfully.
		if !supportedEvent(event.Hook.Events, "issue_comment") &&
			!supportedEvent(event.Hook.Events, "issues") &&
			!supportedEvent(event.Hook.Events, "pull_request") {
			fmt.Println("Unsupported event types", event.Hook.Events)
			http.Error(w, "Unsupported event type", http.StatusNotImplemented)
		}
		pretty.Print(r.Header)
		pretty.Print(event)
		return

	case *github.IssueCommentEvent:
		// TODO: scan issue comments for commands.
		pretty.Print(event)

	case *github.IssuesEvent:
		// TODO: detect close events.
		var err error
		switch event.GetAction() {
		case "opened":
			log.Println("issue opened:", event.GetIssue().GetNumber())
			err = l.addLabel(event, "review/triage")
		case "reopened":
			log.Println("issue reopened:", event.GetIssue().GetNumber())
			err = l.addLabel(event, "review/triage")
		case "closed":
			log.Println("issue closed:", event.GetIssue().GetNumber())
			err = l.rmLabel(event, "review/triage")
		}
		if err != nil {
			log.Println(err)
		}

	case *github.PullRequestEvent:
		data := url.Values{}
		msg := ""
		switch event.GetAction() {
		case "review_requested":
			msg = fmt.Sprintf("%s added %s as a reviewer on: %s",
				event.PullRequest.User.GetLogin(),
				event.RequestedReviewer.GetLogin(),
				event.PullRequest.GetHTMLURL())
		case "review_request_removed":
			msg = fmt.Sprintf("%s removed %s as a reviewer on: %s",
				event.PullRequest.User.GetLogin(),
				event.RequestedReviewer.GetLogin(),
				event.PullRequest.GetHTMLURL())
		default:
			fmt.Printf("Ignoring action: %q\n", event.GetAction())
			pretty.Println(event.Action)
			pretty.Println(event.Sender.Login)
			return
		}
		fmt.Println(msg)

		data.Set("payload", `{"text": "`+msg+`"}`)
		if slackURL != "" {
			resp, err := http.PostForm(slackURL, data)
			if err != nil {
				http.Error(w, "Bad post to slack", http.StatusInternalServerError)
				fmt.Println("Failed post to slack", err)
				return
			}
			fmt.Println()
			fmt.Println(resp.Status, resp.StatusCode)
		}
		fmt.Println(data)
	default:
		fmt.Println("Unsupported event type")
		http.Error(w, "Unsupported event type", http.StatusNotImplemented)
	}
	return
}

func main() {
	flag.Parse()
	if authToken == "" || fGithubOwner == "" || fGithubRepo == "" {
		flag.Usage()
		os.Exit(1)
	}
	client := (*localClient)(NewClient(fGithubOwner, fGithubRepo, authToken))

	addr := ":8901"
	setupRoutes(client)
	fmt.Println("Listening on ", addr)
	http.ListenAndServe(addr, nil)
}
