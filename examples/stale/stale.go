package main

import (
	"fmt"

	"github.com/rubiojr/ghtools/backports"
)

func main() {
	// 15 days ago
	opts := backports.ListOpts{OlderThan: 15}
	res, err := backports.ListStale("orgfoo", "barteam", opts)
	if err != nil {
		panic(err)
	}
	for _, v := range res {
		fmt.Printf("Title: %s\n", v.Title)
		fmt.Printf("  URL: %s\n", v.URL)
		fmt.Printf("  Created: %s\n", v.CreatedAt)
	}
}
