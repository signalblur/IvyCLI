package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

func printWithMarkdown(content string) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Println("Error creating renderer:", err)
		return
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		fmt.Println("Error rendering content:", err)
		return
	}

	fmt.Print(rendered)
}
