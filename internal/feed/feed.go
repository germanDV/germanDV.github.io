package feed

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/entry"
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
func Generate(src, dst string) error {
	feed := Feed{
		Title:       "germandv",
		Link:        "https://germandv.xyz",
		Description: "Programming things",
		LastBuild:   time.Now().Format(time.RFC3339),
		Lang:        "en-us",
		Items:       []Item{},
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && strings.HasSuffix(name, ".md") {
			frontMatter, _, err := editor.ParseMd(filepath.Join(src, name))
			if err != nil {
				return err
			}

			art, err := entry.NewHtmlEntry(frontMatter)
			if err != nil {
				return err
			}

			feed.Items = append(feed.Items, Item{
				Title:       art.Title,
				Link:        getLink(name),
				Description: art.Excerpt,
				Created:     art.Published,
			})
		}
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "feed.rss"))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dst, "feed.rss"))
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(f, "feed", feed)
	if err != nil {
		return err
	}

	return nil
}

func getLink(mdFile string) string {
	baseURL := "https://germandv.xyz/blog/"
	htmlFile := strings.TrimSuffix(mdFile, ".md") + ".html"
	return baseURL + htmlFile
}
