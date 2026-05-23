package main

import (
	"encoding/json"
	"fmt"

	// Make sure this matches your module name in go.mod exactly!
	"depolify-analyzer/pkg/analyzer"
)

func main() {
	// Point this to a valid path on your local machine to test
	testRepoPath := "../../test-repos"

	fmt.Println("Starting DEPOLIFY Repository Extraction...")

	extractedData, err := analyzer.ExtractDetails(testRepoPath)
	if err != nil {
		fmt.Println("Extraction failed:", err)
		return
	}

	// ==========================================
	// STAGE 4: State Validation & The Fail-Safe
	// ==========================================

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

	// Condition A: The Perfect Match
	fmt.Println("✅ STAGE 4 PASSED: Perfect Match Found.")

	fmt.Println("\n--- Final Deployment Blueprint ---")
	fmt.Printf("Frontend: %s at %s\n", extractedData.Frontends[0].Type, extractedData.Frontends[0].AbsolutePath)
	fmt.Printf("Backend: %s at %s\n", extractedData.Backends[0].Type, extractedData.Backends[0].AbsolutePath)
	fmt.Println("-> Handing off to DockerService to build images...")
}
