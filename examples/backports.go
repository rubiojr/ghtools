package main

import (
	"fmt"

	"github.com/rubiojr/ghtools/backports"
)

func main() {
	res, err := backports.ListGroupedBackports("orgbar", "teamfoo")
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
