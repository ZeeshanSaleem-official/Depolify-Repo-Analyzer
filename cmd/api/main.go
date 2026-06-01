package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"depolify-analyzer/pkg/analyzer"
)

func main() {
	targetDir := "./deploify-poison-repo"
	absPath, _ := filepath.Abs(targetDir)
	
	fmt.Println("Scanning repository at:", absPath)

	extractedData, err := analyzer.ExtractDetails(targetDir)
	if err != nil {
		fmt.Println("Extraction failed:", err)
		return
	}

	frontCount := len(extractedData.Frontends)
	backCount := len(extractedData.Backends)

	// Handle shared-root conflicts where multiple frameworks coexist.
	if len(extractedData.Conflicts) > 0 {
		fmt.Printf("Shared-root conflict detected in %d directory(ies).\n", len(extractedData.Conflicts))
		for _, c := range extractedData.Conflicts {
			fmt.Printf(" - %s: %s\n", c.Directory, c.Description)
		}
		printJSON(extractedData)
		return
	}

	// Handle empty or unsupported repository scenarios.
	if frontCount == 0 && backCount == 0 {
		unknownResponse := map[string]interface{}{
			"status":       "error",
			"project_type": "Unknown",
			"message":      "Could not detect a supported framework. Ensure a valid configuration file exists.",
		}

		jsonBytes, _ := json.MarshalIndent(unknownResponse, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	// Handle multiple frontend services.
	if frontCount > 1 {
		fmt.Printf("Conflict detected: Found %d frontends.\n", frontCount)
		printJSON(extractedData)
		return
	}

	// Handle multiple backend services.
	if backCount > 1 {
		fmt.Printf("Conflict detected: Found %d backends.\n", backCount)
		printJSON(extractedData)
		return
	}

	// Process successful single frontend and/or single backend detection.
	successJson, _ := json.MarshalIndent(extractedData, "", "  ")
	fmt.Println(string(successJson))
}

func printJSON(data interface{}) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonBytes))
}
