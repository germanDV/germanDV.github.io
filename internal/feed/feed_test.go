package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}

func teardown() {
	err := os.Remove("testdata/pages/feed.rss")
	if err != nil {
		fmt.Println("Error removing file generated by test!", err)
	}
}

func TestGenerateReadsPagesAndGeneratesRSSFeedFile(t *testing.T) {
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
}