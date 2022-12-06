package entry

import (
	"errors"
	"html/template"
	"strings"
	"time"
)

const (
	InputDateFormat  = "2006-01-02"
	OutputDateFormat = "January 2, 2006"
)

type HtmlEntry struct {
	Filename  string
	Title     string
	Published string
	Revision  string
	Excerpt   string
	Body      template.HTML
}

type MdEntry struct {
	Title     string
	Published string
	Revision  string
	Excerpt   string
}

// NewMdEntry creates a new Markdown entry using the current date.
func NewMdEntry(title string) *MdEntry {
	date := time.Now().Format(InputDateFormat)
	return &MdEntry{
		Title:     title,
		Excerpt:   "",
		Published: date,
		Revision:  date,
	}
}

// NewHtmlEntry creates a new HTML entry taking a front matter as input.
// Title is capitalized and "-" replaced with spaces.
func NewHtmlEntry(fm map[string]string) (*HtmlEntry, error) {
	e := &HtmlEntry{}

	published, ok := fm["published"]
	if !ok {
		return nil, errors.New("Missing publish date in front matter")
	}
	formattedPublished, err := FormatDate(published)
	if err != nil {
		return nil, err
	}
	e.Published = formattedPublished

	revision, ok := fm["revision"]
	if !ok {
		return nil, errors.New("Missing revision date in front matter")
	}
	formattedRevision, err := FormatDate(revision)
	if err != nil {
		return nil, err
	}
	e.Revision = formattedRevision

	title, ok := fm["title"]
	if !ok {
		return nil, errors.New("Missing title in front matter")
	}
	e.Filename = title
	e.Title = parseTitle(title)

	excerpt, ok := fm["excerpt"]
	if !ok {
		return nil, errors.New("Missing excerpt in front matter")
	}
	e.Excerpt = excerpt

	return e, nil
}

func FormatDate(dateStr string) (string, error) {
	parsed, err := time.Parse(InputDateFormat, dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format(OutputDateFormat), nil
}

func parseTitle(title string) string {
	capitalized := []string{}
	for _, w := range strings.Split(title, "-") {
		capitalized = append(capitalized, strings.ToUpper(string(w[0]))+string(w[1:]))
	}
	return strings.Join(capitalized, " ")
}
