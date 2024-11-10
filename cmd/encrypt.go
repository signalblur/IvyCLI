// encrypt.go
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

// Encryption functions

func encrypt(data []byte) ([]byte, error) {
	passphrase := getPassphrase()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(salt, ciphertext...)
	zeroBytes([]byte(passphrase))
	return ciphertext, nil
}

func decrypt(data []byte) ([]byte, error) {
	passphrase := getPassphrase()
	if len(data) < 16 {
		return nil, fmt.Errorf("data too short")
	}
	salt := data[:16]
	ciphertext := data[16:]
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	zeroBytes([]byte(passphrase))
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func getPassphrase() string {
	fmt.Print("Enter passphrase for conversation history encryption: ")
	passphraseBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading passphrase:", err)
		os.Exit(1)
	}
	passphrase := string(passphraseBytes)
	zeroBytes(passphraseBytes)
	return passphrase
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
