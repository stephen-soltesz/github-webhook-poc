// Package webhook defines a generic handler for GitHub App & Webhook events.
package webhook

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/pretty"
)

var (
	enableDebugLogging string

	// Copied from go-github/github/messages.go due to being a private variable.
	eventTypeMapping = map[string]string{
		"check_run":                      "CheckRunEvent",
		"check_suite":                    "CheckSuiteEvent",
		"commit_comment":                 "CommitCommentEvent",
		"create":                         "CreateEvent",
		"delete":                         "DeleteEvent",
		"deployment":                     "DeploymentEvent",
		"deployment_status":              "DeploymentStatusEvent",
		"fork":                           "ForkEvent",
		"gollum":                         "GollumEvent",
		"installation":                   "InstallationEvent",
		"installation_repositories":      "InstallationRepositoriesEvent",
		"issue_comment":                  "IssueCommentEvent",
		"issues":                         "IssuesEvent",
		"label":                          "LabelEvent",
		"marketplace_purchase":           "MarketplacePurchaseEvent",
		"member":                         "MemberEvent",
		"membership":                     "MembershipEvent",
		"milestone":                      "MilestoneEvent",
		"organization":                   "OrganizationEvent",
		"org_block":                      "OrgBlockEvent",
		"page_build":                     "PageBuildEvent",
		"ping":                           "PingEvent",
		"project":                        "ProjectEvent",
		"project_card":                   "ProjectCardEvent",
		"project_column":                 "ProjectColumnEvent",
		"public":                         "PublicEvent",
		"pull_request_review":            "PullRequestReviewEvent",
		"pull_request_review_comment":    "PullRequestReviewCommentEvent",
		"pull_request":                   "PullRequestEvent",
		"push":                           "PushEvent",
		"repository":                     "RepositoryEvent",
		"repository_vulnerability_alert": "RepositoryVulnerabilityAlertEvent",
		"release":                        "ReleaseEvent",
		"status":                         "StatusEvent",
		"team":                           "TeamEvent",
		"team_add":                       "TeamAddEvent",
		"watch":                          "WatchEvent",
	}
)

func init() {
	enableDebugLogging = os.Getenv("DEBUG_LOGGING")
}

// A Handler defines parameters for implementing a GitHub webhook event handler
// as part of a GitHub App (https://developer.github.com/apps/) or ad-hoc
// Webhook server (https://developer.github.com/webhooks/).
//
// To be useful, define at least one event handler function.
type Handler struct {

	// WebhookSecret should match the value used to register the github webhook.
	WebhookSecret string

	// All functions accept the corresponding event type. The functions may
	// return an error. The `ServeHTTP` handler reports errors as HTTP 500
	// failures to the caller.
	CheckRunEvent                     func(*github.CheckRunEvent) error
	CheckSuiteEvent                   func(*github.CheckSuiteEvent) error
	CommitCommentEvent                func(*github.CommitCommentEvent) error
	CreateEvent                       func(*github.CreateEvent) error
	DeleteEvent                       func(*github.DeleteEvent) error
	DeploymentEvent                   func(*github.DeploymentEvent) error
	DeploymentStatusEvent             func(*github.DeploymentStatusEvent) error
	ForkEvent                         func(*github.ForkEvent) error
	GitHubAppAuthorizationEvent       func(*github.GitHubAppAuthorizationEvent) error
	GollumEvent                       func(*github.GollumEvent) error
	InstallationEvent                 func(*github.InstallationEvent) error
	InstallationRepositoriesEvent     func(*github.InstallationRepositoriesEvent) error
	IssueCommentEvent                 func(*github.IssueCommentEvent) error
	IssuesEvent                       func(*github.IssuesEvent) error
	LabelEvent                        func(*github.LabelEvent) error
	MarketplacePurchaseEvent          func(*github.MarketplacePurchaseEvent) error
	MemberEvent                       func(*github.MemberEvent) error
	MembershipEvent                   func(*github.MembershipEvent) error
	MilestoneEvent                    func(*github.MilestoneEvent) error
	OrganizationEvent                 func(*github.OrganizationEvent) error
	OrgBlockEvent                     func(*github.OrgBlockEvent) error
	PageBuildEvent                    func(*github.PageBuildEvent) error
	ProjectEvent                      func(*github.ProjectEvent) error
	ProjectCardEvent                  func(*github.ProjectCardEvent) error
	ProjectColumnEvent                func(*github.ProjectColumnEvent) error
	PublicEvent                       func(*github.PublicEvent) error
	PullRequestEvent                  func(*github.PullRequestEvent) error
	PullRequestReviewEvent            func(*github.PullRequestReviewEvent) error
	PullRequestReviewCommentEvent     func(*github.PullRequestReviewCommentEvent) error
	PushEvent                         func(*github.PushEvent) error
	ReleaseEvent                      func(*github.ReleaseEvent) error
	RepositoryEvent                   func(*github.RepositoryEvent) error
	RepositoryVulnerabilityAlertEvent func(*github.RepositoryVulnerabilityAlertEvent) error
	StatusEvent                       func(*github.StatusEvent) error
	TeamEvent                         func(*github.TeamEvent) error
	TeamAddEvent                      func(*github.TeamAddEvent) error
	WatchEvent                        func(*github.WatchEvent) error

	// supportedEvents is populated automatically based on the values in the specific event
	// functions above on the first PingEvent request.
	supportedEvents []string
}

// ServeHTTP is an http.Handler used to respond to webhook events with the
// corresponding user-provided functions from the Handler.
//
// The GitHub "ping" event is handled automatically. For every event type
// received in a "ping" event, ServeHTTP checks whether there is corresponding,
// non-nil event handler functions. If so, then the ping is successful; if not,
// the ping fails.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if github.WebHookType(r) == "integration_installation" ||
		github.WebHookType(r) == "integration_installation_repositories" {
		// The prefix "integration" is now deprecated and replaced by the short
		// names. Until the old names are removed, both old and new events are
		// delivered. This logic simply ignores the legacy names.
		log.Println("Ignoring legacy event name:", github.WebHookType(r))
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
		//if len(h.supportedEvents) == 0 {
		//h.initSupportedEvents()
		//}
		log.Println("Zen:", event.GetZen())
		if !allEventsSupported(h, event) {
			// if !pingEventsSupported(event, h.supportedEvents) {
			msg := fmt.Sprintln("Unsupported event type:", event.Hook.Events, r.Header)
			httpError(w, msg, http.StatusNotImplemented)
		} else {
			log.Printf("Successful ping for all events: %v", event.Hook.Events)
		}
		return
	}

	if len(enableDebugLogging) != 0 && enableDebugLogging == "1" {
		log.Println(pretty.Sprint(event))
	}

	// Get the event type name.
	eventType := reflect.TypeOf(event).Elem().Name()
	// Get a direct reference to the current Handler.
	rHandler := reflect.Indirect(reflect.ValueOf(h))
	// Lookup the field in the Handler with the same event type name.
	rHandlerFunc := rHandler.FieldByName(eventType)
	if reflect.DeepEqual(rHandlerFunc, reflect.Value{}) {
		// We've received an event for an unknown event type.
		httpError(w, "Unknown event type: "+eventType, http.StatusNotImplemented)
		return
	}
	if rHandlerFunc.IsNil() {
		// We've received an event for a known event type, but it is undefined.
		// Normally, a "ping" event would discover this and the handler would fail to
		// register. However, it's possible for the set of events to change across
		// deployments after a successful "ping" event.
		httpError(w, "Unimplemented handler func for: "+eventType, http.StatusNotImplemented)
		return
	}

	log.Printf("Calling handler for %q", eventType)
	// Call the handler function with current event.
	args := []reflect.Value{reflect.ValueOf(event)}
	ret := rHandlerFunc.Call(args)
	err = nil
	// Handler functions always return an error.
	if len(ret) > 0 && !ret[0].IsNil() {
		err = ret[0].Interface().(error)
	}
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func allEventsSupported(h *Handler, event *github.PingEvent) bool {
	// Ping events occur during webhook registration.
	// If we return true, the webhook is registered successfully.
	for i := range event.Hook.Events {
		eventName, ok := eventTypeMapping[event.Hook.Events[i]]
		if !ok {
			log.Printf("Unrecognized event type! %q", event.Hook.Events[i])
			return false
		}
		// Lookup eventName in the handler struct.
		rHandler := reflect.Indirect(reflect.ValueOf(h))
		handlerField := rHandler.FieldByName(eventName)
		if reflect.DeepEqual(handlerField, reflect.Value{}) {
			// That field does not exist in the header struct.
			return false
		}
		if handlerField.IsNil() {
			// The field exists, but it has a nil value.
			return false
		}
	}
	// All fields exist and have a non-nil value.
	return true
}

// httpError both logs the gien message and writes an error to the given response writer.
func httpError(w http.ResponseWriter, msg string, status int) {
	log.Println(msg)
	http.Error(w, msg, status)
}
