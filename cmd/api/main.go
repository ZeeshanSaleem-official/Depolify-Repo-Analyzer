package main

import (
	"depolify-analyzer/pkg/analyzer"
	"encoding/json"
	"fmt"
)

func main() {
	testRepoPath := "../../test-repos"

	fmt.Println("Starting DEPOLIFY Repository Extraction...")

	extractedData, err := analyzer.ExtractDetails(testRepoPath)
	if err != nil {
		fmt.Println("Extraction failed:", err)
		return
	}

	// Condition B: Missing Component
	if len(extractedData.Frontends) == 0 || len(extractedData.Backends) == 0 {
		fmt.Println("🚨 STAGE 4 ERROR: Incomplete Architecture! DEPOLIFY requires at least 1 Frontend and 1 Backend.")
		return
	}

	// Condition C: Multiple Candidates (The Dashboard Fallback)
	if len(extractedData.Frontends) > 1 {
		fmt.Println("🚨 STAGE 4 CONFLICT: Multiple UI projects detected!")
		fmt.Println("-> Pausing backend deployment.")
		fmt.Println("-> Sending JSON payload to Next.js Dashboard to ask user which Frontend to use...")

		jsonData, _ := json.MarshalIndent(extractedData.Frontends, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	fmt.Println("✅ STAGE 4 PASSED: Perfect Match Found.")

	fmt.Println("\n--- Final Deployment Blueprint ---")

	// Create a structured payload showing exactly what is being sent to the DockerService
	finalBlueprint := map[string]interface{}{
		"frontend": extractedData.Frontends[0],
		"backend":  extractedData.Backends[0],
	}

	// Convert it into readable JSON
	successJson, _ := json.MarshalIndent(finalBlueprint, "", "  ")
	fmt.Println(string(successJson))

	fmt.Println("\n-> Handing off to DockerService to build images...")
}
