package issues

import (
	"context"
	"testing"

	"github.com/google/go-github/github"
)

type fakeIssues struct {
	*github.Issue
	*github.Response
	Error  error
	labels []*github.Label
}

func (f *fakeIssues) Edit(
	ctx context.Context, owner string, repo string, number int,
	issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	return f.Issue, f.Response, f.Error
}

func (f *fakeIssues) AddLabelsToIssue(
	ctx context.Context, owner string, repo string, number int,
	labels []string) ([]*github.Label, *github.Response, error) {
	return f.labels, f.Response, f.Error
}

func (f *fakeIssues) RemoveLabelForIssue(
	ctx context.Context, owner string, repo string, number int,
	label string) (*github.Response, error) {
	return f.Response, f.Error
}

func newLabel(label string) github.Label {
	return github.Label{
		Name: &label,
	}
}

func TestNewEvent(t *testing.T) {
	_ = NewEvent((&github.Client{}).Issues, &github.IssuesEvent{})
}

func TestEvent_AddIssueLabels(t *testing.T) {
	tests := []struct {
		name        string
		IssuesEvent *github.IssuesEvent
		Issues      *fakeIssues
		ctx         context.Context
		labels      []string
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
			labels: []string{"okay2"},
		},
		/*{
			name: "okay-no-action",
			IssuesEvent: &github.IssuesEvent{
				Issue: &github.Issue{
					Labels: []github.Label{
						newLabel("okay"),
					},
				},
			},
			labels: []string{"okay"},
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				IssuesEvent: tt.IssuesEvent,
				Issues:      tt.Issues,
			}
			_, _, err := ev.AddIssueLabels(tt.ctx, tt.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.AddIssueLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if tt.Issues != nil && !reflect.DeepEqual(issue, tt.Issues.Issue) {
			//t.Errorf("Event.AddIssueLabels() got = %v, want %v", issue, tt.Issues.Issue)
			//}
		})
	}
}

func TestEvent_RemoveIssueLabels(t *testing.T) {
	tests := []struct {
		name        string
		IssuesEvent *github.IssuesEvent
		Issues      *fakeIssues
		ctx         context.Context
		labels      []string
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
			labels: []string{"okay"},
		},
		/*
			{
				name: "okay-already-missing",
				IssuesEvent: &github.IssuesEvent{
					Issue: &github.Issue{
						Labels: []github.Label{
							newLabel("foo"),
						},
					},
				},
				labels: []string{"okay"},
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &Event{
				IssuesEvent: tt.IssuesEvent,
				Issues:      tt.Issues,
			}
			_, err := ev.RemoveIssueLabels(tt.ctx, tt.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.RemoveIssueLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if tt.Issues != nil && !reflect.DeepEqual(issue, tt.Issues.Issue) {
			//t.Errorf("Event.RemoveIssueLabel() got = %v, want %v", issue, tt.Issues.Issue)
			//}
		})
	}
}
