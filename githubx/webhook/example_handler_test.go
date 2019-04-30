// Package webhook defines a generic handler for GitHub App & Webhook events.
package webhook_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/githubx/webhook"
)

func Example() {
	eventHandler := &webhook.Handler{
		PushEvent: func(event *github.PushEvent) error {
			fmt.Printf("%#v\n", event)
			return nil
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/event_handler", eventHandler)
	http.ListenAndServe(":8888", mux)
}
