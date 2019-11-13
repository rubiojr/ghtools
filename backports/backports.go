package backports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/rubiojr/ghtools/client"
)

// Backport represents a backported PR
type Backport struct {
	Version       string
	VersionTitle  string
	Title         string
	State         string
	URL           string
	IssueNumber   int
	ParentVersion string
	ParentURL     string
	CreatedAt     time.Time
}

// Backport states
const (
	Merged int = iota
	Open
	Any
)

// ListOpts are options passed to the different list functions
type ListOpts struct {
	Since     string // Date string formatted as Year-Month-Day (2006-01-02)
	OlderThan int    // List issues/PRs older than this (days)
}

var defaultListOpts = ListOpts{
	Since:     time.Now().AddDate(0, 0, -15).Format("2006-01-02"),
	OlderThan: 30,
}

// BackportGroup groups backport by PR title
type BackportGroup map[string][]*Backport

func searchGroupBackports(opts *github.SearchOptions, query, state string) (BackportGroup, error) {
	cl, _ := client.Singleton()
	groupped := BackportGroup{}

	for {
		sr, resp, err := cl.Search.Issues(context.Background(), query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range sr.Issues {
			if b, err := parseBackport(&issue); err == nil {
				if b.State == "closed" {
					b.State = state
				}
				if _, ok := groupped[b.Title]; ok {
					groupped[b.Title] = append(groupped[b.Title], b)
				} else {
					groupped[b.Title] = []*Backport{b}
				}
			} else {
				return nil, err
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return groupped, nil
}

// ListStale lists backports older than X days (15 by default)
//
// Configure the time threshold passing a ListOpts argument like:
//
//     ListStale("fooorg", "barteam", ListOpts{OlderThan: 60})
//
func ListStale(org, team string, lopts ListOpts) ([]*Backport, error) {
	if lopts == (ListOpts{}) {
		lopts = defaultListOpts
	}
	opts := &github.SearchOptions{Sort: "created", Order: "desc"}
	opts.ListOptions.PerPage = 100

	createdBefore := time.Now().AddDate(0, 0, -lopts.OlderThan).Format("2006-01-02")
	baseQuery := fmt.Sprintf(
		"org:%s team:%s/%s is:open created:<=%s is:pr in:title Backport",
		org, org, team, createdBefore,
	)

	list, err := searchBackports(opts, baseQuery)

	return list, err
}

// ListGroupedBackports groups backports by PR
//
// All the backports from a PR will be added to the same map key
func ListGroupedBackports(org, team string, lopts ListOpts) (BackportGroup, error) {
	if lopts == (ListOpts{}) {
		lopts = defaultListOpts
	}
	opts := &github.SearchOptions{Sort: "created", Order: "asc"}
	opts.ListOptions.PerPage = 100

	baseQuery := fmt.Sprintf("org:%s team:%s/%s created:>=%s is:pr in:title Backport", org, org, team, lopts.Since)

	m := BackportGroup{}
	appendToExisting := func(res BackportGroup) {
		for k, v := range res {
			if _, ok := m[k]; ok {
				m[k] = append(m[k], v...)
			} else {
				m[k] = v
			}
		}
	}

	res, err := searchGroupBackports(opts, fmt.Sprintf("%s is:open", baseQuery), "open")
	if err != nil {
		return nil, err
	}
	appendToExisting(res)

	res, err = searchGroupBackports(opts, fmt.Sprintf("%s is:merged", baseQuery), "merged")
	if err != nil {
		return nil, err
	}
	appendToExisting(res)

	res, err = searchGroupBackports(opts, fmt.Sprintf("%s is:closed is:unmerged", baseQuery), "closed")
	if err != nil {
		return nil, err
	}
	appendToExisting(res)

	return m, nil
}

func parseBackport(issue *github.Issue) (*Backport, error) {
	state := *issue.State
	title := *issue.Title
	tokens := strings.Split(title, " ")
	parentVersion := tokens[1]
	repoUrlTokens := strings.Split(issue.GetRepositoryURL(), "/")
	nwo := fmt.Sprintf("%s/%s", repoUrlTokens[len(repoUrlTokens)-2], repoUrlTokens[len(repoUrlTokens)-1])
	parentURL := fmt.Sprintf("https://github.com/%s/pull/%s", nwo, parentVersion)
	version := strings.Trim(tokens[3], ":")
	t := strings.Split(title, ":")
	ft := strings.TrimSpace(strings.Join(t[1:len(t)], ":"))
	if len(t) > 1 {
		return &Backport{
			ParentURL:     parentURL,
			ParentVersion: parentVersion,
			Version:       version,
			VersionTitle:  title,
			State:         state,
			Title:         ft,
			URL:           *issue.HTMLURL,
			IssueNumber:   *issue.Number,
			CreatedAt:     *issue.CreatedAt}, nil
	} else {
		return nil, fmt.Errorf("Error parsing backport %s", title)
	}
}

func searchBackports(opts *github.SearchOptions, query string) ([]*Backport, error) {
	cl, _ := client.Singleton()
	list := []*Backport{}

	for {
		sr, resp, err := cl.Search.Issues(context.Background(), query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range sr.Issues {
			if b, err := parseBackport(&issue); err == nil {
				list = append(list, b)
			} else {
				return nil, err
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return list, nil
}
