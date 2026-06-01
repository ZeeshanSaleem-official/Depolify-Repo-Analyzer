package analyzer

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// groupByDirectory detects shared-root conflicts where multiple frameworks
// coexist in the same directory, which may cause ambiguous deployments.
func groupByDirectory(repo *ExtractedRepo) []Conflict {
	// Group services by their absolute directory path.
	dirMap := make(map[string][]DeploymentDetails)

	for _, f := range repo.Frontends {
		dirMap[f.AbsolutePath] = append(dirMap[f.AbsolutePath], f)
	}
	for _, b := range repo.Backends {
		dirMap[b.AbsolutePath] = append(dirMap[b.AbsolutePath], b)
	}

	// Identify directories containing multiple services.
	var conflicts []Conflict
	for dir, services := range dirMap {
		if len(services) > 1 {
			// Collect service types for conflict description.
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

// ResolveConflicts processes scan results to handle port contentions
// and assign unique container names for shared-root monorepos.
func ResolveConflicts(repo *ExtractedRepo) *ExtractedRepo {
	// Detect shared-root conflicts.
	sharedRootConflicts := groupByDirectory(repo)

	// Map conflicted directories for quick lookup.
	conflictedDirs := make(map[string]bool)
	for _, c := range sharedRootConflicts {
		conflictedDirs[c.Directory] = true
	}

	usedPorts := make(map[string]bool)

	// assignUniquePort resolves port collisions by incrementing the port number.
	assignUniquePort := func(requestedPort string) string {
		port, err := strconv.Atoi(requestedPort)
		if err != nil {
			return requestedPort // Fallback if port isn't a simple number
		}

		// Increment port number until an unused port is found.
		for usedPorts[strconv.Itoa(port)] {
			port++
		}

		finalPort := strconv.Itoa(port)
		usedPorts[finalPort] = true
		return finalPort
	}

	// generateName creates a unique identifier based on directory and project type.
	generateName := func(absPath string, projType ProjectType) string {
		dirName := filepath.Base(absPath)
		if dirName == "." || dirName == "/" || dirName == "" {
			dirName = "root"
		}
		// Ensure the project type string is formatted for container names.
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
