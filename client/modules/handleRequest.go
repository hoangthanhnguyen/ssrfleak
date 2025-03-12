package modules

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// URLSuffix is the suffix added to all request URLs
const URLSuffix = "/@v/v1.info"

// Global verbose flag
var Verbose bool = false

// SetVerbose sets the verbose flag globally
func SetVerbose(verbose bool) {
	Verbose = verbose
}

// DebugPrintf prints debug messages when verbose mode is enabled
func DebugPrintf(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf(format, args...)
	}
}

// SendRequest sends an HTTP GET request to baseURL + path + URLSuffix
func SendRequest(baseURL string, path string) error {
	// Ensure baseURL ends with a slash
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// Remove leading slash from path if it exists
	path = strings.TrimPrefix(path, "/")

	// Check if payload length exceeds limit
	if len(path) > MaxPayloadLength {
		return fmt.Errorf("payload length (%d) exceeds maximum allowed length (%d)", len(path), MaxPayloadLength)
	}

	// Create full URL by concatenating base URL, path and suffix
	fullURL := baseURL + path + URLSuffix

	// Debug output in verbose mode
	DebugPrintf("\nURL: %s\n", fullURL)

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		// Disable automatic redirects to prevent HTTP->HTTPS conversion
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Debug output in verbose mode
			if len(via) > 0 {
				DebugPrintf("Redirect detected: %s -> %s\n",
					via[len(via)-1].URL.String(), req.URL.String())
			}
			// Return error to prevent following redirects
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	DebugPrintf("Response: %s\n", resp.Status)
	resp.Body.Close()

	return nil
}
