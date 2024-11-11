package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func loadConversationHistory(passphrase string) ([]map[string]string, error) {
	var messages []map[string]string

	historyFile := getHistoryFilePath()
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	decryptedData, err := decrypt(data, passphrase)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decryptedData, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func saveConversationHistory(messages []map[string]string, passphrase string) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	encryptedData, err := encrypt(data, passphrase)
	if err != nil {
		return err
	}

	historyFile := getHistoryFilePath()
	err = os.WriteFile(historyFile, encryptedData, 0600)
	if err != nil {
		return err
	}
	return nil
}

func resetConversationHistory() error {
	historyFile := getHistoryFilePath()
	err := os.Remove(historyFile)
	if err != nil && !os.IsNotExist(err) {
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
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating config directory:", err)
		os.Exit(1)
	}
	historyFile := filepath.Join(configDir, "history.enc")
	return historyFile
}
