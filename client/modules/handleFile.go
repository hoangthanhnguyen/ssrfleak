package modules

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// MaxPayloadLength is the maximum allowed length for request path
const MaxPayloadLength = 2000

// SplitHexString splits a hex string into chunks of specified size
func SplitHexString(hexString string, chunkSize int) []string {
	var chunks []string
	strLen := len(hexString)

	for i := 0; i < strLen; i += chunkSize {
		end := i + chunkSize
		if end > strLen {
			end = strLen
		}
		chunks = append(chunks, hexString[i:end])
	}

	return chunks
}

// CalculateMD5 computes MD5 hash of a string
func CalculateMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateOptimalChunkSize determines best chunk size based on data length
// while ensuring path length doesn't exceed MaxPayloadLength
func CalculateOptimalChunkSize(data string) int {
	dataLen := len(data)

	// Calculate maximum chunk size based on MaxPayloadLength
	// Format: chunk/transferID/index/checksum/data
	// Estimate overhead for path components (excluding data)
	// transferID = 32 chars (MD5)
	// index = ~10 chars max (for large files)
	// checksum = 32 chars (MD5)
	// separators = ~10 chars
	const pathOverhead = 84 // "chunk/" + transferID + "/" + index + "/" + checksum + "/"
	maxChunkSize := MaxPayloadLength - pathOverhead

	// Ensure maxChunkSize is positive
	if maxChunkSize <= 0 {
		return 500 // Fallback to small size if constraints are too tight
	}

	// For very small files (< 10KB), use smaller chunks
	if dataLen < 20000 {
		return Min(1000, maxChunkSize)
	}

	// For small files (< 100KB), use moderate chunks
	if dataLen < 200000 {
		return Min(1500, maxChunkSize)
	}

	// For medium files (< 1MB)
	if dataLen < 2000000 {
		return Min(5000, maxChunkSize)
	}

	// For large files (< 10MB)
	if dataLen < 20000000 {
		return Min(10000, maxChunkSize)
	}

	// For very large files
	return Min(15000, maxChunkSize)
}

// Min returns the smaller of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// EncryptFile reads a file, encrypts it with AES-256 using a provided key string,
// and returns the encrypted data as a hex string
func OriginalEncryptFile(filePath string, keyString string) (string, error) {
	// Read file to encrypt
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Convert string key to 32-byte key using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(keyString))
	key := hasher.Sum(nil)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode which provides authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Create nonce (number used once)
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt and authenticate data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Convert to hex
	return hex.EncodeToString(ciphertext), nil
}

// EncryptFile is a wrapper for the actual EncryptFile function in your existing module
// This is just a placeholder that calls the real function
func EncryptFile(filePath, key string) (string, error) {
	// Assuming the actual implementation is already in fw/modules
	// and we're just providing a consistent interface here
	return OriginalEncryptFile(filePath, key)
}
