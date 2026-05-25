package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"depolify-analyzer/pkg/analyzer"
)

func main() {
	// TARGET DIRECTORY
	targetDir := "../../test-repos/dummy-express"

	absPath, _ := filepath.Abs(targetDir)
	fmt.Println("🔍 Scanning repository at:", absPath)
	fmt.Println("🚀 Running DEPOLIFY Extraction Engine...\n")

	// Run the extraction engine
	extractedData, err := analyzer.ExtractDetails(targetDir)
	if err != nil {
		fmt.Println("❌ Extraction CRASHED:", err)
		return
	}

	// ==========================================
	// STAGE 4: FAIL-SAFE & CONFLICT RESOLUTION
	// ==========================================
	frontCount := len(extractedData.Frontends)
	backCount := len(extractedData.Backends)

	// SCENARIO 1: Empty or Unsupported Repository
	if frontCount == 0 && backCount == 0 {
		fmt.Println("❌ STAGE 4 FAILED: No recognizable frameworks found.")

		// Package it as an official API response for your Next.js Dashboard
		unknownResponse := map[string]interface{}{
			"status":       "error",
			"project_type": "Unknown",
			"message":      "DEPOLIFY could not detect a supported framework. Please ensure you have a valid package.json, go.mod, or requirements.txt.",
		}

		jsonBytes, _ := json.MarshalIndent(unknownResponse, "", "  ")
		fmt.Println(string(jsonBytes))

		return // Safely halt the Docker hand-off
	}

	// SCENARIO 2: Frontend Conflict (Monorepo with too many options)
	if frontCount > 1 {
		fmt.Printf("⚠️ CONFLICT DETECTED: Found %d Frontends!\n", frontCount)
		fmt.Println("-> AUTOMATION HALTED: Sending data to Next.js UI for user selection.")
		printJSON(extractedData)
		return
	}

	// SCENARIO 3: Backend Conflict (Monorepo with too many APIs)
	if backCount > 1 {
		fmt.Printf("⚠️ CONFLICT DETECTED: Found %d Backends!\n", backCount)
		fmt.Println("-> AUTOMATION HALTED: Sending data to Next.js UI for user selection.")
		printJSON(extractedData)
		return
	}

	// SCENARIO 4: The Perfect Match (1 Frontend and/or 1 Backend)
	fmt.Println("✅ STAGE 4 PASSED: Perfect Match Found.")
	fmt.Println("\n--- Final Deployment Blueprint ---")

	// Convert the perfect match into the final JSON payload
	successJson, _ := json.MarshalIndent(extractedData, "", "  ")
	fmt.Println(string(successJson))

	fmt.Println("\n-> Handing off to DockerService to build images...")
}

// printJSON is a helper to output the options when a conflict occurs
func printJSON(data interface{}) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println("\n--- Available Deployment Options ---")
	fmt.Println(string(jsonBytes))
}
