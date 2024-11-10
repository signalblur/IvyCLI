package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

func main() {
	// Define command-line flags
	configFlag := flag.String("config", "", "Path to configuration file")
	noHistoryFlag := flag.Bool("no-history", false, "Disable conversation history")
	resetHistoryFlag := flag.Bool("reset-history", false, "Reset conversation history")
	noColorFlag := flag.Bool("no-color", false, "Disable syntax highlighting")
	timeoutFlag := flag.Int("timeout", 30, "HTTP request timeout in seconds")
	flag.Parse()

	// Load API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY must be set via environment variables.")
		os.Exit(1)
	}

	// Load passphrase from environment variable
	passphrase := os.Getenv("IVYCLI_PASSPHRASE")
	if passphrase == "" {
		fmt.Fprintln(os.Stderr, "Error: IVYCLI_PASSPHRASE must be set via environment variables.")
		os.Exit(1)
	}

	// Check if config file is specified via command-line flag or environment variable
	if *configFlag == "" {
		*configFlag = os.Getenv("IVYCLI_CONFIG_PATH")
	}
	if *configFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: Config file must be specified via --config flag or IVYCLI_CONFIG_PATH environment variable.")
		os.Exit(1)
	}

	// Load configurations from file
	var model string
	var systemPrompt string
	var responseColor string
	var maxHistorySize int

	configData, err := os.ReadFile(*configFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		os.Exit(1)
	}
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config file:", err)
		os.Exit(1)
	}
	if val, ok := config["model"].(string); ok {
		model = val
	} else {
		fmt.Fprintln(os.Stderr, "Error: Model must be specified in the config file.")
		os.Exit(1)
	}
	if val, ok := config["system_prompt"].(string); ok {
		systemPrompt = val
	}
	if val, ok := config["response_color"].(string); ok {
		responseColor = val
	} else {
		responseColor = "#FFFFFF" // Default color (white)
	}
	if val, ok := config["max_history_size"].(float64); ok {
		maxHistorySize = int(val)
	} else {
		maxHistorySize = 10 // Default max history size
	}

	// Handle reset history flag
	if *resetHistoryFlag {
		err := resetConversationHistory()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error resetting conversation history:", err)
			os.Exit(1)
		}
		fmt.Println("Conversation history has been reset.")
		os.Exit(0)
	}

	// Parse command-line arguments or read from stdin
	var prompt string
	if flag.NArg() == 0 {
		fmt.Println("Enter your message (end with Ctrl+D):")
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
			os.Exit(1)
		}
		prompt = strings.Join(lines, "\n")
	} else {
		prompt = strings.Join(flag.Args(), " ")
	}

	// Initialize messages slice
	var messages []map[string]string

	// Include the system prompt if it's set
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	// Conversation history handling
	if !*noHistoryFlag {
		// Load conversation history
		historyMessages, err := loadConversationHistory(passphrase)
		if err != nil {
			// If history file doesn't exist, continue without error
			if !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "Error loading conversation history:", err)
			}
		} else {
			messages = append(messages, historyMessages...)
		}
	}

	// Append the user's message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	// Create the request body
	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating request:", err)
		os.Exit(1)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating HTTP request:", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request securely
	client := &http.Client{
		Timeout: time.Duration(*timeoutFlag) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		if errMsg, ok := errorResponse["error"].(map[string]interface{}); ok {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg["message"])
		} else {
			fmt.Fprintf(os.Stderr, "Error: Received non-200 response status: %s\n", resp.Status)
		}
		os.Exit(1)
	}

	// Handle the response
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding response:", err)
		os.Exit(1)
	}

	// Extract the assistant's reply
	if choices, ok := responseBody["choices"].([]interface{}); ok && len(choices) > 0 {
		if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
			content := strings.TrimSpace(message["content"].(string))

			// Insert a newline before printing the response
			fmt.Println()

			// Output formatting and syntax highlighting
			if *noColorFlag {
				fmt.Println(content)
			} else {
				printWithMarkdown(content, responseColor)
			}

			// Save conversation history if not disabled
			if !*noHistoryFlag {
				// Append the assistant's reply to messages
				messages = append(messages, map[string]string{
					"role":    "assistant",
					"content": content,
				})
				// Remove the system prompt before saving history
				var historyMessages []map[string]string
				for _, msg := range messages {
					if msg["role"] != "system" {
						historyMessages = append(historyMessages, msg)
					}
				}
				// Limit history size
				if len(historyMessages) > maxHistorySize*2 {
					historyMessages = historyMessages[len(historyMessages)-maxHistorySize*2:]
				}
				err = saveConversationHistory(historyMessages, passphrase)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error saving conversation history:", err)
				}
			}

			os.Exit(0)
		}
	}

	fmt.Fprintln(os.Stderr, "Error: Unexpected response format.")
	os.Exit(1)
}
