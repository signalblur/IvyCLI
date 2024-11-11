package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func runREPL(apiKey, passphrase, model, systemPrompt string, maxHistorySize int, enableMarkdown, noHistory bool, timeout int) {
	fmt.Println("Entering REPL mode. Press Ctrl+C to exit.")
	reader := bufio.NewReader(os.Stdin)
	var messages []map[string]string
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	if !noHistory {
		historyMessages, err := loadConversationHistory(passphrase)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error loading conversation history:", err)
		} else {
			messages = append(messages, historyMessages...)
		}
	}
	for {
		fmt.Print("> ")
		prompt, err := reader.ReadString('\n')
		if err != nil {
			if err == syscall.EINTR {
				fmt.Println("\nExiting REPL mode.")
				break
			}
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			break
		}
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			continue
		}
		errPrompt := handlePrompt(prompt, &messages, apiKey, passphrase, model, maxHistorySize, enableMarkdown, noHistory, timeout)
		if errPrompt != nil {
			fmt.Fprintln(os.Stderr, "Error handling prompt:", errPrompt)
		}
	}
}
