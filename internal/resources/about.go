package resources

import _ "embed"

// Version will be injected at build time using -ldflags.
//
// Example:
//
//	go build -ldflags "-X github.com/game_tool_box/internal/resources.Version=v1.2.3" ./cmd/game_tool_box
var Version = "dev"

//go:embed about.md
var aboutMarkdown string

func AboutMarkdown() string {
	return aboutMarkdown
}
