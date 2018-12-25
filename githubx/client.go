package githubx

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// NewClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewClient(authToken string) *github.Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	return github.NewClient(oauth2.NewClient(ctx, tokenSource))
}
