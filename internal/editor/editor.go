package editor

import (
	"bufio"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"germandv.xyz/internal/entry"
	"germandv.xyz/internal/filer"
	"github.com/russross/blackfriday/v2"
)

func readFrontMatter(scanner *bufio.Scanner) (map[string]string, error) {
	frontMatter := make(map[string]string)
	openingDelimiterSeen := false

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")

		if line == "---" {
			if openingDelimiterSeen {
				// End of the front matter
				return frontMatter, nil
			} else {
				// Beginning of the front matter
				openingDelimiterSeen = true
			}
		} else if openingDelimiterSeen {
			keyvalue := strings.SplitN(line, ":", 2)
			if len(keyvalue) != 2 {
				return nil, errors.New("invalid front matter key-value pair")
			}
			frontMatter[keyvalue[0]] = strings.Trim(keyvalue[1], " ")
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("no content found")
}

func readBody(scanner *bufio.Scanner) ([]byte, error) {
	body := []byte{}
	for scanner.Scan() {
		body = append(body, scanner.Bytes()...)
		body = append(body, '\n')
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ParseMd(fp string) (map[string]string, []byte, error) {
	f, err := os.Open(fp)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	frontMatter, err := readFrontMatter(scanner)
	if err != nil {
		return nil, nil, err
	}

	body, err := readBody(scanner)
	if err != nil {
		return nil, nil, err
	}

	return frontMatter, body, nil
}

type PageLink struct {
	Link        string
	Title       string
	Date        time.Time
	DateDisplay string
	Tags        []string
}

// GenerateIndex (re)creates the index.html page listing all published entries.
func GenerateIndex() error {
	files, err := filer.ListPages()
	if err != nil {
		return err
	}

	links := []PageLink{}

	for _, file := range files {
		title := strings.ReplaceAll(strings.TrimSuffix(file, ".html"), "-", " ")

		// Read .md file to get front matter.
		entryMd, err := filer.GetPublishedEntry(file)
		if err != nil {
			return err
		}
		defer entryMd.Close()
		scanner := bufio.NewScanner(entryMd)
		frontMatter, err := readFrontMatter(scanner)
		if err != nil {
			return err
		}

		date, err := time.Parse(entry.InputDateFormat, frontMatter["revision"])
		if err != nil {
			return err
		}
		dateDisplay, err := entry.FormatDate(frontMatter["revision"])
		if err != nil {
			return err
		}

		tags := strings.Split(frontMatter["tags"], ",")
		if tags[0] == "" {
			tags = []string{}
		}

		links = append(links, PageLink{
			Link:        file,
			Title:       title,
			Date:        date,
			DateDisplay: dateDisplay,
			Tags:        tags,
		})
	}

	indexWriter, err := filer.CreateIndex()
	if err != nil {
		return err
	}
	defer indexWriter.Close()

	index := filepath.Join("templates", "index.html")
	footer := filepath.Join("templates", "footer.html")
	tmpl, err := template.ParseFiles(index, footer)
	if err != nil {
		return err
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].Date.After(links[j].Date)
	})

	err = tmpl.ExecuteTemplate(indexWriter, "index", links)
	if err != nil {
		return err
	}

	return nil
}

// Publish reads the .md file from `src`, converts it to .html and saves it in `dst`.
// It also adds a link to the newly published entry to the index.
func Publish(entryfile string) error {
	frontMatter, body, err := ParseMd(entryfile)
	if err != nil {
		return err
	}

	entry, err := entry.NewHtmlEntry(frontMatter)
	if err != nil {
		return err
	}

	entry.Body = template.HTML(blackfriday.Run(body))

	f, err := filer.CreatePage(entry.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	layout := filepath.Join("templates", "layout.html")
	footer := filepath.Join("templates", "footer.html")
	tmpl, err := template.ParseFiles(layout, footer)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(f, "layout", entry)
	if err != nil {
		return err
	}

	err = filer.Publish(entryfile)
	if err != nil {
		return err
	}

	return nil
}

// PublishAll reads all .md files from `src`, converts them to .html and saves them in `dst`.
func PublishAll() error {
	drafts, err := filer.ListDrafts()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, draft := range drafts {
		wg.Add(1)
		go func(waitgroup *sync.WaitGroup, draftname string) {
			defer waitgroup.Done()
			Publish(draftname)
		}(&wg, draft)
	}

	wg.Wait()
	return nil
}

// Draft creates a .md file in `src` and pre-populates the front matter.
func Draft(title string) error {
	f, err := filer.CreateDraft(title + ".md")
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(filepath.Join("templates", "entry.md"))
	if err != nil {
		return err
	}

	err = tpl.Execute(f, entry.NewMdEntry(title))
	if err != nil {
		return err
	}

	return nil
}

// Preview reads a draft .md file and returns its HTML version,
// without persisting anything to disk.
func Preview(filename string) (*template.Template, *entry.HtmlEntry, error) {
	frontMatter, body, err := ParseMd(filename)
	if err != nil {
		return nil, nil, err
	}

	entry, err := entry.NewHtmlEntry(frontMatter)
	if err != nil {
		return nil, nil, err
	}

	entry.Body = template.HTML(blackfriday.Run(body))
	layout := filepath.Join("templates", "layout.html")
	footer := filepath.Join("templates", "footer.html")
	tmpl, err := template.ParseFiles(layout, footer)
	if err != nil {
		return nil, nil, err
	}

	return tmpl, entry, nil
}
