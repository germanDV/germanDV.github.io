package main

import (
	"germandv.xyz/editor"
	"germandv.xyz/server"
)

func main() {
	// Convert MD to HTML
	editor.ToHTML("test.md")

	// Start server
	s := server.New(4000)
	s.Listen()
}
