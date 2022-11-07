package entry

import (
	"errors"
	"html/template"
	"strings"
	"time"
)

const (
	inputDateFormat  = "2006-01-02"
	outputDateFormat = "January 2, 2006"
)

type Entry struct {
	Title     string
	Published string
	Revision  string
	Excerpt   string
	Body      template.HTML
}

func New(title string) *Entry {
	date := time.Now().Format(inputDateFormat)
	return &Entry{
		Title:     title,
		Excerpt:   "",
		Published: date,
		Revision:  date,
	}
}

func NewFromFrontMatter(fm map[string]string) (*Entry, error) {
	e := &Entry{}

	published, ok := fm["published"]
	if !ok {
		return nil, errors.New("Missing publish date in front matter")
	}
	formattedPublished, err := e.formatDate(published)
	if err != nil {
		return nil, err
	}
	e.Published = formattedPublished

	revision, ok := fm["revision"]
	if !ok {
		return nil, errors.New("Missing revision date in front matter")
	}
	formattedRevision, err := e.formatDate(revision)
	if err != nil {
		return nil, err
	}
	e.Revision = formattedRevision

	title, ok := fm["title"]
	if !ok {
		return nil, errors.New("Missing title in front matter")
	}
	e.Title = title

	excerpt, ok := fm["excerpt"]
	if !ok {
		return nil, errors.New("Missing excerpt in front matter")
	}
	e.Excerpt = excerpt

	return e, nil
}

func (e *Entry) formatDate(dateStr string) (string, error) {
	parsed, err := time.Parse(inputDateFormat, dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format(outputDateFormat), nil
}

func (e *Entry) GetTitle() string {
	capitalized := []string{}
	for _, w := range strings.Split(e.Title, "-") {
		capitalized = append(capitalized, strings.ToUpper(string(w[0]))+string(w[1:]))
	}
	return strings.Join(capitalized, " ")
}
