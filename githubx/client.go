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

// NewClient ...
func NewClient(installationID int64) *github.Client {
	if installationID == 0 {
		authToken, found := os.LookupEnv("GITHUB_AUTH_TOKEN")
		if !found {
			log.Println("WARNING: GITHUB_AUTH_TOKEN is not set.")
			return nil
		}
		return NewPersonalClient(authToken)
	}

	// GitHub Apps spe
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

// NewPersonalClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewPersonalClient(authToken string) *github.Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	return github.NewClient(oauth2.NewClient(ctx, tokenSource))
}

// NewAppClient - x
func NewAppClient(privateKey string, appID, installationID int64) *github.Client {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the integration ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewKeyFromFile(tr, (int)(appID), (int)(installationID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	return github.NewClient(&http.Client{Transport: itr})
}
