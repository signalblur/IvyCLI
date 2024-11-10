package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
)

func printWithMarkdown(content, responseColor string) {
	colorCode := hexToANSI(responseColor)

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

	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		fmt.Printf("%s%s\x1b[0m\n", colorCode, line)
	}
}

func hexToANSI(hexColor string) string {
	hexColor = strings.TrimPrefix(hexColor, "#")
	if len(hexColor) != 6 {
		return "\x1b[38;2;255;255;255m"
	}

	r, err1 := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, err2 := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, err3 := strconv.ParseInt(hexColor[4:6], 16, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return "\x1b[38;2;255;255;255m"
	}

	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
