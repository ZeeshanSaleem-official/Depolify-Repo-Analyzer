package analyzer

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ResolveConflicts intercepts the scan results to fix Port Contentions
// and assign unique container names for Shared-Root Monorepos.
func ResolveConflicts(repo *ExtractedRepo) *ExtractedRepo {
	usedPorts := make(map[string]bool)

	// HELPER 1: Dynamically assign a free port if there is a collision
	assignUniquePort := func(requestedPort string) string {
		port, err := strconv.Atoi(requestedPort)
		if err != nil {
			return requestedPort // Fallback if port isn't a simple number
		}

		// If port 3000 is taken, bump to 3001, 3002, etc.
		for usedPorts[strconv.Itoa(port)] {
			port++
		}

		finalPort := strconv.Itoa(port)
		usedPorts[finalPort] = true
		return finalPort
	}

	// HELPER 2: Generate unique container names (e.g., "root-Go", "root-Next.js")
	generateName := func(absPath string, projType ProjectType) string {
		dirName := filepath.Base(absPath)
		if dirName == "." || dirName == "/" || dirName == "" {
			dirName = "root"
		}
		// Clean the string so it's Docker-safe
		safeType := strings.ReplaceAll(string(projType), " ", "-")
		return fmt.Sprintf("%s-%s", dirName, safeType)
	}

	var resolvedFrontends []DeploymentDetails
	for _, f := range repo.Frontends {
		f.Name = generateName(f.AbsolutePath, f.Type)
		f.DefaultPort = assignUniquePort(f.DefaultPort)
		resolvedFrontends = append(resolvedFrontends, f)
	}

	var resolvedBackends []DeploymentDetails
	for _, b := range repo.Backends {
		b.Name = generateName(b.AbsolutePath, b.Type)
		b.DefaultPort = assignUniquePort(b.DefaultPort)
		resolvedBackends = append(resolvedBackends, b)
	}

	return &ExtractedRepo{
		Frontends: resolvedFrontends,
		Backends:  resolvedBackends,
	}
}
