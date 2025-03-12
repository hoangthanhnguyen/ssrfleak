package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Global encryption key
var encryptionKey string = "abcd" // Default value

// SetEncryptionKey sets the encryption key from external source
func SetEncryptionKey(key string) {
	encryptionKey = key
	log.Printf("Encryption key has been set")
}

// ConcatenateChunks combines all chunks in a file transfer into a single string
func ConcatenateChunks(transfer *FileTransfer) string {
	var builder strings.Builder
	builder.Grow(transfer.FileSize) // Pre-allocate memory for efficiency

	// Append chunks in order
	for i := 0; i < transfer.TotalChunks; i++ {
		chunk, exists := transfer.Chunks[i]
		if exists {
			builder.WriteString(chunk)
		} else {
			log.Printf("Warning: Missing chunk %d for transfer %s", i, transfer.ID)
		}
	}

	return builder.String()
}

// DecryptAES256 decrypts data that was encrypted with AES-256
func DecryptAES256(encryptedHex string) ([]byte, error) {

	// Convert the hex string to bytes
	encryptedData, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %v", err)
	}

	// Convert string key to 32-byte key using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(encryptionKey)) // Sử dụng biến global
	key := hasher.Sum(nil)

	// Phần còn lại giữ nguyên
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode which provides authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// The nonce is at the beginning of the ciphertext
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt and authenticate data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return plaintext, nil
}

// ProcessCompletedTransfer handles a completed transfer, decrypts the data and saves to a file
func ProcessCompletedTransfer(transferID string) (string, error) {
	transfersMutex.RLock()
	transfer, exists := transfers[transferID]
	transfersMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("transfer not found: %s", transferID)
	}

	// Concatenate all chunks
	hexData := ConcatenateChunks(transfer)

	// Decrypt the data
	outputData, err := DecryptAES256(hexData)
	if err != nil {
		return "", fmt.Errorf("error decrypting data: %v", err)
	}

	if verboseMode {
		log.Printf("Successfully decrypted data for transfer %s", transferID)
	}

	// Create output directory if it doesn't exist
	outputDir := "received_files"
	os.MkdirAll(outputDir, 0755)

	// Use the original filename if available, fallback to transfer ID if not
	var outputFilename string
	if transfer.Filename != "" {
		outputFilename = transfer.Filename
	} else {
		outputFilename = transfer.ID + ".bin"
	}

	// Create output file
	outputPath := filepath.Join(outputDir, outputFilename)
	err = os.WriteFile(outputPath, outputData, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing file: %v", err)
	}

	log.Printf("Processed transfer %s: saved %d bytes to %s",
		transfer.ID, len(outputData), outputPath)

	return outputPath, nil
}

// CleanupTransfer removes a transfer from memory
func CleanupTransfer(transferID string) {
	transfersMutex.Lock()
	delete(transfers, transferID)
	transfersMutex.Unlock()

	if verboseMode {
		log.Printf("Cleaned up transfer %s", transferID)
	}
}

// ScheduleCleanup removes old transfers that haven't been completed
func ScheduleCleanup() {
	// This could run in a goroutine on a timer
	transfersMutex.Lock()
	defer transfersMutex.Unlock()

	now := time.Now()
	timeout := 30 * time.Minute

	for id, transfer := range transfers {
		// Remove transfers that haven't been updated in 30 minutes
		if now.Sub(transfer.LastUpdated) > timeout {
			delete(transfers, id)
			log.Printf("Auto-cleaned up stale transfer %s (file: %s)",
				id, transfer.Filename)
		}
	}
}
