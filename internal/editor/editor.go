package editor

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"germandv.xyz/internal/entry"
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
				return nil, errors.New("Invalid front matter key-value pair")
			}
			frontMatter[keyvalue[0]] = strings.Trim(keyvalue[1], " ")
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("No content found")
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

func parseMd(filename string) (map[string]string, []byte, error) {
	fp := filepath.Join("entries", filename)

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
	Link  string
	Title string
}

// GenerateIndex (re)creates the index.html page listing all published entries.
func GenerateIndex() error {
	files, err := os.ReadDir("pages")
	if err != nil {
		return err
	}

	links := []PageLink{}

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && strings.HasSuffix(name, ".html") && name != "index.html" {
			title := strings.ReplaceAll(strings.TrimSuffix(name, ".html"), "-", " ")
			links = append(links, PageLink{Link: name, Title: title})
		}
	}

	indexWriter, err := os.Create(filepath.Join("pages", "index.html"))
	if err != nil {
		return err
	}
	defer indexWriter.Close()

	tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(indexWriter, "index", links)
	if err != nil {
		return err
	}

	return nil
}

// Publish reads the .md file from entries/, converts it to .html and saves it in pages/.
// It also adds a link to the newly published entry to the index.
func Publish(filename string) error {
	frontMatter, body, err := parseMd(filename)
	if err != nil {
		return err
	}

	entry, err := entry.NewFromFrontMatter(frontMatter)
	if err != nil {
		return err
	}

	entry.Body = template.HTML(blackfriday.Run(body))

	dst := fmt.Sprintf("%s.html", entry.Title)
	f, err := os.Create(filepath.Join("pages", dst))
	if err != nil {
		return err
	}
	defer f.Close()

	layout := filepath.Join("templates", "layout.html")
	tmpl, err := template.ParseFiles(layout)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(f, "layout", entry)
	if err != nil {
		return err
	}

	return nil
}

// PublishAll reads all .md files from entries/, converts them to .html and saves them in pages/.
func PublishAll() error {
	files, err := os.ReadDir("entries")
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() && strings.HasSuffix(name, ".md") {
			err := Publish(name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Draft creates a .md file in entries/ and pre-populates the front matter.
func Draft(title string) error {
	filename := title + ".md"

	f, err := os.Create(filepath.Join("entries", filename))
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(filepath.Join("templates", "entry.md"))
	if err != nil {
		return err
	}

	err = tpl.Execute(f, entry.New(title))
	if err != nil {
		return err
	}

	return nil
}
