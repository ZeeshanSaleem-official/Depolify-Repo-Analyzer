package analyzer

import (
	"os"
	"path/filepath"
)

// ExtractDetails scans the repository and identifies all deployable services.
func ExtractDetails(repoRoot string) (*ExtractedRepo, error) {
	result := &ExtractedRepo{
		Frontends: []DeploymentDetails{},
		Backends:  []DeploymentDetails{},
	}

	// Detect workspace configurations.
	if hasFile(repoRoot, "go.work") || hasFile(repoRoot, "pnpm-workspace.yaml") || hasFile(repoRoot, "lerna.json") {
		// Workspace parsing can be implemented here.
	}

	// Recursively scan the repository.
	filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories to optimize scanning.
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".git", "vendor", "dist", ".next":
				return filepath.SkipDir
			}
		}

		// Analyze directory for supported frameworks.
		if d.IsDir() {
			result = appendDetails(result, ScanDirectory(path))
		}

		return nil
	})

	resolvedResult := ResolveConflicts(result)

	return resolvedResult, nil
}

// ScanDirectory checks a directory for supported frontend and backend frameworks.
func ScanDirectory(dirPath string) *ExtractedRepo {
	tempResult := &ExtractedRepo{
		Frontends: []DeploymentDetails{},
		Backends:  []DeploymentDetails{},
	}

	// Execute frontend detectors.
	for _, detect := range frontendDetectors {
		details := detect(dirPath)
		if details != nil {
			tempResult.Frontends = append(tempResult.Frontends, *details)
			break
		}
	}
	// Execute backend detectors.
	for _, detect := range backendDetectors {
		details := detect(dirPath)
		if details != nil {
			tempResult.Backends = append(tempResult.Backends, *details)
			break
		}
	}
	return tempResult
}

// appendDetails merges the temporary scan results into the main result struct.
func appendDetails(main *ExtractedRepo, new *ExtractedRepo) *ExtractedRepo {
	main.Frontends = append(main.Frontends, new.Frontends...)
	main.Backends = append(main.Backends, new.Backends...)
	return main
}
