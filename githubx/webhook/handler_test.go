// Package webhook defines a generic handler for GitHub App & Webhook events.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/pretty"
)

func genMAC(message, key string, hashFunc func() hash.Hash) string {
	mac := hmac.New(hashFunc, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func newRequest(method, payload, key, event string) *http.Request {
	hash := genMAC(payload, key, sha1.New)
	r := httptest.NewRequest(method, "/", bytes.NewBufferString(payload))
	r.Header.Set("X-Github-Event", event)
	r.Header.Set("X-Hub-Signature", "sha1="+hash)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func mustReadAll(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		WebhookSecret                     string
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
		supportedEvents                   []string
	}
	tests := []struct {
		name         string
		fields       fields
		event        string
		payload      string
		method       string
		status       int
		breakWebhook bool
	}{
		{
			name: "ping",
			fields: fields{
				WebhookSecret: "test",
				PushEvent: func(event *github.PushEvent) error {
					pretty.Print(event)
					return nil
				},
			},
			event:   "ping",
			payload: mustReadAll("testdata/ping.json"),
			method:  http.MethodPost,
			status:  http.StatusOK,
		},
		{
			name: "ping: missing push function",
			fields: fields{
				WebhookSecret: "test",
			},
			event:   "ping",
			payload: mustReadAll("testdata/ping.json"),
			method:  http.MethodPost,
			status:  http.StatusNotImplemented,
		},
		{
			name: "ping: incorrect method",
			fields: fields{
				WebhookSecret: "test",
			},
			event:   "ping",
			payload: mustReadAll("testdata/ping.json"),
			method:  http.MethodGet,
			status:  http.StatusMethodNotAllowed,
		},
		{
			name: "push: returns error",
			fields: fields{
				WebhookSecret: "test",
				PushEvent: func(event *github.PushEvent) error {
					return fmt.Errorf("Return failure")
				},
			},
			event:   "push",
			payload: mustReadAll("testdata/empty.json"),
			method:  http.MethodPost,
			status:  http.StatusInternalServerError,
		},
		{
			name: "push: parse webhook failure",
			fields: fields{
				WebhookSecret: "test",
			},
			event:   "push",
			payload: mustReadAll("testdata/bad.json"),
			method:  http.MethodPost,
			status:  http.StatusInternalServerError,
		},
		{
			name: "ping: broken webhook secret",
			fields: fields{
				WebhookSecret: "test",
				PushEvent: func(event *github.PushEvent) error {
					return nil
				},
			},
			event:        "ping",
			payload:      mustReadAll("testdata/ping.json"),
			method:       http.MethodPost,
			status:       http.StatusInternalServerError,
			breakWebhook: true,
		},
		{
			name: "integration_installation: ignored",
			fields: fields{
				WebhookSecret: "test",
				PushEvent: func(event *github.PushEvent) error {
					return nil
				},
			},
			event:   "integration_installation",
			payload: mustReadAll("testdata/empty.json"),
			method:  http.MethodPost,
			status:  http.StatusOK,
		},
		{
			name: "issues",
			fields: fields{
				WebhookSecret: "test",
				IssuesEvent: func(event *github.IssuesEvent) error {
					return nil
				},
			},
			event:   "issues",
			payload: mustReadAll("testdata/issues.json"),
			method:  http.MethodPost,
			status:  http.StatusOK,
		},
		{
			name: "push: without handler function",
			fields: fields{
				WebhookSecret: "test",
				// PushEvent: not defined.
			},
			event:   "push",
			payload: mustReadAll("testdata/push.json"),
			method:  http.MethodPost,
			status:  http.StatusNotImplemented,
		},
		{
			name: "ping: unsupported event",
			fields: fields{
				WebhookSecret: "test",
			},
			event:   "ping",
			payload: mustReadAll("testdata/ping-unsupported.json"),
			method:  http.MethodPost,
			status:  http.StatusNotImplemented,
		},
	}
	for _, tt := range tests {
		var r *http.Request
		enableDebugLogging = "1"
		if tt.breakWebhook {
			r = newRequest(tt.method, tt.payload, "this is not the secret", tt.event)
		} else {
			r = newRequest(tt.method, tt.payload, tt.fields.WebhookSecret, tt.event)
		}
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				WebhookSecret:                     tt.fields.WebhookSecret,
				CheckRunEvent:                     tt.fields.CheckRunEvent,
				CheckSuiteEvent:                   tt.fields.CheckSuiteEvent,
				CommitCommentEvent:                tt.fields.CommitCommentEvent,
				CreateEvent:                       tt.fields.CreateEvent,
				DeleteEvent:                       tt.fields.DeleteEvent,
				DeploymentEvent:                   tt.fields.DeploymentEvent,
				DeploymentStatusEvent:             tt.fields.DeploymentStatusEvent,
				ForkEvent:                         tt.fields.ForkEvent,
				GitHubAppAuthorizationEvent:       tt.fields.GitHubAppAuthorizationEvent,
				GollumEvent:                       tt.fields.GollumEvent,
				InstallationEvent:                 tt.fields.InstallationEvent,
				InstallationRepositoriesEvent:     tt.fields.InstallationRepositoriesEvent,
				IssueCommentEvent:                 tt.fields.IssueCommentEvent,
				IssuesEvent:                       tt.fields.IssuesEvent,
				LabelEvent:                        tt.fields.LabelEvent,
				MarketplacePurchaseEvent:          tt.fields.MarketplacePurchaseEvent,
				MemberEvent:                       tt.fields.MemberEvent,
				MembershipEvent:                   tt.fields.MembershipEvent,
				MilestoneEvent:                    tt.fields.MilestoneEvent,
				OrganizationEvent:                 tt.fields.OrganizationEvent,
				OrgBlockEvent:                     tt.fields.OrgBlockEvent,
				PageBuildEvent:                    tt.fields.PageBuildEvent,
				ProjectEvent:                      tt.fields.ProjectEvent,
				ProjectCardEvent:                  tt.fields.ProjectCardEvent,
				ProjectColumnEvent:                tt.fields.ProjectColumnEvent,
				PublicEvent:                       tt.fields.PublicEvent,
				PullRequestEvent:                  tt.fields.PullRequestEvent,
				PullRequestReviewEvent:            tt.fields.PullRequestReviewEvent,
				PullRequestReviewCommentEvent:     tt.fields.PullRequestReviewCommentEvent,
				PushEvent:                         tt.fields.PushEvent,
				ReleaseEvent:                      tt.fields.ReleaseEvent,
				RepositoryEvent:                   tt.fields.RepositoryEvent,
				RepositoryVulnerabilityAlertEvent: tt.fields.RepositoryVulnerabilityAlertEvent,
				StatusEvent:                       tt.fields.StatusEvent,
				TeamEvent:                         tt.fields.TeamEvent,
				TeamAddEvent:                      tt.fields.TeamAddEvent,
				WatchEvent:                        tt.fields.WatchEvent,
				supportedEvents:                   tt.fields.supportedEvents,
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Errorf("wrong status got %v; want %v", w.Code, tt.status)
			}
		})
	}
}
