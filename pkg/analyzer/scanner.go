package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExtractDetails
func ExtractDetails(repoRoot string) (*ExtractedRepo, error) {
	result := &ExtractedRepo{
		Frontends: []DeploymentDetails{},
		Backends:  []DeploymentDetails{},
	}

	// STAGE 1: The Configuration Check
	if hasFile(repoRoot, "go.work") || hasFile(repoRoot, "pnpm-workspace.yaml") || hasFile(repoRoot, "lerna.json") {
		// In the future, you could parse these files to find exact workspace paths!
	}

	// STAGE 2: Deep Recursive Scan
	// We replace the hardcoded "Level 2" with a smart, recursive folder walker
	filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("🚨 SCANNER ERROR: Cannot access path: %s\n", path)
			return nil // Skip errors and keep scanning
		}

		// Blacklist Enforcement: If it's a directory we don't want to scan, tell Go to skip it entirely!
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "dist" || d.Name() == ".next" {
				return filepath.SkipDir // 🚨 Crucial to prevent infinite loops and lag!
			}
		}

		// If it's a directory, run our Strategy Detectors on it!
		if d.IsDir() {
			result = appendDetails(result, ScanDirectory(path))
		}

		return nil
	})

	return result, nil
}

// scanDirectory checks for signature files
func ScanDirectory(dirPath string) *ExtractedRepo {
	tempResult := &ExtractedRepo{
		Frontends: []DeploymentDetails{},
		Backends:  []DeploymentDetails{},
	}

	// 1. Run all Frontend Detectors
	for _, detect := range frontendDetectors {
		details := detect(dirPath)
		if details != nil {
			tempResult.Frontends = append(tempResult.Frontends, *details)
			break
		}
	}
	// 2. Run all Backend Detectors
	for _, detect := range backendDetectors {
		details := detect(dirPath)
		if details != nil {
			tempResult.Backends = append(tempResult.Backends, *details)
			break
		}
	}
	return tempResult

}

// UTILITY HELPERS

// appendDetails merges the temporary scan results into the main result struct

func appendDetails(main *ExtractedRepo, new *ExtractedRepo) *ExtractedRepo {
	main.Frontends = append(main.Frontends, new.Frontends...)
	main.Backends = append(main.Backends, new.Backends...)
	return main
}
