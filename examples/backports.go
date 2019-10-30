package main

import (
	"fmt"
	"time"

	"github.com/rubiojr/ghtools/backports"
)

func main() {
	// 15 days ago
	opts := backports.ListOpts{Since: time.Now().AddDate(0, 0, -15).Format("2006-01-02")}
	res, err := backports.ListGroupedBackports("orgbar", "teamfoo", opts)
	if err != nil {
		panic(err)
	}
	for k, v := range res {
		fmt.Printf("%s\n", k)
		for _, issue := range v {
			if issue.State == "closed" {
				fmt.Printf("  🔴  %s: %s\n", issue.Version, issue.URL)
			} else if issue.State == "open" {
				fmt.Printf("  🐣  %s: %s\n", issue.Version, issue.URL)
			} else if issue.State == "merged" {
				fmt.Printf("  ✅  %s: %s\n", issue.Version, issue.URL)
			}
		}
	}
}
