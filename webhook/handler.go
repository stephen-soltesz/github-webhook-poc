package webhook

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/slice"
	"github.com/stephen-soltesz/pretty"
)

var (
	debug string
)

func init() {
	debug = os.Getenv("DEBUG_HANDLER")
}

// Handler ...
type Handler struct {
	// WebhookSecret should be the same value used to register this webhook in github.
	WebhookSecret string

	IssuesEvent func(*github.IssuesEvent) error
	PushEvent   func(*github.PushEvent) error

	// supportedEvents is populated automatically based on the values in the specific event
	// functions above on the first PingEvent request.
	supportedEvents []string
}

func (h *Handler) initSupportedEvents() {
	if len(h.supportedEvents) != 0 {
		// run initialization only once.
		return
	}
	if h.IssuesEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "issues")
	}
	if h.PushEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "push")
	}
}

// Request responds and routes events according ot functions defined in the handler.
// If no handler is defined, then
func (h *Handler) Request(w http.ResponseWriter, r *http.Request) {
	// Restrict event handling to POST requests.
	if r.Method != http.MethodPost {
		httpError(w, "Unsupported event type", http.StatusMethodNotAllowed)
		return
	}
	// webhookSecret should match the secret used when registering the webhook.
	payload, err := github.ValidatePayload(r, []byte(h.WebhookSecret))
	if err != nil {
		httpError(w, "Payload did not validate", http.StatusInternalServerError)
		return
	}
	log.Println("--")
	log.Println("Handling request for:", github.WebHookType(r))
	// Convert the payload into a specific github event type.
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println(string(payload))
		httpError(w, "Failed to parse webhook", http.StatusInternalServerError)
		return
	}

	// Check for the PingEvent type to handle differently than all other events.
	if event, ok := event.(*github.PingEvent); ok {
		if len(h.supportedEvents) == 0 {
			h.initSupportedEvents()
		}
		if !pingEventsSupported(event, h.supportedEvents) {
			msg := fmt.Sprintln("Unsupported event type:", event.Hook.Events, r.Header)
			httpError(w, msg, http.StatusNotImplemented)
		}
		return
	}

	if len(debug) != 0 && debug == "1" {
		log.Println(pretty.Sprint(event))
	}

	// Route all other event types to the corresponding handler function.
	switch event := event.(type) {
	case *github.PushEvent:
		err = h.PushEvent(event)
	case *github.IssuesEvent:
		err = h.IssuesEvent(event)
	case *github.InstallationRepositoriesEvent:
		pretty.Print(event)
	default:
		err = fmt.Errorf("Unsupported event type: %s", pretty.Sprint(event))
	}
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func pingEventsSupported(event *github.PingEvent, supported []string) bool {
	// Ping events occur during first registration.
	// If we return without error, the webhook is registered successfully.
	for i := range event.Hook.Events {
		if !slice.ContainsString(supported, event.Hook.Events[i]) {
			return false
		}
	}
	return true
}

// httpError both logs the gien message and writes an error to the given response writer.
func httpError(w http.ResponseWriter, msg string, status int) {
	log.Println(msg)
	http.Error(w, msg, status)
}
