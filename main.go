package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/feed"
	"germandv.xyz/internal/filer"
	"germandv.xyz/internal/server"
)

func main() {
	startServer := flag.Bool("serve", false, "Start web server")
	publishEverything := flag.Bool("publish-all", false, "Publish all entries")
	publishDraft := flag.Bool("publish", false, "Choose draft entry to publish")
	entryToCreate := flag.String("draft", "", "Entry to be created as a draft")
	rss := flag.Bool("feed", false, "Generate RSS feed")
	flag.Parse()
	if *startServer {
		serve()
	} else if *publishEverything {
		publishAll()
		generateFeed()
	} else if *publishDraft {
		publish()
		generateFeed()
	} else if *entryToCreate != "" {
		create(*entryToCreate)
	} else if *rss {
		generateFeed()
	} else {
		// By default, start the web server.
		serve()
	}
}

func serve() {
	portStr, ok := os.LookupEnv("PORT")
	if !ok {
		portStr = "4000"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("PORT is not a number")
	}
	s := server.New(port)
	s.Listen()
}

func create(title string) {
	must(editor.Draft(title), fmt.Sprintf("Error creating draft entry %q\n", title))
	fmt.Printf("%q created!\n", title+".md")
}

func publish() {
	// List draft entries
	drafts, err := filer.ListDrafts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(drafts) == 0 {
		fmt.Println("You have no draft entries to publish")
		os.Exit(0)
	}

	fmt.Println("Select the number of the entry you wish to publish")
	for id, name := range drafts {
		parts := strings.Split(name, "/")
		filename := parts[len(parts)-1]
		fmt.Printf("[%d] %s\n", id, filename)
	}

	var answer uint
	fmt.Scanf("%d", &answer)

	entryToPublish, ok := drafts[answer]
	if !ok {
		fmt.Printf("No entry with ID %d\n", answer)
		os.Exit(1)
	}

	// Publish
	must(editor.Publish(entryToPublish), fmt.Sprintf("Error publishing entry %q\n", entryToPublish))
	must(editor.GenerateIndex(), "Error generating index.html")
	fmt.Printf("%q published!\n", entryToPublish)
}

func publishAll() {
	must(editor.PublishAll(), "Error publishing all entries")
	must(editor.GenerateIndex(), "Error generating index.html")
	fmt.Println("All entries published!")
}

func generateFeed() {
	must(feed.Generate(), "Error generating rss feed")
	fmt.Println("RSS feed generated!")
}

func must(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		fmt.Println(err)
		os.Exit(1)
	}
}
