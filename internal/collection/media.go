package collection

import (
	"fmt"
)

const domainFolder = "https://collectablesquids.folder.run"
const domainFirebase = ""

var defaultMediaURLProvider MediaURLProvider = FolderURLProvider{}

// This pattern allows me to supply my own inputs during testing
type MediaURLProvider interface {
	GetImageURL(*Collectable) (string, error)
}

type FolderURLProvider struct{}
type FirebaseURLProvider struct{}

func (f FolderURLProvider) GetImageURL(c *Collectable) (string, error) {
	return fmt.Sprintf("%s/%s", domainFolder, c.Filename), nil
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
