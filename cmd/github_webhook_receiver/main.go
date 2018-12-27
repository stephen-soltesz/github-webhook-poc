// github_webhook_receiver is a proof of concept for creating a github
// webhook that automatically adds and removes labels from issues.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stephen-soltesz/github-webhook-poc/local"
	"github.com/stephen-soltesz/github-webhook-poc/webhook"

	// "github.com/kr/pretty"
	"github.com/stephen-soltesz/github-webhook-poc/config"
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
	privateKey    string
	fListenAddr   string
)

func init() {
	authToken = os.Getenv("GITHUB_AUTH_TOKEN")
	webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
	privateKey = os.Getenv("GITHUB_PRIVATE_KEY")
	// appID = os.Getenv("GITHUB_APP_ID")
	flag.StringVar(&fListenAddr, "addr", ":3000", "The github user or organization name.")

	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

func usageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", usage)
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if (authToken == "" && privateKey == "") || webhookSecret == "" {
		flag.Usage()
		os.Exit(1)
	}

	client := &local.Client{
		Ignore: config.Load(local.DefaultConfigURL),
	}

	handle := &webhook.Handler{
		WebhookSecret: webhookSecret,
		IssuesEvent:   client.IssuesEvent,
		PushEvent:     client.PushEvent,
	}
	//	http.HandleFunc("/", usageHandler)
	// http.HandleFunc("/event_handler", handle.Request)
	http.HandleFunc("/", handle.Request)

	fmt.Println("Listening on ", fListenAddr)
	log.Fatal(http.ListenAndServe(fListenAddr, nil))
}
