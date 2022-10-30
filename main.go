package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"germandv.xyz/editor"
	"germandv.xyz/server"
)

func main() {
	startServer := flag.Bool("server", false, "Start web server")
	entryToPublish := flag.String("publish", "", "Entry to be published")
	entryToCreate := flag.String("draft", "", "Entry to be created as a draft")
	flag.Parse()

	if *startServer {
		s := server.New(4000)
		s.Listen()
	} else if *entryToPublish != "" {
		if !strings.HasSuffix(*entryToPublish, ".md") {
			*entryToPublish = *entryToPublish + ".md"
		}
		err := editor.ToHTML(*entryToPublish)
		if err != nil {
			fmt.Printf("Error publishing entry %q\n", *entryToPublish)
			fmt.Println(err)
			os.Exit(1)
		}
		// TODO: add link to index
		fmt.Printf("%q published!\n", *entryToPublish)
	} else if *entryToCreate != "" {
		file := *entryToCreate + ".md"
		f, err := os.Create(filepath.Join("entries", file))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		tpl, err := template.ParseFiles(filepath.Join("templates", "entry.md"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		date := time.Now().Format("January 2, 2006")
		err = tpl.Execute(f, editor.Entry{
			Title:     *entryToCreate,
			Published: date,
			Revision:  date,
		})

		fmt.Printf("%q created!\n", file)
	} else {
		fmt.Println("Unknown operation.")
	}
}
