package collection

import (
	"fmt"
)

const domainGithub = "https://raw.githubusercontent.com/A-440Hz/squids/main/"
const domainFirebase = ""

var defaultMediaURLProvider MediaURLProvider = GithubURLProvider{}

// This pattern allows me to supply my own inputs during testing
type MediaURLProvider interface {
	GetImageURL(*Collectable) (string, error)
}

type GithubURLProvider struct{}
type FirebaseURLProvider struct{}

func (f GithubURLProvider) GetImageURL(c *Collectable) (string, error) {
	return fmt.Sprintf("%s/%s", domainGithub, c.Filename), nil
}

func (f FirebaseURLProvider) GetImageURL(c *Collectable) (string, error) {
	// use firebase package and return a SignedURL with expiry
	return domainFirebase, nil
}

func GetImageURL(c *Collectable) (string, error) {
	// options:
	// log error &&...
	// return a blank/fallback url
	// retry fetching a few times before failing
	// trigger an HTTP Internal error or 404
	return defaultMediaURLProvider.GetImageURL(c)
}
