package main

import (
	"fmt"
	"fw/modules"
	"log"
	"path/filepath"
	"time"
)

func main() {
	// Parse command line arguments
	args, err := modules.ParseArgs()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Set verbose mode globally
	modules.SetVerbose(args.Verbose)

	// Get arguments
	filePath := args.FilePath
	encryptionKey := args.EncryptionKey
	baseURL := args.BaseURL

	fmt.Println("Encrypting file...")
	encryptedData, err := modules.EncryptFile(filePath, encryptionKey)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	// Get filename without path
	fileName := filepath.Base(filePath)

	// Calculate unique ID for this file transfer
	transferID := modules.CalculateMD5(encryptedData)

	// Split hex string into smaller chunks
	chunkSize := modules.CalculateOptimalChunkSize(encryptedData)
	chunks := modules.SplitHexString(encryptedData, chunkSize)
	totalChunks := len(chunks)
	fmt.Printf("File: %s (Transfer ID: %s)\n", fileName, transferID)
	fmt.Printf("Split into %d chunks (chunk size: %d chars)\n", totalChunks, chunkSize)

	// Send initialization request
	fmt.Print("Initializing transfer... ")
	initPath := fmt.Sprintf("init/%s/%d/%d/%s",
		transferID, totalChunks, len(encryptedData), fileName)

	modules.DebugPrintf("Init path length: %d characters\n", len(initPath))

	err = modules.SendRequest(baseURL, initPath)
	if err != nil {
		fmt.Println("Failed!")
		log.Fatalf("Failed to initialize transfer: %v", err)
	}
	fmt.Println("Done!")

	// Process each chunk
	transferStartTime := time.Now()

	// Initialize progress bar
	modules.UpdateProgressInPlace(0, totalChunks, transferStartTime)

	for i, chunk := range chunks {
		// Create path for this chunk
		chunkChecksum := modules.CalculateMD5(chunk)
		chunkPath := fmt.Sprintf("chunk/%s/%d/%s/%s",
			transferID, i, chunkChecksum, chunk)

		if modules.Verbose && len(chunkPath) > modules.MaxPayloadLength {
			fmt.Printf("\nWarning: Chunk %d path exceeds maximum length: %d > %d\n",
				i+1, len(chunkPath), modules.MaxPayloadLength)
		}

		// Try to send with retries
		maxRetries := 3
		success := false

		for attempt := 1; attempt <= maxRetries; attempt++ {
			modules.UpdateProgressInPlace(i+1, totalChunks, transferStartTime)

			err := modules.SendRequest(baseURL, chunkPath)
			if err == nil {
				success = true
				break
			}

			if attempt < maxRetries {
				modules.DebugPrintf("\nRetry %d/%d: %v\n", attempt, maxRetries, err)
				modules.UpdateProgressInPlace(i+1, totalChunks, transferStartTime)
				time.Sleep(time.Duration(attempt) * time.Second)
			} else {
				modules.UpdateProgressInPlace(i+1, totalChunks, transferStartTime)
			}
		}

		if !success {
			fmt.Println("\nFailed to send chunk after multiple attempts")
			return
		}

		// Update progress
		modules.UpdateProgressInPlace(i+1, totalChunks, transferStartTime)

		// Small delay between chunks
		if i < totalChunks-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Final progress update
	modules.UpdateProgressInPlace(totalChunks, totalChunks, transferStartTime)
	fmt.Println()

	// Send completion request
	fmt.Print("Finalizing transfer... ")
	completePath := fmt.Sprintf("complete/%s/%s",
		transferID, modules.CalculateMD5(encryptedData))

	err = modules.SendRequest(baseURL, completePath)
	if err != nil {
		fmt.Println("Failed!")
		log.Printf("Warning: Failed to send completion notification: %v", err)
	} else {
		fmt.Println("Done!")
	}

	// Print transfer summary
	totalDuration := time.Since(transferStartTime)
	fmt.Printf("\nFile %s uploaded successfully!\n", fileName)
	fmt.Printf("Total transfer time: %s\n", modules.FormatDuration(totalDuration))
}
