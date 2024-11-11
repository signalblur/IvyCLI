package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

func main() {
	configFlag := flag.String("config", "", "Path to configuration file")
	noMarkdownFlag := flag.Bool("disable-markdown", false, "Disable markdown formatting")
	resetHistoryFlag := flag.Bool("reset-history", false, "Reset conversation history")
	noHistoryFlag := flag.Bool("no-history", false, "Disable conversation history")
	timeoutFlag := flag.Int("timeout", 30, "HTTP request timeout in seconds")
	flag.Parse()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY must be set via environment variables.")
		os.Exit(1)
	}

	passphrase := os.Getenv("IVYCLI_PASSPHRASE")
	if passphrase == "" {
		fmt.Fprintln(os.Stderr, "Error: IVYCLI_PASSPHRASE must be set via environment variables.")
		os.Exit(1)
	}

	if *configFlag == "" {
		*configFlag = os.Getenv("IVYCLI_CONFIG_PATH")
	}
	if *configFlag == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting current user:", err)
			os.Exit(1)
		}
		*configFlag = filepath.Join(usr.HomeDir, ".config", "ivycli", "config.json")
	}

	if _, err := os.Stat(*configFlag); os.IsNotExist(err) {
		err = firstTimeSetup(*configFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error during first-time setup:", err)
			os.Exit(1)
		}
	}

	var model string
	var systemPrompt string
	var maxHistorySize int
	var enableMarkdown bool

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
	if val, ok := config["max_history_size"].(float64); ok {
		maxHistorySize = int(val)
	} else {
		maxHistorySize = 10
	}
	if val, ok := config["enable_markdown"].(bool); ok {
		enableMarkdown = val
	} else {
		enableMarkdown = true
	}

	if *noMarkdownFlag {
		enableMarkdown = false
	}

	if *resetHistoryFlag {
		err := resetConversationHistory()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error resetting conversation history:", err)
			os.Exit(1)
		}
		fmt.Println("Conversation history has been reset.")
		os.Exit(0)
	}

	var prompt string
	if flag.NArg() == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
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
			fmt.Fprintln(os.Stderr, "Error: No prompt provided.")
			os.Exit(1)
		}
	} else {
		prompt = strings.Join(flag.Args(), " ")
	}

	var messages []map[string]string

	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	if !*noHistoryFlag {
		historyMessages, err := loadConversationHistory(passphrase)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "Error loading conversation history:", err)
			}
		} else {
			messages = append(messages, historyMessages...)
		}
	}

	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating request:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating HTTP request:", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding response:", err)
		os.Exit(1)
	}

	if choices, ok := responseBody["choices"].([]interface{}); ok && len(choices) > 0 {
		if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
			content := strings.TrimSpace(message["content"].(string))

			if enableMarkdown {
				printWithMarkdown(content)
			} else {
				fmt.Println(content)
			}

			if !*noHistoryFlag {
				messages = append(messages, map[string]string{
					"role":    "assistant",
					"content": content,
				})
				var historyMessages []map[string]string
				for _, msg := range messages {
					if msg["role"] != "system" {
						historyMessages = append(historyMessages, msg)
					}
				}
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

func firstTimeSetup(configPath string) error {
	fmt.Println("IvyCLI First-Time Setup")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the OpenAI model to use (e.g., gpt-4): ")
	model, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gpt-4"
	}

	fmt.Print("Enter a system prompt (optional): ")
	systemPrompt, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	systemPrompt = strings.TrimSpace(systemPrompt)

	fmt.Print("Enter maximum conversation history size (default 10): ")
	maxHistorySizeStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	maxHistorySizeStr = strings.TrimSpace(maxHistorySizeStr)
	maxHistorySize := 10
	if maxHistorySizeStr != "" {
		maxHistorySize, err = strconv.Atoi(maxHistorySizeStr)
		if err != nil {
			fmt.Println("Invalid number, using default 10.")
			maxHistorySize = 10
		}
	}

	fmt.Print("Enable markdown formatting? (yes/no, default yes): ")
	enableMarkdownStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	enableMarkdownStr = strings.TrimSpace(strings.ToLower(enableMarkdownStr))
	enableMarkdown := true
	if enableMarkdownStr == "no" || enableMarkdownStr == "n" {
		enableMarkdown = false
	}

	config := map[string]interface{}{
		"model":            model,
		"system_prompt":    systemPrompt,
		"max_history_size": maxHistorySize,
		"enable_markdown":  enableMarkdown,
	}

	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return err
	}

	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		return err
	}

	fmt.Println("Configuration saved to", configPath)
	return nil
}
