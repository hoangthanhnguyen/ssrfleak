package server

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Global verbose mode flag
var verboseMode bool = false

// FileTransfer represents an in-progress file transfer
type FileTransfer struct {
	ID          string
	Filename    string
	TotalChunks int
	FileSize    int
	Chunks      map[int]string // Map of chunk index to chunk data
	Created     time.Time
	LastUpdated time.Time
	Checksum    string
}

// In-memory storage for file transfers
var (
	transfers      = make(map[string]*FileTransfer)
	transfersMutex sync.RWMutex
)

// SetVerboseMode enables or disables verbose logging
func SetVerboseMode(verbose bool) {
	verboseMode = verbose
	if verbose {
		log.Printf("Verbose logging enabled")
	}
}

// IsVerboseMode returns the current verbose mode state
func IsVerboseMode() bool {
	return verboseMode
}

// HandleRequest processes incoming GET requests for file transfer
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Parse path components
	path := strings.Trim(r.URL.Path, "/")
	components := strings.Split(path, "/")

	if len(components) < 2 {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Determine request type
	action := components[0]

	switch action {
	case "init":
		handleInitRequest(w, r, components)
	case "chunk":
		handleChunkRequest(w, r, components)
	case "complete":
		handleCompleteRequest(w, r, components)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

// handleInitRequest processes initialization requests
func handleInitRequest(w http.ResponseWriter, r *http.Request, components []string) {
	// Expected format: init/transferID/totalChunks/fileSize/filename
	if len(components) != 5 {
		http.Error(w, "Invalid init request format", http.StatusBadRequest)
		return
	}

	transferID := components[1]
	totalChunks, err := strconv.Atoi(components[2])
	if err != nil {
		http.Error(w, "Invalid total chunks", http.StatusBadRequest)
		return
	}

	fileSize, err := strconv.Atoi(components[3])
	if err != nil {
		http.Error(w, "Invalid file size", http.StatusBadRequest)
		return
	}

	filename := components[4]

	// Create new transfer record
	transfer := &FileTransfer{
		ID:          transferID,
		Filename:    filename,
		TotalChunks: totalChunks,
		FileSize:    fileSize,
		Chunks:      make(map[int]string),
		Created:     time.Now(),
		LastUpdated: time.Now(),
	}

	// Store the transfer
	transfersMutex.Lock()
	transfers[transferID] = transfer
	transfersMutex.Unlock()

	log.Printf("Initialized transfer %s for file '%s': expecting %d chunks, %d bytes",
		transferID, filename, totalChunks, fileSize)

	// Respond with success
	fmt.Fprint(w, "OK")
}

// handleChunkRequest processes incoming chunk data
func handleChunkRequest(w http.ResponseWriter, r *http.Request, components []string) {
	// Expected format: chunk/transferID/index/checksum/data
	if len(components) != 5 {
		http.Error(w, "Invalid chunk request format", http.StatusBadRequest)
		return
	}

	transferID := components[1]
	chunkIndex, err := strconv.Atoi(components[2])
	if err != nil {
		http.Error(w, "Invalid chunk index", http.StatusBadRequest)
		return
	}

	chunkChecksum := components[3]
	chunkData := components[4]

	// Verify the chunk data with checksum
	if calculateMD5(chunkData) != chunkChecksum {
		http.Error(w, "Checksum verification failed", http.StatusBadRequest)
		return
	}

	// Find the transfer record
	transfersMutex.Lock()
	transfer, exists := transfers[transferID]
	if !exists {
		transfersMutex.Unlock()
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	// Store the chunk
	transfer.Chunks[chunkIndex] = chunkData
	transfer.LastUpdated = time.Now()
	transfersMutex.Unlock()

	if verboseMode {
		log.Printf("Received chunk %d/%d for transfer %s (file: %s)",
			chunkIndex+1, transfer.TotalChunks, transferID, transfer.Filename)
	}

	// Respond with success
	fmt.Fprint(w, "OK")
}

// handleCompleteRequest processes completion requests and concatenates all chunks
func handleCompleteRequest(w http.ResponseWriter, r *http.Request, components []string) {
	// Debug: In ra thông tin gói complete nhận được
	if verboseMode {
		log.Printf("DEBUG: Received complete request: %s", strings.Join(components, "/"))
	}

	// Expected format: complete/transferID/checksum
	if len(components) != 3 {
		log.Printf("ERROR: Invalid complete request format, got %d components instead of 3", len(components))
		http.Error(w, "Invalid complete request format", http.StatusBadRequest)
		return
	}

	transferID := components[1]
	expectedChecksum := components[2]

	if verboseMode {
		log.Printf("DEBUG: Processing complete request for transferID=%s with checksum=%s",
			transferID, expectedChecksum)
	}

	// Find the transfer record
	transfersMutex.Lock()
	transfer, exists := transfers[transferID]
	if !exists {
		log.Printf("ERROR: Transfer not found with ID: %s", transferID)
		transfersMutex.Unlock()
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	if verboseMode {
		log.Printf("DEBUG: Found transfer record: filename=%s, totalChunks=%d, receivedChunks=%d",
			transfer.Filename, transfer.TotalChunks, len(transfer.Chunks))

		// Debug: Liệt kê các chunk đã nhận
		var missingChunks []int
		for i := 0; i < transfer.TotalChunks; i++ {
			if _, ok := transfer.Chunks[i]; !ok {
				missingChunks = append(missingChunks, i)
			}
		}

		if len(missingChunks) > 0 {
			log.Printf("DEBUG: Missing chunks: %v", missingChunks)
		} else {
			log.Printf("DEBUG: All chunks received successfully")
		}
	}

	// Check if all chunks are received
	if len(transfer.Chunks) != transfer.TotalChunks {
		log.Printf("ERROR: Incomplete transfer, got %d chunks but expected %d",
			len(transfer.Chunks), transfer.TotalChunks)
		transfersMutex.Unlock()
		http.Error(w, fmt.Sprintf("Missing chunks: %d/%d received",
			len(transfer.Chunks), transfer.TotalChunks), http.StatusBadRequest)
		return
	}

	// Concatenate all chunks in correct order
	if verboseMode {
		log.Printf("DEBUG: Starting chunk concatenation for transfer %s", transferID)
	}
	completeData := ConcatenateChunks(transfer)
	if verboseMode {
		log.Printf("DEBUG: Concatenation complete, data size: %d bytes", len(completeData))
	}

	// Verify the full data checksum
	actualChecksum := calculateMD5(completeData)
	if verboseMode {
		log.Printf("DEBUG: Calculated checksum: %s, expected: %s", actualChecksum, expectedChecksum)
	}

	if actualChecksum != expectedChecksum {
		log.Printf("ERROR: Checksum verification failed. Expected: %s, Got: %s",
			expectedChecksum, actualChecksum)
		transfersMutex.Unlock()
		http.Error(w, "Full data checksum verification failed", http.StatusBadRequest)
		return
	}

	log.Printf("DEBUG: Checksum verified successfully")

	// Store the checksum
	transfer.Checksum = expectedChecksum
	transfersMutex.Unlock()

	// Process the completed transfer (save to file)
	if verboseMode {
		log.Printf("DEBUG: Processing completed transfer to save file...")
	}
	outputPath, err := ProcessCompletedTransfer(transferID)
	if err != nil {
		log.Printf("ERROR: Failed to process completed transfer: %v", err)
		http.Error(w, fmt.Sprintf("Error processing transfer: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Transfer %s completed successfully: saved to %s",
		transferID, outputPath)

	// Respond with success
	fmt.Fprint(w, "OK")
	if verboseMode {
		log.Printf("DEBUG: Complete request processed successfully")
	}
}

// calculateMD5 computes MD5 hash of a string
func calculateMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
