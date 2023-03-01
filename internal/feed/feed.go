package feed

import (
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/entry"
	"germandv.xyz/internal/filer"
)

type Item struct {
	Title       string
	Link        string
	Description string
	Created     string
}

type Feed struct {
	Title       string
	Link        string
	Description string
	LastBuild   string
	Lang        string
	Items       []Item
}

// Generate creates a `feed.rss` file with all entries.
func Generate() error {
	feed := Feed{
		Title:       "germandv",
		Link:        "https://germandv.me",
		Description: "Programming things",
		LastBuild:   time.Now().Format(time.RFC3339),
		Lang:        "en-us",
		Items:       []Item{},
	}

	files, err := filer.ListPublished()
	if err != nil {
		return err
	}

	for _, file := range files {
		frontMatter, _, err := editor.ParseMd(file)
		if err != nil {
			return err
		}

		art, err := entry.NewHtmlEntry(frontMatter)
		if err != nil {
			return err
		}

		feed.Items = append(feed.Items, Item{
			Title:       art.Title,
			Link:        getLink(file),
			Description: art.Excerpt,
			Created:     art.Published,
		})
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "feed.xml"))
	if err != nil {
		return err
	}

	f, err := filer.CreateFeed()
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(f, "feed", feed)
	if err != nil {
		return err
	}

	return nil
}

func getLink(mdFilepath string) string {
	baseURL := "https://germandv.me/blog/"
	parts := strings.Split(mdFilepath, "/")
	mdFile := parts[len(parts)-1]
	htmlFile := strings.TrimSuffix(mdFile, ".md") + ".html"
	return baseURL + htmlFile
}
