package issues

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
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

func TestNewEvent(t *testing.T) {
	_ = NewEvent(&github.Client{}, &github.IssuesEvent{})
}

func TestEvent_AddIssueLabel(t *testing.T) {
	tests := []struct {
		name        string
		IssuesEvent *github.IssuesEvent
		Issues      *fakeIssues
		ctx         context.Context
		label       string
		wantErr     bool
	}{
		{
			name: "okay",
			IssuesEvent: &github.IssuesEvent{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("okay"),
					},
				},
			},
			Issues: &fakeIssues{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("okay2"),
					},
				},
			},
		},
		{
			name: "okay-no-action",
			IssuesEvent: &github.IssuesEvent{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("okay"),
					},
				},
			},
			label: "okay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				IssuesEvent: tt.IssuesEvent,
				Issues:      tt.Issues,
			}
			issue, _, err := ev.AddIssueLabel(tt.ctx, tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.AddIssueLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.Issues != nil && !reflect.DeepEqual(issue, tt.Issues.Issue) {
				t.Errorf("Event.AddIssueLabel() got = %v, want %v", issue, tt.Issues.Issue)
			}
		})
	}
}

func TestEvent_RemoveIssueLabel(t *testing.T) {
	tests := []struct {
		name        string
		IssuesEvent *github.IssuesEvent
		Issues      *fakeIssues
		ctx         context.Context
		label       string
		wantErr     bool
	}{
		{
			name: "okay",
			IssuesEvent: &github.IssuesEvent{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("okay"),
					},
				},
			},
			Issues: &fakeIssues{
				Issue: &github.Issue{
					Labels: []github.Label{},
				},
			},
			label: "okay",
		},
		{
			name: "okay-already-missing",
			IssuesEvent: &github.IssuesEvent{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("foo"),
					},
				},
			},
			label: "okay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				IssuesEvent: tt.IssuesEvent,
				Issues:      tt.Issues,
			}
			issue, _, err := ev.RemoveIssueLabel(tt.ctx, tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.RemoveIssueLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.Issues != nil && !reflect.DeepEqual(issue, tt.Issues.Issue) {
				t.Errorf("Event.RemoveIssueLabel() got = %v, want %v", issue, tt.Issues.Issue)
			}
		})
	}
}
