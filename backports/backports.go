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
	Version      string
	VersionTitle string
	Title        string
	State        string
	URL          string
}

// Backport states
const (
	Merged int = iota
	Open
	Any
)

func searchBackports(opts *github.SearchOptions, query, state string) (map[string][]*Backport, error) {
	cl, _ := client.Singleton()
	groupped := map[string][]*Backport{}

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

// ListGroupedBackports groups backports by PR
//
// All the backports from a PR will be added to the same map key
func ListGroupedBackports(org, team string) (map[string][]*Backport, error) {
	since := time.Now().AddDate(0, 0, -15).Format("2006-01-02")
	opts := &github.SearchOptions{Sort: "created", Order: "asc"}
	opts.ListOptions.PerPage = 100

	baseQuery := fmt.Sprintf("team:%s/%s created:>=%s is:pr in:title Backport", org, team, since)

	query := fmt.Sprintf("%s is:open", baseQuery)
	m, err := searchBackports(opts, query, "open")

	query = fmt.Sprintf("%s is:merged", baseQuery)
	res, err := searchBackports(opts, query, "merged")
	if err == nil {
		for k, v := range res {
			if _, ok := m[k]; ok {
				m[k] = append(m[k], v...)
			} else {
				m[k] = v
			}
		}
	}

	query = fmt.Sprintf("%s is:closed is:unmerged", baseQuery)
	res, err = searchBackports(opts, query, "closed")
	if err == nil {
		for k, v := range res {
			if _, ok := m[k]; ok {
				m[k] = append(m[k], v...)
			} else {
				m[k] = v
			}
		}
	}

	return m, err
}

func parseBackport(issue *github.Issue) (*Backport, error) {
	state := *issue.State
	title := *issue.Title
	tokens := strings.Split(title, " ")
	version := strings.Trim(tokens[3], ":")
	t := strings.Split(title, ":")
	ft := strings.Join(t[1:len(t)], ":")
	if len(t) > 1 {
		return &Backport{Version: version, VersionTitle: title, State: state, Title: ft, URL: *issue.HTMLURL}, nil
	} else {
		return nil, fmt.Errorf("Error parsing backport %s", title)
	}
}
