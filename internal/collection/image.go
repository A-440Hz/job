package collection

import (
	"fmt"
)

const domainFolder = "https://collectablesquids.folder.run"
const domainFirebase = ""

var defaultImageURLProvider ImageURLProvider = FolderImageURLProvider{}

// This pattern allows me to supply my own inputs during testing
type ImageURLProvider interface {
	GetImageURL(*Collectable) (string, error)
}

type FolderImageURLProvider struct{}
type FirebaseImageURLProvider struct{}

func (f FolderImageURLProvider) GetImageURL(c *Collectable) (string, error) {
	return fmt.Sprintf("%s/%s", domainFolder, c.Filename), nil
}

func (f FirebaseImageURLProvider) GetImageURL(c *Collectable) (string, error) {
	// use firebase package and return a SignedURL with expiry
	return domainFirebase, nil
}

func GetImageURL(c *Collectable) (string, error) {
	// options:
	// log error &&...
	// return a blank/fallback url
	// retry fetching a few times before failing
	// trigger an HTTP Internal error or 404
	return defaultImageURLProvider.GetImageURL(c)
}
