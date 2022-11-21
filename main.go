package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/feed"
	"germandv.xyz/internal/server"
)

const (
	src = "entries"
	dst = "pages"
)

func main() {
	startServer := flag.Bool("serve", false, "Start web server")
	entryToPublish := flag.String("publish", "", "Entry to be published (use 'all' to publish everything)")
	entryToCreate := flag.String("draft", "", "Entry to be created as a draft")
	rss := flag.Bool("feed", false, "Generate RSS feed")
	flag.Parse()
	if *startServer {
		serve()
	} else if *entryToPublish == "all" {
		publishAll()
	} else if *entryToPublish != "" {
		publish(*entryToPublish)
	} else if *entryToCreate != "" {
		create(*entryToCreate)
	} else if *rss {
		generateFeed()
	} else {
		fmt.Println("Unknown operation. Run -h for help.")
	}
}

func serve() {
	s := server.New(4000)
	s.Listen()
}

func create(title string) {
	must(editor.Draft(title, src), fmt.Sprintf("Error creating draft entry %q\n", title))
	fmt.Printf("%q created!\n", title+".md")
}

func publish(title string) {
	if !strings.HasSuffix(title, ".md") {
		title = title + ".md"
	}
	must(editor.Publish(title, src, dst), fmt.Sprintf("Error publishing entry %q\n", title))
	must(editor.GenerateIndex(dst), "Error generating index.html")
	fmt.Printf("%q published!\n", title)
}

func publishAll() {
	must(editor.PublishAll(src, dst), "Error publishing all entries")
	must(editor.GenerateIndex(dst), "Error generating index.html")
	fmt.Println("All entries published!")
}

func generateFeed() {
	must(feed.Generate(src, dst), "Error generating rss feed")
	fmt.Println("RSS feed generated!")
}

func must(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		fmt.Println(err)
		os.Exit(1)
	}
}
