package filer

import (
	"os"
	"path/filepath"
	"strings"
)

var src string
var indexDst string
var dst string

func init() {
	if os.Getenv("ENV") == "testing" {
		src = "testdata/entries"
		indexDst = "testdata/docs"
	} else {
		src = "entries"
		indexDst = "docs"
	}
	dst = filepath.Join(indexDst, "blog")
}

func list(dir string) (map[uint]string, error) {
	results := make(map[uint]string)
	var id uint = 0

	files, err := os.ReadDir(filepath.Join(src, dir))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && strings.HasSuffix(name, ".md") {
			id++
			results[id] = filepath.Join(src, dir, name)
		}
	}

	return results, nil
}

// ListDrafts returns a list of entries in `draft/` and assigns an ID.
func ListDrafts() (map[uint]string, error) {
	return list("draft")
}

// ListPublished returns a list of entries in `published/` and assigns an ID.
func ListPublished() (map[uint]string, error) {
	return list("published")
}

// ListPages returns a list of published HTML pages.
func ListPages() ([]string, error) {
	pages := []string{}

	files, err := os.ReadDir(dst)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() &&
			strings.HasSuffix(name, ".html") &&
			name != "index.html" &&
			name != "about.html" {
			pages = append(pages, name)
		}
	}

	return pages, nil
}

func move(src, dst string) error {
	return os.Rename(src, dst)
}

// Publish moves an entry from `draft/` to `published/`
func Publish(from string) error {
	parts := strings.Split(from, "/")
	filename := parts[len(parts)-1]
	to := filepath.Join(src, "published", filename)
	return move(from, to)
}

// CreateFeed creates a `feed.xml` file
func CreateFeed() (*os.File, error) {
	return os.Create(filepath.Join(dst, "feed.xml"))
}

// CreateIndex creates `index.html`
func CreateIndex() (*os.File, error) {
	return os.Create(filepath.Join(indexDst, "index.html"))
}

// CreatePages creates an html file
func CreatePage(filename string) (*os.File, error) {
	return os.Create(filepath.Join(dst, filename+".html"))
}

// CreateDraft creates a .md draft file
func CreateDraft(filename string) (*os.File, error) {
	return os.Create(filepath.Join(src, "draft", filename))
}

// GetPublishedEntry retruns a .md file from the `published/` dir
func GetPublishedEntry(filename string) (*os.File, error) {
	path := filepath.Join(src, "published", strings.TrimSuffix(filename, "html")+"md")
	return os.Open(path)
}
