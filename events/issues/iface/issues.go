package iface

import (
	"context"

	"github.com/google/go-github/github"
)

// Issues defines the interface used by the issues event logic.
type Issues interface {
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
	AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, *github.Response, error)
	RemoveLabelForIssue(ctx context.Context, owner string, repo string, number int, label string) (*github.Response, error)
}

// IssuesImpl implements the Issues interface.
type IssuesImpl struct {
	*github.IssuesService
}

// NewIssues creates a new Issues instance.
func NewIssues(service *github.IssuesService) *IssuesImpl {
	return &IssuesImpl{service}
}

// Edit an issue.
func (i *IssuesImpl) Edit(
	ctx context.Context, owner string, repo string, number int,
	issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	return i.IssuesService.Edit(ctx, owner, repo, number, issue)
}

// AddLabelsToIssue adds the given lables to the issue number in the owner repo.
func (i *IssuesImpl) AddLabelsToIssue(
	ctx context.Context, owner string, repo string, number int,
	labels []string) ([]*github.Label, *github.Response, error) {
	return i.IssuesService.AddLabelsToIssue(ctx, owner, repo, number, labels)
}

// RemoveLabelForIssue removes the given label from the repo issue.
func (i *IssuesImpl) RemoveLabelForIssue(
	ctx context.Context, owner string, repo string, number int,
	label string) (*github.Response, error) {
	return i.IssuesService.RemoveLabelForIssue(ctx, owner, repo, number, label)
}
