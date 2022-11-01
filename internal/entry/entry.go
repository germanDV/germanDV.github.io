package entry

import (
	"errors"
	"html/template"
	"time"
)

const (
	inputDateFormat   = "2006-01-02"
	outpoutDateFormat = "January 2, 2006"
)

type Entry struct {
	Title     string
	Published string
	Revision  string
	Body      template.HTML
}

func New(title string) *Entry {
	date := time.Now().Format(inputDateFormat)
	return &Entry{
		Title:     title,
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

	return e, nil
}

func (e *Entry) formatDate(dateStr string) (string, error) {
	parsed, err := time.Parse(inputDateFormat, dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format(outpoutDateFormat), nil
}
