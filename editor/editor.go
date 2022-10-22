package editor

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func ToHTML(filename string) error {
	frontMatter, body, err := parseMd(filename)

	published, ok := frontMatter["published"]
	if !ok {
		return errors.New("Missing published date in front matter")
	}

	title, ok := frontMatter["title"]
	if !ok {
		return errors.New("Missing title date in front matter")
	}

	dst := title + ".html"
	f, err := os.Create(filepath.Join("templates", dst))
	if err != nil {
		return err
	}
	defer f.Close()

	output := blackfriday.Run(body)
	f.WriteString("{{define \"body\"}}\n<main>\n")
	f.WriteString(fmt.Sprintf("<time datetime=%q>%s</time>\n", published, published))
	f.Write(output)
	f.WriteString("</main>\n{{end}}")
	return nil
}
