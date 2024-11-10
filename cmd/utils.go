// utils.go
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/quick"
)

func printWithSyntaxHighlighting(content, responseColor string) {
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	codeLang := ""
	codeBuffer := ""

	colorCode := hexToANSI(responseColor)

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				inCodeBlock = false
				quick.Highlight(os.Stdout, codeBuffer, codeLang, "terminal16m", "monokai")
				codeBuffer = ""
			} else {
				inCodeBlock = true
				codeLang = strings.TrimSpace(strings.TrimPrefix(line, "```"))
				if codeLang == "" {
					codeLang = "plaintext"
				}
			}
			continue
		}

		if inCodeBlock {
			codeBuffer += line + "\n"
		} else {
			fmt.Printf("%s%s\x1b[0m\n", colorCode, line)
		}
	}

	if inCodeBlock && codeBuffer != "" {
		quick.Highlight(os.Stdout, codeBuffer, codeLang, "terminal16m", "monokai")
	}
}

func hexToANSI(hexColor string) string {
	hexColor = strings.TrimPrefix(hexColor, "#")
	if len(hexColor) != 6 {
		// Default to white if invalid
		return "\x1b[38;2;255;255;255m"
	}

	r, err1 := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, err2 := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, err3 := strconv.ParseInt(hexColor[4:6], 16, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		// Default to white if parsing fails
		return "\x1b[38;2;255;255;255m"
	}

	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}
