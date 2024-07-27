package entry

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewMDEntry(t *testing.T) {
	t.Parallel()

	date := time.Now().Format(InputDateFormat)
	title := "my-test-entry"
	e := NewMdEntry(title)

	if e.Title != title {
		t.Errorf("want title %q, got %q", title, e.Title)
	}
	if e.Published != date {
		t.Errorf("want published date to be %s, got %s", date, e.Published)
	}
	if e.Revision != date {
		t.Errorf("want revision date to be %s, got %s", date, e.Revision)
	}
}

func TestNewHTMLEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  map[string]string
		output *HtmlEntry
		err    error
	}{
		{
			input: map[string]string{
				"revision": "",
				"title":    "",
				"excerpt":  "",
			},
			output: nil,
			err:    errors.New("missing publish date in front matter"),
		},
		{
			input: map[string]string{
				"published": "bad-date",
				"revision":  "",
				"title":     "",
				"excerpt":   "",
			},
			output: nil,
			err:    errors.New("parsing time \"bad-date\" as \"2006-01-02\": cannot parse \"bad-date\" as \"2006\""),
		},
		{
			input: map[string]string{
				"published": "1987-08-06",
				"title":     "",
				"excerpt":   "",
			},
			output: nil,
			err:    errors.New("missing revision date in front matter"),
		},
		{
			input: map[string]string{
				"published": "1987-08-06",
				"revision":  "1987-08-06",
				"excerpt":   "",
			},
			output: nil,
			err:    errors.New("missing title in front matter"),
		},
		{
			input: map[string]string{
				"published": "1987-08-06",
				"revision":  "1987-08-06",
				"title":     "a-title",
			},
			output: nil,
			err:    errors.New("missing excerpt in front matter"),
		},
		{
			input: map[string]string{
				"published": "1987-bad-06",
				"revision":  "1987-08-06",
				"title":     "a-title",
				"excerpt":   "blah blah blah",
			},
			output: nil,
			err:    errors.New("parsing time \"1987-bad-06\" as \"2006-01-02\": cannot parse \"bad-06\" as \"01\""),
		},
		{
			input: map[string]string{
				"published": "1987-08-06",
				"revision":  "1987-08-06",
				"title":     "a-title-foo-bar",
				"excerpt":   "blah blah blah",
			},
			output: &HtmlEntry{
				Filename:  "a-title-foo-bar",
				Published: "August 6, 1987",
				Revision:  "August 6, 1987",
				Title:     "A Title Foo Bar",
				Excerpt:   "blah blah blah",
			},
			err: nil,
		},
	}

	for i, tt := range tests {
		testname := fmt.Sprintf("test#%d", i)
		t.Run(testname, func(t *testing.T) {
			htmlEntry, err := NewHtmlEntry(tt.input)
			cmpHtmlEntries(t, htmlEntry, tt.output)
			cmpErrors(t, err, tt.err)
		})
	}
}

func cmpHtmlEntries(t *testing.T, got, want *HtmlEntry) {
	t.Helper()

	if want == nil && got != nil {
		t.Errorf("want no output, got %+v", got)
	}
	if want != nil && got == nil {
		t.Errorf("got no output, wanted %+v", want)
	}

	if want != nil && got != nil {
		if want.Filename != got.Filename {
			t.Errorf("want filename %q, got %q", want.Filename, got.Filename)
		}
		if want.Title != got.Title {
			t.Errorf("want title %q, got %q", want.Title, got.Title)
		}
		if want.Published != got.Published {
			t.Errorf("want published date %q, got %q", want.Published, got.Published)
		}
		if want.Revision != got.Revision {
			t.Errorf("want revision date %q, got %q", want.Revision, got.Revision)
		}
		if want.Excerpt != got.Excerpt {
			t.Errorf("want excerpt date %q, got %q", want.Excerpt, got.Excerpt)
		}
	}
}

func cmpErrors(t *testing.T, got, want error) {
	t.Helper()
	if want != nil && got == nil {
		t.Errorf("want error %s, but got nil", want)
	}
	if want == nil && got != nil {
		t.Errorf("want no error, got %s", got)
	}
	if want != nil && got != nil {
		if want.Error() != got.Error() {
			t.Errorf("want error %q, got %q", want.Error(), got.Error())
		}
	}
}
