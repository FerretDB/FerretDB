// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func isPackageToBeClean(p *github.PackageVersion) bool {
	daysBack := 90
	toBeClean := false

	if time.Now().After(p.UpdatedAt.Add(time.Duration(daysBack) * 24 * time.Hour)) {
		log.Printf("Stale version: %v (%v, %s)", p.GetID(), p.GetVersion(), p.UpdatedAt)
		toBeClean = true
	} else {
		log.Printf("skip version: %v (%v, %s)", p.GetID(), p.GetVersion(), p.UpdatedAt)
	}

	return toBeClean
}

func main() {
	ctx := context.Background()
	tokenName := "ROBOT_TOKEN"
	token := os.Getenv(tokenName)
	if token == "" {
		log.Fatalf("env variable %v is not found, please set it before run it", tokenName)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	packageType := "container"
	packageName := "ferretdb-dev"
	orgName := "FerretDB"
	pageSize := 100
	pageIndex := 1
	var versions []string
	for {
		listOption := github.ListOptions{Page: pageIndex, PerPage: pageSize}
		packageListOption := github.PackageListOptions{ListOptions: listOption}
		packages, _, err := client.Organizations.PackageGetAllVersions(ctx, orgName, packageType, packageName, &packageListOption)
		if err != nil {
			log.Printf("Failed to get versions for page %d", pageIndex)
		}
		for _, v := range packages {
			if isPackageToBeClean(v) {
				versions = append(versions, fmt.Sprintf("%v", v.GetID()))
			}
		}
		if len(packages) < pageSize {
			log.Printf("Come to the last page (page %d)", pageIndex)
			break
		}
		pageIndex++
	}
	log.Println(versions)
	staleVersions := strings.Join(versions, ", ")
	log.Printf("STALE_VERSIONS is: %q.", staleVersions)
}
