package analyzer

import (
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
	if hasFile(repoRoot, "go.work") || hasFile(repoRoot, "pnpm-workspace.yaml") {
		// Optimization placeholder for production
	}

	// Level 1 Scan (Root)
	result = appendDetails(result, ScanDirectory(repoRoot))

	// Level 2 (one folder down)
	enteries, err := os.ReadDir(repoRoot)

	if err == nil {
		for _, entry := range enteries {
			// Blacklist enforcement (node modules etc)
			if entry.IsDir() && entry.Name() != "node_modules" && entry.Name() != ".git" && entry.Name() != "vendor" {
				subDirPath := filepath.Join(repoRoot, entry.Name())
				result = appendDetails(result, ScanDirectory(subDirPath))
			}
		}
	}
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
