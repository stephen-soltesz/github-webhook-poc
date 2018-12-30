// Package webhook defines a generic handler for GitHub App & Webhook events.
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
	enableDebugLogging string
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
		if len(h.supportedEvents) == 0 {
			h.initSupportedEvents()
		}
		log.Println("Zen:", event.GetZen())
		if !pingEventsSupported(event, h.supportedEvents) {
			msg := fmt.Sprintln("Unsupported event type:", event.Hook.Events, r.Header)
			httpError(w, msg, http.StatusNotImplemented)
		} else {
			log.Printf("Successful ping for events: %v", h.supportedEvents)
		}
		return
	}

	if len(enableDebugLogging) != 0 && enableDebugLogging == "1" {
		log.Println(pretty.Sprint(event))
	}

	// Handle all other event types using the corresponding event function.
	switch event := event.(type) {
	case *github.CheckRunEvent:
		err = h.CheckRunEvent(event)
	case *github.CheckSuiteEvent:
		err = h.CheckSuiteEvent(event)
	case *github.CommitCommentEvent:
		err = h.CommitCommentEvent(event)
	case *github.CreateEvent:
		err = h.CreateEvent(event)
	case *github.DeleteEvent:
		err = h.DeleteEvent(event)
	case *github.DeploymentEvent:
		err = h.DeploymentEvent(event)
	case *github.DeploymentStatusEvent:
		err = h.DeploymentStatusEvent(event)
	case *github.ForkEvent:
		err = h.ForkEvent(event)
	case *github.GitHubAppAuthorizationEvent:
		err = h.GitHubAppAuthorizationEvent(event)
	case *github.GollumEvent:
		err = h.GollumEvent(event)
	case *github.InstallationEvent:
		err = h.InstallationEvent(event)
	case *github.InstallationRepositoriesEvent:
		err = h.InstallationRepositoriesEvent(event)
	case *github.IssueCommentEvent:
		err = h.IssueCommentEvent(event)
	case *github.IssuesEvent:
		err = h.IssuesEvent(event)
	case *github.LabelEvent:
		err = h.LabelEvent(event)
	case *github.MarketplacePurchaseEvent:
		err = h.MarketplacePurchaseEvent(event)
	case *github.MemberEvent:
		err = h.MemberEvent(event)
	case *github.MembershipEvent:
		err = h.MembershipEvent(event)
	case *github.MilestoneEvent:
		err = h.MilestoneEvent(event)
	case *github.OrganizationEvent:
		err = h.OrganizationEvent(event)
	case *github.OrgBlockEvent:
		err = h.OrgBlockEvent(event)
	case *github.PageBuildEvent:
		err = h.PageBuildEvent(event)
	case *github.ProjectEvent:
		err = h.ProjectEvent(event)
	case *github.ProjectCardEvent:
		err = h.ProjectCardEvent(event)
	case *github.ProjectColumnEvent:
		err = h.ProjectColumnEvent(event)
	case *github.PublicEvent:
		err = h.PublicEvent(event)
	case *github.PullRequestEvent:
		err = h.PullRequestEvent(event)
	case *github.PullRequestReviewEvent:
		err = h.PullRequestReviewEvent(event)
	case *github.PullRequestReviewCommentEvent:
		err = h.PullRequestReviewCommentEvent(event)
	case *github.PushEvent:
		err = h.PushEvent(event)
	case *github.ReleaseEvent:
		err = h.ReleaseEvent(event)
	case *github.RepositoryEvent:
		err = h.RepositoryEvent(event)
	case *github.RepositoryVulnerabilityAlertEvent:
		err = h.RepositoryVulnerabilityAlertEvent(event)
	case *github.StatusEvent:
		err = h.StatusEvent(event)
	case *github.TeamEvent:
		err = h.TeamEvent(event)
	case *github.TeamAddEvent:
		err = h.TeamAddEvent(event)
	case *github.WatchEvent:
		err = h.WatchEvent(event)
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

func (h *Handler) initSupportedEvents() {
	if len(h.supportedEvents) != 0 {
		// run initialization only once.
		return
	}
	if h.CheckRunEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "check_run")
	}
	if h.CheckSuiteEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "check_suite")
	}
	if h.CommitCommentEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "commit_comment")
	}
	if h.CreateEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "create")
	}
	if h.DeleteEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "delete")
	}
	if h.DeploymentEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "deployment")
	}
	if h.DeploymentStatusEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "deployment_status")
	}
	if h.ForkEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "fork")
	}
	if h.GollumEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "gollum")
	}
	if h.InstallationEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "installation")
	}
	if h.InstallationRepositoriesEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "installation_repositories")
	}
	if h.IssueCommentEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "issue_comment")
	}
	if h.IssuesEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "issues")
	}
	if h.LabelEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "label")
	}
	if h.MarketplacePurchaseEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "marketplace_purchase")
	}
	if h.MemberEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "member")
	}
	if h.MembershipEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "membership")
	}
	if h.MilestoneEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "milestone")
	}
	if h.OrganizationEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "organization")
	}
	if h.OrgBlockEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "org_block")
	}
	if h.PageBuildEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "page_build")
	}
	if h.ProjectEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "project")
	}
	if h.ProjectCardEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "project_card")
	}
	if h.ProjectColumnEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "project_column")
	}
	if h.PublicEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "public")
	}
	if h.PullRequestReviewEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "pull_request_review")
	}
	if h.PullRequestReviewCommentEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "pull_request_review_comment")
	}
	if h.PullRequestEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "pull_request")
	}
	if h.PushEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "push")
	}
	if h.RepositoryEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "repository")
	}
	if h.RepositoryVulnerabilityAlertEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "repository_vulnerability_alert")
	}
	if h.ReleaseEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "release")
	}
	if h.StatusEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "status")
	}
	if h.TeamEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "team")
	}
	if h.TeamAddEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "team_add")
	}
	if h.WatchEvent != nil {
		h.supportedEvents = append(h.supportedEvents, "watch")
	}
}
