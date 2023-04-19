package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type FileSlug struct {
	slug     string
	fileName string
}

func main() {
    dir := filepath.Join("website","blog")
	fs, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	slugs := GetBlogSlugs(fs)
    pass := true
	for _, slug := range slugs {

		fo, err := os.Open(filepath.Join(dir, slug.fileName))
		defer fo.Close()
        
		if err != nil {
			log.Fatalf("Couldn't open file: %s", slug.fileName)
			continue
		}

        serr := VerifySlug(slug, fo)

        if serr != nil {
            log.Print(serr)
            pass = false
        }
        
	}
    
    if !pass {
        log.Fatal("One or more blog posts are not correctly formated")
    }

}

func GetBlogSlugs(fs []fs.DirEntry) []FileSlug {
	// start with 4 digits then a 0[1-9] or 1 [0 1 or 2] then - and either 0 [1-9] or [1 or 2][0-9] or 3[0 or 1] - slug(any) and end with .md
	r := regexp.MustCompile(`^\d{4}\-(?:0[1-9]|1[012])\-(?:0[1-9]|[12][0-9]|3[01])-(.*).md$`)

	var fileSlugs []FileSlug
	for _, f := range fs {

		fn := f.Name()
		m, err := regexp.MatchString(`.md$`, fn)

		if !m {
			continue
		}

		if err != nil {
			log.Fatalf("regexp error, %s", err)
		}

		sm := r.FindStringSubmatch(fn)
		if len(sm) > 2 {
			log.Fatalf("File %s is not correctly formated (yyyy-mm-dd-'slug'.md)", fn)
			continue
		}

		fileSlugs = append(fileSlugs, FileSlug{sm[len(sm)-1], fn})
	}

	return fileSlugs
}

func VerifySlug(fS FileSlug, f io.Reader) error {
	r := regexp.MustCompile("^slug: (.*)")

	pass := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sm := r.FindStringSubmatch(scanner.Text())
		if len(sm) > 1 && sm[len(sm)-1] == fS.slug {
			pass = true
		}
	}

	if !pass {
        return fmt.Errorf("Slug is not correctly formated in file %s",fS.fileName)
	}

    return nil
}
