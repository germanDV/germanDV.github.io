package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"germandv.xyz/internal/editor"
	"germandv.xyz/internal/server"
)

// TODO:
//    - Embed all assets so binary is self-contained
//    - Gzip http responses
//    - Add RSS feed
//    - Add Favicon
//    - Add SQLite to count views

func main() {
	startServer := flag.Bool("serve", false, "Start web server")
	entryToPublish := flag.String("publish", "", "Entry to be published")
	all := flag.Bool("all", false, "Publish all entries")
	entryToCreate := flag.String("draft", "", "Entry to be created as a draft")
	flag.Parse()
	if *startServer {
		serve()
	} else if *all {
		publishAll()
	} else if *entryToPublish != "" {
		publish(*entryToPublish)
	} else if *entryToCreate != "" {
		create(*entryToCreate)
	} else {
		fmt.Println("Unknown operation. Run -h for help.")
	}
}

func serve() {
	s := server.New(4000)
	s.Listen()
}

func create(title string) {
	err := editor.Draft(title)
	if err != nil {
		fmt.Printf("Error creating draft entry %q\n", title)
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%q created!\n", title+".md")
}

func publish(title string) {
	if !strings.HasSuffix(title, ".md") {
		title = title + ".md"
	}

	err := editor.Publish(title)
	if err != nil {
		fmt.Printf("Error publishing entry %q\n", title)
		fmt.Println(err)
		os.Exit(1)
	}

	err = editor.GenerateIndex()
	if err != nil {
		fmt.Println("Error generating index.html")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%q published!\n", title)
}

func publishAll() {
	err := editor.PublishAll()
	if err != nil {
		fmt.Printf("Error publishing all entries")
		fmt.Println(err)
		os.Exit(1)
	}

	err = editor.GenerateIndex()
	if err != nil {
		fmt.Printf("Error generating index.html")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("All entries published!")
}
