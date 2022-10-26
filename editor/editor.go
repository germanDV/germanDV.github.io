package editor

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"
)

type Entry struct {
	Title     string
	Published string
	Revision  string
	Body      template.HTML
}

func (e *Entry) scanFrontMatter(fm map[string]string) error {
	published, ok := fm["published"]
	if !ok {
		return errors.New("Missing published date in front matter")
	}
	formattedPublished, err := formatDate(published)
	if err != nil {
		return err
	}
	e.Published = formattedPublished

	revision, ok := fm["revision"]
	if ok {
		formattedRevision, err := formatDate(revision)
		if err != nil {
			return err
		}
		e.Revision = formattedRevision
	} else {
		e.Revision = ""
	}

	title, ok := fm["title"]
	if !ok {
		return errors.New("Missing title date in front matter")
	}
	e.Title = title

	return nil
}

func formatDate(dateStr string) (string, error) {
	inLayout := "2006-01-02"
	outLayout := "January 2, 2006"
	parsed, err := time.Parse(inLayout, dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format(outLayout), nil
}

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

func ToHTML(filename string) error {
	entry := Entry{}
	frontMatter, body, err := parseMd(filename)

	err = entry.scanFrontMatter(frontMatter)
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
