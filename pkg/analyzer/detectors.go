package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings" // Added for our new dynamic parsers
)

// ==========================================
// FRONTEND DETECTORS
// ==========================================

// Detect NextJS
func detectNextJS(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "next") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	rawStartScript := scripts["start"]
	smartPort := extractPort(rawStartScript, "3000")

	return &DeploymentDetails{
		Type:         TypeNextJS,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     "npm start", // Docker will safely run the user's custom script
		DefaultPort:  smartPort,
	}
}

// Detect React
// Detect React
func detectReact(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "react") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	rawStartScript := scripts["start"]
	smartPort := extractPort(rawStartScript, "3000")

	// THE VITE UPGRADE: Dynamically switch the serve folder!
	startCmd := "npx serve -s build"
	if hasFile(dirPath, "vite.config.js") || hasFile(dirPath, "vite.config.ts") {
		startCmd = "npx serve -s dist"
	}

	return &DeploymentDetails{
		Type:         TypeReact,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     startCmd,
		DefaultPort:  smartPort,
	}
}

// Detect Static Html
func detectStaticHTML(dirPath string) *DeploymentDetails {
	if hasFile(dirPath, "index.html") && !hasFile(dirPath, "package.json") {
		return &DeploymentDetails{
			Type:         TypeStaticHTML,
			AbsolutePath: dirPath,
			BuildCmd:     "",
			StartCmd:     "npx serve",
			DefaultPort:  "80",
		}
	}
	return nil
}

// ==========================================
// BACKEND DETECTORS
// ==========================================

// Detect Express
func detectExpress(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "express") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	startCmd := "node index.js" // Smart fallback
	rawStartScript := ""

	if customStart, exists := scripts["start"]; exists {
		startCmd = "npm start"
		rawStartScript = customStart
	}

	smartPort := extractPort(rawStartScript, "5000")

	return &DeploymentDetails{
		Type:         TypeExpress,
		AbsolutePath: dirPath,
		BuildCmd:     "npm install",
		StartCmd:     startCmd,
		DefaultPort:  smartPort,
	}
}

// Detect Go
func detectGo(dirPath string) *DeploymentDetails {
	if !hasFile(dirPath, "go.mod") {
		return nil
	}

	// 1. Smart Default for simple apps
	buildCmd := "go build -o main ."
	startCmd := "./main"

	// 2. The Enterprise Upgrade: Check if they use the "cmd" folder structure
	if hasDir(dirPath, "cmd") {
		buildCmd = "go build -o main ./cmd/..."
	}

	// 3. The Heroku Strategy: Check for a Procfile override
	if procCmd := getProcfileCommand(dirPath); procCmd != "" {
		startCmd = procCmd
	}

	return &DeploymentDetails{
		Type:         TypeGo,
		AbsolutePath: dirPath,
		BuildCmd:     buildCmd,
		StartCmd:     startCmd,
		DefaultPort:  "8080",
	}
}

// Detect Python
func detectPython(dirPath string) *DeploymentDetails {
	if !hasFile(dirPath, "requirements.txt") {
		return nil
	}

	startCmd := "python main.py" // Smart fallback

	// The Heroku Strategy: Check for a Procfile override
	if procCmd := getProcfileCommand(dirPath); procCmd != "" {
		startCmd = procCmd
	}

	return &DeploymentDetails{
		Type:         TypePython,
		AbsolutePath: dirPath,
		BuildCmd:     "pip install -r requirements.txt",
		StartCmd:     startCmd,
		DefaultPort:  "8000",
	}
}

// ==========================================
// UTILITY HELPERS
// ==========================================

// hasDependency safely opens package.json and checks for a specific package
func hasDependency(dirPath, depName string) bool {
	pkgPath := filepath.Join(dirPath, "package.json")
	file, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	err = json.Unmarshal(file, &pkg)
	if err != nil {
		// 🚨 DEBUG UPGRADE: This will tell us EXACTLY why it's failing!
		fmt.Printf("🚨 DEBUG: JSON Parsing failed in %s -> %v\n", pkgPath, err)
		return false
	}
	_, exists := pkg.Dependencies[depName]
	return exists
}

// hasFile checks if a file exists in a directory
func hasFile(dirPath, fileName string) bool {
	info, err := os.Stat(filepath.Join(dirPath, fileName))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// hasDir checks if a directory exists inside a given path
func hasDir(dirPath, dirName string) bool {
	info, err := os.Stat(filepath.Join(dirPath, dirName))
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir() // Returns true ONLY if it is a directory
}

// ==========================================
// DYNAMIC EXTRACTION HELPERS
// ==========================================

// getPackageScripts safely extracts the "scripts" object from package.json
func getPackageScripts(dirPath string) map[string]string {
	pkgPath := filepath.Join(dirPath, "package.json")
	file, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(file, &pkg); err != nil {
		return nil
	}

	return pkg.Scripts
}

// extractPort hunts the start script for custom port declarations
func extractPort(script string, defaultPort string) string {
	if script == "" {
		return defaultPort
	}

	// Scenario 1: Look for "PORT=8000 node server.js"
	if strings.Contains(script, "PORT=") {
		parts := strings.Split(script, "PORT=")
		fields := strings.Fields(parts[1])
		if len(fields) > 0 {
			return fields[0]
		}
	}

	// Scenario 2: Look for "next start -p 4000"
	if strings.Contains(script, "-p ") {
		parts := strings.Split(script, "-p ")
		fields := strings.Fields(parts[1])
		if len(fields) > 0 {
			return fields[0]
		}
	}

	return defaultPort
}

// getProcfileCommand mimics Heroku's Buildpack extraction
func getProcfileCommand(dirPath string) string {
	procfilePath := filepath.Join(dirPath, "Procfile")
	file, err := os.ReadFile(procfilePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "web:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "web:"))
		}
	}

	return ""
}
