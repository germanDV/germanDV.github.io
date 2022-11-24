package feed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateReadsPagesAndGeneratesRSSFeedFile(t *testing.T) {
	t.Parallel()

	// Change directory to the root to facilitate access to `templates/`.
	err := os.Chdir("../../")
	if err != nil {
		t.Errorf("Error changing directory: %s", err)
	}

	os.Setenv("SRC", "testdata/entries")
	os.Setenv("DST", "testdata/pages")
	err = Generate()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	readAndValidateFeed(t, "testdata/pages")
}

func readAndValidateFeed(t *testing.T, dir string) {
	t.Helper()

	f, err := os.Open(filepath.Join(dir, "feed.rss"))
	if err != nil {
		t.Errorf("Error opening RSS file: %s", err)
	}
	defer f.Close()

	// TODO: validate the actual contents of the file.
}
