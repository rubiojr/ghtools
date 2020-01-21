// Package client provides helpers to initialize a GitHub client.
package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

var ghcInstance *github.Client
var ghcOnce sync.Once

// Singleton returns a GitHub client singleton.
//
// A GitHub personal access token is required.
//
// Singleton will try to read the token from the environment variable
// GITHUB_TOKEN or read it from the operating system keychain.
//
// To add the token to the macOS keychain you can use the command line
// utility "security" like this:
//
//   security add-generic-password -a github -s GITHUB_TOKEN -w
//
// To add the token to GNOME keyring use "secret-tool":
//
//   secret-tool store --label="GitHub Token" service GITHUB_TOKEN username github
func Singleton() (*github.Client, error) {
	var err error

	ghcOnce.Do(func() {
		creds := os.Getenv("GITHUB_TOKEN")
		if creds != "" {
			ghcInstance, err = newGHClientFromToken(creds)
		} else {
			creds, err = keyring.Get("GITHUB_TOKEN", "github")
			if err != nil {
				err = fmt.Errorf("GitHub token not found in keyring. E: %v", err)
			} else {
				ghcInstance, err = newGHClientFromToken(creds)
			}
		}
	})

	if err != nil {
		ghcInstance = nil
	}
	return ghcInstance, err
}

func newGHClientFromToken(token string) (*github.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token can't be empty")
	}
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}

func newGHClientFromFile(creds string) *github.Client {
	ctx := context.Background()

	key, err := ioutil.ReadFile(creds)
	if err != nil {
		panic(err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(key))},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
