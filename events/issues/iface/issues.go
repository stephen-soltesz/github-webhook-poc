package iface

import (
	"context"

	"github.com/google/go-github/github"
)

// Issues defines the interface used by the issues event logic.
type Issues interface {
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
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
