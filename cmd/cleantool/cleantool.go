package main

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "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"},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	packageType := "container"
	packageName := "ferretdb-dev"
	orgName := "FerretDB"
	daysBack := 90
	pageSize := 100
	var versions []int64
	for i := 1; i < 100; i++ {
		listOption := github.ListOptions{Page: i, PerPage: pageSize}
		packageListOption := github.PackageListOptions{ListOptions: listOption}
		packages, _, err := client.Organizations.PackageGetAllVersions(ctx, orgName, packageType, packageName, &packageListOption)
		if err != nil {
			log.Printf("Failed to get versions for page %d", i)
		}
		for _, v := range packages {
			if time.Now().After(v.UpdatedAt.Add(time.Duration(daysBack) * 24 * time.Hour)) {
				log.Printf("Stale version: %v (%v, %s)", v.GetID(), v.GetVersion(), v.UpdatedAt)
				versions = append(versions, v.GetID())
			} else {
				log.Printf("skip version: %v (%v, %s)", v.GetID(), v.GetVersion(), v.UpdatedAt)
			}
		}
		if len(packages) < pageSize {
			log.Printf("Come to the last page (page %d)", i)
			break
		}

	}
	log.Println(versions)
}
