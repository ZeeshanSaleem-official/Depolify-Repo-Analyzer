package analyzer

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// groupByDirectory detects Shared-Root Conflicts — the scenario where
// multiple frameworks (e.g., go.mod + package.json) coexist in the
// same directory, causing platforms like Vercel to silently pick one.
func groupByDirectory(repo *ExtractedRepo) []Conflict {
	// 1. Merge all services into a single view, keyed by directory
	dirMap := make(map[string][]DeploymentDetails)

	for _, f := range repo.Frontends {
		dirMap[f.AbsolutePath] = append(dirMap[f.AbsolutePath], f)
	}
	for _, b := range repo.Backends {
		dirMap[b.AbsolutePath] = append(dirMap[b.AbsolutePath], b)
	}

	// 2. Any directory with > 1 service is a Shared-Root Conflict
	var conflicts []Conflict
	for dir, services := range dirMap {
		if len(services) > 1 {
			// Build a human-readable list of what collided
			var types []string
			for _, s := range services {
				types = append(types, string(s.Type))
			}

			conflicts = append(conflicts, Conflict{
				Type:      ConflictSharedRoot,
				Directory: dir,
				Services:  services,
				Description: fmt.Sprintf(
					"Directory contains %d co-located services (%s). "+
						"Platforms like Vercel will silently deploy only one.",
					len(services),
					strings.Join(types, " + "),
				),
			})
		}
	}

	return conflicts
}

// ResolveConflicts intercepts the scan results to fix Port Contentions
// and assign unique container names for Shared-Root Monorepos.
func ResolveConflicts(repo *ExtractedRepo) *ExtractedRepo {
	// PHASE 1: Detect Shared-Root Conflicts
	// This catches the Vercel failure mode — co-located go.mod + package.json
	sharedRootConflicts := groupByDirectory(repo)

	// Build a lookup of directories involved in shared-root conflicts
	// so individual service objects can self-identify as conflicted.
	conflictedDirs := make(map[string]bool)
	for _, c := range sharedRootConflicts {
		conflictedDirs[c.Directory] = true
	}

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
		if conflictedDirs[f.AbsolutePath] {
			f.ConflictRef = f.AbsolutePath
		}
		resolvedFrontends = append(resolvedFrontends, f)
	}

	var resolvedBackends []DeploymentDetails
	for _, b := range repo.Backends {
		b.Name = generateName(b.AbsolutePath, b.Type)
		b.DefaultPort = assignUniquePort(b.DefaultPort)
		if conflictedDirs[b.AbsolutePath] {
			b.ConflictRef = b.AbsolutePath
		}
		resolvedBackends = append(resolvedBackends, b)
	}

	return &ExtractedRepo{
		Frontends: resolvedFrontends,
		Backends:  resolvedBackends,
		Conflicts: sharedRootConflicts,
	}
}

