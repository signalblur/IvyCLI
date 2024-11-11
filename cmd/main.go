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
	// Define flags
	noMarkdownFlag := flag.Bool("disable-markdown", false, "Disable markdown formatting")
	resetHistoryFlag := flag.Bool("reset-history", false, "Reset conversation history")
	noHistoryFlag := flag.Bool("no-history", false, "Disable conversation history")
	replFlag := flag.Bool("repl", false, "Enter REPL mode for interactive conversation")
	timeoutFlag := flag.Int("timeout", 30, "HTTP request timeout in seconds")
	helpFlag := flag.Bool("help", false, "Display help information")

	// Get user information and paths
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting current user:", err)
		os.Exit(1)
	}
	configPath := filepath.Join(usr.HomeDir, ".config", "ivycli", "config.json")

	// Set custom usage function for help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "  [options] [prompt]")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nConfiguration file location:")
		fmt.Fprintf(os.Stderr, "  %s\n", configPath)
		fmt.Fprintln(os.Stderr, "Conversation history file location:")
		historyFilePath := getHistoryFilePath()
		fmt.Fprintf(os.Stderr, "  %s\n", historyFilePath)
	}

	// Parse flags
	flag.Parse()

	// Display help and exit if help flag is set
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	configDir := filepath.Dir(configPath)

	// Check if ~/.config/ivycli/ directory exists
	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		// Run first-time setup
		err = firstTimeSetup(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error during first-time setup:", err)
			os.Exit(1)
		}
		// Exit after setup to prevent "No prompt provided" error
		os.Exit(0)
	}

	// Read API key and passphrase from environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	passphrase := os.Getenv("IVYCLI_PASSPHRASE")
	if apiKey == "" || passphrase == "" {
		// Prompt the user to enter them, write to shell profile, set them in current process
		err = setupEnvironmentVariables()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error setting up environment variables:", err)
			os.Exit(1)
		}
		// Re-read the environment variables
		apiKey = os.Getenv("OPENAI_API_KEY")
		passphrase = os.Getenv("IVYCLI_PASSPHRASE")
	}

	// Load configuration from config file
	var model string
	var systemPrompt string
	var maxHistorySize int
	var enableMarkdown bool

	configData, err := os.ReadFile(configPath)
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
	} else {
		// Set default system prompt
		systemPrompt = "You are a technical assistant. Provide concise, accurate answers to technical questions, assuming the user has a strong background in technology. Focus on brevity and clarity."
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

	if *replFlag {
		runREPL(apiKey, passphrase, model, systemPrompt, maxHistorySize, enableMarkdown, *noHistoryFlag, *timeoutFlag)
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

	// Initialize messages
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

	err = handlePrompt(prompt, &messages, apiKey, passphrase, model, maxHistorySize, enableMarkdown, *noHistoryFlag, *timeoutFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error handling prompt:", err)
		os.Exit(1)
	}
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
	if systemPrompt == "" {
		// Set default system prompt
		systemPrompt = "You are a technical assistant. Provide concise, accurate answers to technical questions, assuming the user has a strong background in technology. Focus on brevity and clarity."
	}

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

	// Prompt for API key and passphrase (always prompt and overwrite)
	err = setupEnvironmentVariables()
	if err != nil {
		return err
	}

	// Prompt to create shell alias
	fmt.Print("Would you like to create a shell alias for easier use? (yes/no, default yes): ")
	aliasResponse, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	aliasResponse = strings.TrimSpace(strings.ToLower(aliasResponse))
	if aliasResponse == "" || aliasResponse == "yes" || aliasResponse == "y" {
		fmt.Print("Enter the alias name you would like to use (default 'sgpt'): ")
		aliasName, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		aliasName = strings.TrimSpace(aliasName)
		if aliasName == "" {
			aliasName = "sgpt"
		}
		err = setupShellAlias(aliasName)
		if err != nil {
			return err
		}
	}

	fmt.Println("First-time setup is complete. Please restart your terminal or source your profile to apply changes.")
	return nil
}

func setupEnvironmentVariables() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your OpenAI API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("Enter a passphrase for encrypting conversation history: ")
	passphrase, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	passphrase = strings.TrimSpace(passphrase)

	// Set environment variables in the current process
	os.Setenv("OPENAI_API_KEY", apiKey)
	os.Setenv("IVYCLI_PASSPHRASE", passphrase)

	// Write environment variables to shell profile
	shell := os.Getenv("SHELL")
	var profilePath string
	if strings.Contains(shell, "bash") {
		profilePath = filepath.Join(os.Getenv("HOME"), ".bashrc")
	} else if strings.Contains(shell, "zsh") {
		profilePath = filepath.Join(os.Getenv("HOME"), ".zshrc")
	} else {
		profilePath = filepath.Join(os.Getenv("HOME"), ".profile")
	}

	envVars := fmt.Sprintf("\nexport OPENAI_API_KEY=\"%s\"\nexport IVYCLI_PASSPHRASE=\"%s\"\n", apiKey, passphrase)

	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(envVars)
	if err != nil {
		return err
	}

	fmt.Printf("Environment variables added to %s.\n", profilePath)
	return nil
}

func setupShellAlias(aliasName string) error {
	shell := os.Getenv("SHELL")
	var profilePath string
	if strings.Contains(shell, "bash") {
		profilePath = filepath.Join(os.Getenv("HOME"), ".bashrc")
	} else if strings.Contains(shell, "zsh") {
		profilePath = filepath.Join(os.Getenv("HOME"), ".zshrc")
	} else {
		profilePath = filepath.Join(os.Getenv("HOME"), ".profile")
	}

	aliasFunction := fmt.Sprintf(`
%s() {
  local args=()
  while [[ "$1" == -* ]]; do
    args+=("$1")
    shift
  done
  printf '%%s\n' "$*" | IvyCLI "${args[@]}"
}
`, aliasName)

	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(aliasFunction)
	if err != nil {
		return err
	}

	fmt.Printf("Alias '%s' added to %s.\n", aliasName, profilePath)
	return nil
}

func handlePrompt(prompt string, messages *[]map[string]string, apiKey, passphrase, model string, maxHistorySize int, enableMarkdown, noHistory bool, timeout int) error {
	*messages = append(*messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": *messages,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		if errMsg, ok := errorResponse["error"].(map[string]interface{}); ok {
			return fmt.Errorf("error: %s", errMsg["message"])
		}
		return fmt.Errorf("error: received non-200 response status: %s", resp.Status)
	}

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	if choices, ok := responseBody["choices"].([]interface{}); ok && len(choices) > 0 {
		if message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{}); ok {
			content := strings.TrimSpace(message["content"].(string))

			if enableMarkdown {
				printWithMarkdown(content)
			} else {
				fmt.Println(content)
			}

			*messages = append(*messages, map[string]string{
				"role":    "assistant",
				"content": content,
			})

			if !noHistory {
				var historyMessages []map[string]string
				for _, msg := range *messages {
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

			return nil
		}
	}

	return fmt.Errorf("error: unexpected response format")
}
