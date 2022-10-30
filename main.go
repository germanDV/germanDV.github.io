package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"germandv.xyz/editor"
	"germandv.xyz/server"
)

func main() {
	startServer := flag.Bool("serve", false, "Start web server")
	entryToPublish := flag.String("publish", "", "Entry to be published")
	entryToCreate := flag.String("draft", "", "Entry to be created as a draft")
	flag.Parse()
	if *startServer {
		serve()
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
	fmt.Printf("%q published!\n", title)
}
