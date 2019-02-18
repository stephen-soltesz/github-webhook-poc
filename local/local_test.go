package local

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/stephen-soltesz/github-webhook-poc/events/issues/iface"
)

type fakeIssues struct {
	*github.Issue
	*github.Response
	Error error
}

func (f *fakeIssues) Edit(
	ctx context.Context, owner string, repo string, number int,
	issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	return f.Issue, f.Response, f.Error
}

func newLabel(label string) github.Label {
	return github.Label{
		Name: &label,
	}
}

func newString(s string) *string {
	return &s
}

func TestNewConfig(t *testing.T) {
	_ = NewConfig(time.Second)
}

func TestConfig_IssuesEvent(t *testing.T) {
	tests := []struct {
		name            string
		githubAuthToken string
		issue           *github.Issue
		event           *github.IssuesEvent
		wantErr         bool
		wantEventErr    bool
	}{
		{
			name:            "successful-add",
			githubAuthToken: "test",
			issue:           &github.Issue{Labels: []github.Label{newLabel("okay")}},
			event:           &github.IssuesEvent{Action: newString("opened")},
		},
		{
			name:            "successful-remove",
			githubAuthToken: "test",
			issue:           &github.Issue{Labels: []github.Label{newLabel("okay")}},
			event:           &github.IssuesEvent{Action: newString("closed")},
		},
		{
			name:            "successful-ignore-default",
			githubAuthToken: "test",
			issue:           &github.Issue{Labels: []github.Label{newLabel("okay")}},
			event:           &github.IssuesEvent{Action: newString("unsupported-action")},
		},
		{
			name:    "error-client",
			wantErr: true,
		},
		{
			name:            "error-add",
			githubAuthToken: "test",
			issue:           &github.Issue{Labels: []github.Label{newLabel("okay")}},
			event:           &github.IssuesEvent{Action: newString("opened")},
			wantEventErr:    true,
		},
	}
	for _, tt := range tests {
		getIface := func(client *github.Client) iface.Issues {
			if tt.wantEventErr {
				return &fakeIssues{
					Issue: tt.issue,
					Error: fmt.Errorf("fake error"),
				}
			} else {
				return &fakeIssues{
					Issue: tt.issue,
				}
			}
		}
		if tt.event != nil {
			tt.event.Issue = tt.issue
		}
		if tt.githubAuthToken != "" {
			os.Setenv("GITHUB_AUTH_TOKEN", tt.githubAuthToken)
		} else {
			os.Unsetenv("GITHUB_AUTH_TOKEN")
		}
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Delay:    time.Millisecond,
				getIface: getIface,
			}
			if err := c.IssuesEvent(tt.event); (err != nil) != tt.wantErr {
				t.Errorf("Config.IssuesEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstallationEvent(t *testing.T) {
	event := &github.InstallationEvent{}
	_ = InstallationEvent(event)
}

func TestInstallationRepositoriesEvent(t *testing.T) {
	event := &github.InstallationRepositoriesEvent{}
	_ = InstallationRepositoriesEvent(event)
}

func newInt64(i int64) *int64 {
	return &i
}

func Test_getSafeID(t *testing.T) {
	event := &github.IssuesEvent{
		Installation: &github.Installation{
			ID: newInt64(1),
		},
	}
	_ = getSafeID(event)
}

func Test_getIface(t *testing.T) {
	_ = getIface(&github.Client{})
}

func Test_genSprintLabels(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want []string
	}{
		{
			name: "sprint 1 2019",
			now:  time.Date(2019, 01, 1, 0, 0, 0, 0, time.UTC),
			want: []string{"Sprint 1", "2019"},
		},
		{
			name: "sprint 4 2019",
			now:  time.Date(2019, 02, 17, 0, 0, 0, 0, time.UTC),
			want: []string{"Sprint 4", "2019"},
		},
		{
			name: "sprint 26 2019",
			now:  time.Date(2019, 12, 29, 0, 0, 0, 0, time.UTC),
			want: []string{"Sprint 26", "2019"},
		},
		{
			name: "sprint 1 2020",
			now:  time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC),
			want: []string{"Sprint 1", "2020"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genSprintLabels(tt.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("genSprintLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
