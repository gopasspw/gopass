package main

import (
	"fmt"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/dominikschulz/github-releases/ghrel"
)

var args struct {
	User    string `arg:"required"`
	Project string `arg:"required"`
	Version string `arg:""`
	URL     bool   `arg:""`
}

func main() {
	arg.MustParse(&args)

	r, err := ghrel.FetchLatestStableRelease(args.User, args.Project)
	if err != nil {
		fmt.Printf("Failed to fetch releases for %s/%s: %s", args.User, args.Project, err)
		os.Exit(1)
	}

	if len(args.Version) < 1 {
		fmt.Println(r.Name)
		os.Exit(0)
	}
	args.Version = strings.TrimPrefix(args.Version, "v")
	r.Name = strings.TrimPrefix(r.Name, "v")
	if r.Name != args.Version {
		fmt.Printf("Not latest. Your Version %s - Latest: %s\n", args.Version, r.Name)
		if len(r.Assets) > 0 && args.URL {
			fmt.Printf("URL: %s\n", r.Assets[0].URL)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
