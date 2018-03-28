// github_webhook_receiver is a proof of concept for creating a github
// webhook using the go-github package. github_webhook_receiver only supports
// issue events.
package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/google/go-github/github"
	"github.com/kr/pretty"
)

var (
	slackURL string
)

func init() {
	slackURL = os.Getenv("SLACK_DEST_URL")
}

func setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	http.HandleFunc("/event_handler", eventHandler)
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

func eventHandler(w http.ResponseWriter, r *http.Request) {
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
		pretty.Print(event)

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
	addr := ":8901"
	setupRoutes()
	fmt.Println("Listening on ", addr)
	http.ListenAndServe(addr, nil)
}
