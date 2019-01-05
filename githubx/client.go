// Package githubx extends operations in the google/go-github package.
package githubx

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// NewClient creates a new, authenticated *github.Client. Depending on the value
// of installationID, NewClient performs authentication using available
// environment variables. If required environment variables are not found,
// NewClient logs a warning and returns nil.
//
// If the installationID is zero, then NewClient authenticates using a personal
// access token from the GITHUB_AUTH_TOKEN environment vairable. Personal access
// tokens perform actions as the user associated with the token.
//
// If the installationID is not zero, then NewClient authenticates using a
// Github App private key read from the absolute path contained in
// GITHUB_PRIVATE_KEY and the Github App ID (About -> ID, *not* "OAuth Client
// ID") contained in the GITHUB_APP_ID environment variables. Both values are
// created during Github App registration. Github Apps perform actions as the
// application not as a user, so permissions may have different restrictions.
//
// Multiple users may install a single Github App, so each installation receives
// a unique "installation ID". All events related to a specific installation
// include the installation ID.
func NewClient(installationID int64) *github.Client {

	// Personal access token authentication.
	if installationID == 0 {
		authToken, found := os.LookupEnv("GITHUB_AUTH_TOKEN")
		if !found {
			log.Println("WARNING: GITHUB_AUTH_TOKEN is not set.")
			return nil
		}
		return NewPersonalClient(authToken)
	}

	// GitHub Apps authentication.
	privateKey, found := os.LookupEnv("GITHUB_PRIVATE_KEY")
	if !found {
		log.Println("WARNING: GITHUB_PRIVATE_KEY is not set.")
		return nil
	}
	appIDStr, found := os.LookupEnv("GITHUB_APP_ID")
	if !found {
		log.Println("WARNING: GITHUB_APP_ID is not set.")
		return nil
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 32)
	if err != nil {
		log.Println(err)
		return nil
	}
	return NewAppClient(privateKey, appID, installationID)
}

// NewPersonalClient creates a *github.Client authenticated using the given
// Github authToken. Future operations are performed as the user associated with
// the token.
func NewPersonalClient(authToken string) *github.Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	return github.NewClient(oauth2.NewClient(ctx, tokenSource))
}

// NewAppClient creates a new *github.Client authenticated using the given
// privateKey file name, appID, and installationID. Future operations are
// performed as the Github App associated with these parameters.
func NewAppClient(privateKey string, appID, installationID int64) *github.Client {
	// Create a new "installation" (a.k.a. "Github Apps") transport that
	// authenticates using the given private key for the given app and installation
	// IDs.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, (int)(appID), (int)(installationID), privateKey)
	if err != nil {
		log.Println(err)
		return nil
	}
	// Use the installation transport with a new *github.Client.
	return github.NewClient(&http.Client{Transport: itr})
}
