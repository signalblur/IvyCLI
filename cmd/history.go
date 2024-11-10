package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func loadConversationHistory() ([]map[string]string, error) {
	var messages []map[string]string

	historyFile := getHistoryFilePath()
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	decryptedData, err := decrypt(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decryptedData, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func saveConversationHistory(messages []map[string]string) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	// Encrypt the data
	encryptedData, err := encrypt(data)
	if err != nil {
		return err
	}

	historyFile := getHistoryFilePath()
	err = os.WriteFile(historyFile, encryptedData, 0600) // File permissions set to owner read/write
	if err != nil {
		return err
	}
	return nil
}

func getHistoryFilePath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting current user:", err)
		os.Exit(1)
	}
	configDir := filepath.Join(usr.HomeDir, ".config", "ivycli")
	os.MkdirAll(configDir, 0700) // Ensure directory exists with restricted permissions
	historyFile := filepath.Join(configDir, "history.enc")
	return historyFile
}
