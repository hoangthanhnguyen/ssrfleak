package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	modules "server/modules"
)

func main() {
	// Parse command line flags
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	// Set encryption key
	encryptionKey := "abcdefghi1234567890" // Or from environment variable
	modules.SetEncryptionKey(encryptionKey)

	// Set verbose mode if flag is present
	modules.SetVerboseMode(*verbose)

	// Set up HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Basic request logging
		if modules.IsVerboseMode() {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
		}

		// Handle the request
		modules.HandleRequest(w, r)
	})

	// Start periodic cleanup routine
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			modules.ScheduleCleanup()
		}
	}()

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
