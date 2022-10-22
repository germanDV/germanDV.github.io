package main

import (
	"germandv.xyz/editor"
	"germandv.xyz/server"
)

func main() {
	// Convert MD to HTML and save it.
	editor.ToHTML("test.md")

	// Start server.
	s := server.New(4000)
	s.Listen()
}
